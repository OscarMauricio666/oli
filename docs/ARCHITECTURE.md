# Arquitectura del Sistema

Este documento describe la arquitectura de Ollama CLI, incluyendo diagramas de componentes, flujos de datos y decisiones de diseño.

## Visión General

Ollama CLI sigue una arquitectura modular con separación clara de responsabilidades. El sistema se compone de cuatro módulos principales que interactúan para proporcionar análisis de código asistido por IA.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CAPA DE PRESENTACIÓN                                │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                     main.go (Punto de Entrada)                        │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
└──────────────────────────────────────┼──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CAPA DE ORQUESTACIÓN                                │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                     cli/app.go (Orquestador)                          │  │
│  └───────────────────────────────────┬───────────────────────────────────┘  │
└──────────────────────────────────────┼──────────────────────────────────────┘
                                       │
                     ┌─────────────────┼─────────────────┐
                     ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          CAPA DE SERVICIOS                                  │
│  ┌─────────────────────┐  ┌─────────────────────┐  ┌─────────────────────┐  │
│  │ prompt/builder.go   │  │   llm/client.go     │  │   llm/ollama.go     │  │
│  │ Constructor Prompts │  │   Interfaz LLM      │  │   Implementación    │  │
│  └─────────────────────┘  └─────────────────────┘  └──────────┬──────────┘  │
└───────────────────────────────────────────────────────────────┼─────────────┘
                                                                │
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CAPA DE DATOS                                     │
│  ┌─────────────────────────────┐  ┌─────────────────────────────┐           │
│  │   mcp/filesystem.go        │  │      mcp/git.go             │           │
│  │   Contexto del FS          │  │      Contexto de Git        │           │
│  └──────────────┬──────────────┘  └──────────────┬──────────────┘           │
└─────────────────┼────────────────────────────────┼──────────────────────────┘
                  │                                │
                  ▼                                ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        SERVICIOS EXTERNOS                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐              │
│  │   Ollama API    │  │ Sistema Archivos│  │       Git       │              │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Diagrama de Componentes

### Componentes Principales

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INTERFACES                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────┐      ┌─────────────────────────────┐       │
│  │     <<interface>>           │      │      <<interface>>          │       │
│  │        Client               │      │     ContextProvider         │       │
│  ├─────────────────────────────┤      ├─────────────────────────────┤       │
│  │ + Generate(ctx, req, chunk) │      │ + Name() string             │       │
│  │   error                     │      │ + Gather(ctx, workDir)      │       │
│  └──────────────┬──────────────┘      │   (ContextResult, error)    │       │
│                 │                     └──────────────┬──────────────┘       │
│                 │                                    │                      │
└─────────────────┼────────────────────────────────────┼──────────────────────┘
                  │                                    │
                  ▼                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           IMPLEMENTACIONES                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────┐      ┌─────────────────────────────┐       │
│  │      OllamaClient           │      │    FilesystemProvider       │       │
│  ├─────────────────────────────┤      ├─────────────────────────────┤       │
│  │ - baseURL: string           │      │ + Name() string             │       │
│  │ - httpClient: *http.Client  │      │ + Gather(ctx, workDir)      │       │
│  ├─────────────────────────────┤      └─────────────────────────────┘       │
│  │ + Generate(ctx, req, chunk) │                                            │
│  └─────────────────────────────┘      ┌─────────────────────────────┐       │
│                                       │       GitProvider           │       │
│                                       ├─────────────────────────────┤       │
│                                       │ + Name() string             │       │
│                                       │ + Gather(ctx, workDir)      │       │
│                                       └─────────────────────────────┘       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                            ORQUESTADOR                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │                              App                                    │    │
│  ├─────────────────────────────────────────────────────────────────────┤    │
│  │ - config: Config                                                    │    │
│  │ - client: Client                                                    │    │
│  │ - providers: []ContextProvider                                      │    │
│  │ - builder: *Builder                                                 │    │
│  ├─────────────────────────────────────────────────────────────────────┤    │
│  │ + Run(ctx context.Context, task string) error                       │    │
│  │ - gatherContext(ctx, workDir) []ContextResult                       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Flujo de Ejecución

### Secuencia Principal

```
  Usuario        main.go         App       FS Provider   Git Provider    Builder      OllamaClient    Ollama API
     │              │              │             │             │            │              │              │
     │──"pregunta"─►│              │             │             │            │              │              │
     │              │──Run(task)──►│             │             │            │              │              │
     │              │              │             │             │            │              │              │
     │              │              │──Gather()──►│             │            │              │              │
     │              │              │◄──archivos──│             │            │              │              │
     │              │              │             │             │            │              │              │
     │              │              │──Gather()───────────────►│            │              │              │
     │              │              │◄──git info────────────────│            │              │              │
     │              │              │             │             │            │              │              │
     │              │              │──Build(contexts, task)───────────────►│              │              │
     │              │              │◄──systemPrompt, userPrompt─────────────│              │              │
     │              │              │             │             │            │              │              │
     │              │              │──Generate(request, callback)─────────────────────────►│              │
     │              │              │             │             │            │              │──POST /api/──►│
     │              │              │             │             │            │              │    generate   │
     │              │              │             │             │            │              │◄──chunk 1─────│
     │◄─print(chunk 1)─────────────│◄────────────callback(chunk 1)─────────────────────────│              │
     │              │              │             │             │            │              │◄──chunk 2─────│
     │◄─print(chunk 2)─────────────│◄────────────callback(chunk 2)─────────────────────────│              │
     │              │              │             │             │            │              │◄──done────────│
     │              │◄──nil────────│             │             │            │              │              │
     │◄──exit 0─────│              │             │             │            │              │              │
     │              │              │             │             │            │              │              │
```

### Flujo de Recopilación de Contexto

```
                              ┌─────────────────┐
                              │     INICIO      │
                              └────────┬────────┘
                                       │
                                       ▼
                          ┌────────────────────────┐
                          │ Obtener directorio de  │
                          │       trabajo          │
                          └────────────┬───────────┘
                                       │
                     ┌─────────────────┴─────────────────┐
                     ▼                                   ▼
        ┌────────────────────────┐          ┌────────────────────────┐
        │  FilesystemProvider    │          │     GitProvider        │
        │      .Gather()         │          │       .Gather()        │
        └────────────┬───────────┘          └────────────┬───────────┘
                     │                                   │
                     ▼                                   ▼
        ┌────────────────────────┐          ┌────────────────────────┐
        │   WalkDir desde        │          │  ¿Es repositorio Git?  │
        │     workDir            │          └────────────┬───────────┘
        └────────────┬───────────┘                       │
                     │                          ┌────────┴────────┐
                     ▼                          │                 │
        ┌────────────────────────┐              ▼                 ▼
        │  ¿Profundidad > 3?     │         ┌────────┐        ┌────────┐
        └────────────┬───────────┘         │   NO   │        │   SÍ   │
                     │                     └────┬───┘        └────┬───┘
            ┌────────┴────────┐                 │                 │
            ▼                 ▼                 ▼                 ▼
       ┌────────┐        ┌────────┐    ┌──────────────┐  ┌──────────────┐
       │   SÍ   │        │   NO   │    │  Retornar    │  │ git branch   │
       └────┬───┘        └────┬───┘    │  "No es      │  │ git log      │
            │                 │        │  repo git"   │  │ git status   │
            ▼                 ▼        └──────────────┘  └──────┬───────┘
       ┌─────────┐   ┌────────────────┐                        │
       │ SkipDir │   │¿Dir ignorado?  │                        ▼
       └─────────┘   └────────┬───────┘                ┌──────────────┐
                              │                        │  Formatear   │
                     ┌────────┴────────┐               │  resultado   │
                     ▼                 ▼               └──────────────┘
                ┌────────┐        ┌────────┐
                │   SÍ   │        │   NO   │
                └────┬───┘        └────┬───┘
                     │                 │
                     ▼                 ▼
                ┌─────────┐   ┌────────────────┐
                │ SkipDir │   │¿Total < 50?    │
                └─────────┘   └────────┬───────┘
                                       │
                              ┌────────┴────────┐
                              ▼                 ▼
                         ┌────────┐        ┌────────┐
                         │   SÍ   │        │   NO   │
                         └────┬───┘        └────┬───┘
                              │                 │
                              ▼                 ▼
                      ┌──────────────┐  ┌──────────────┐
                      │ Agregar a    │  │  Terminar    │
                      │    lista     │  │    walk      │
                      └──────────────┘  └──────────────┘
```

## Flujo de Comunicación con Ollama

```
┌──────────────────┐                    ┌──────────────────┐                    ┌──────────────────┐
│   OllamaClient   │                    │    HTTP Client   │                    │   Ollama API     │
└────────┬─────────┘                    └────────┬─────────┘                    └────────┬─────────┘
         │                                       │                                       │
         │  1. Crear GenerateRequest             │                                       │
         │─────────────────────────►             │                                       │
         │                                       │                                       │
         │  2. Marshal a JSON                    │                                       │
         │─────────────────────────►             │                                       │
         │                                       │                                       │
         │  3. NewRequestWithContext(POST)       │                                       │
         │──────────────────────────────────────►│                                       │
         │                                       │                                       │
         │                                       │  4. POST /api/generate                │
         │                                       │──────────────────────────────────────►│
         │                                       │                                       │
         │                                       │         [Procesamiento del modelo]    │
         │                                       │                                       │
         │                                       │  5. NDJSON chunk 1                    │
         │                                       │◄──────────────────────────────────────│
         │  6. response body                     │                                       │
         │◄──────────────────────────────────────│                                       │
         │                                       │                                       │
         │  7. Unmarshal JSON                    │                                       │
         │─────────────────────────►             │                                       │
         │                                       │                                       │
         │  8. callback(response.Response)       │                                       │
         │─────────────────────────►             │                                       │
         │                                       │                                       │
         │                    [Repetir para cada chunk...]                               │
         │                                       │                                       │
         │                                       │  9. {"done": true}                    │
         │                                       │◄──────────────────────────────────────│
         │  10. EOF                              │                                       │
         │◄──────────────────────────────────────│                                       │
         │                                       │                                       │
         │  11. return nil                       │                                       │
         │─────────────────────────►             │                                       │
         │                                       │                                       │
```

## Diagrama de Estados

### Estados del Cliente LLM

```
                                    ┌─────────────────────┐
                                    │                     │
                      Crear cliente │       IDLE          │◄──────────────────────────┐
              ┌────────────────────►│                     │                           │
              │                     └──────────┬──────────┘                           │
              │                                │                                      │
              │                    Generate()  │                                      │
              │                     llamado    │                                      │
              │                                ▼                                      │
              │                     ┌─────────────────────┐                           │
              │                     │                     │         Marshal Error     │
              │                     │     PREPARING       │──────────────────────────►│
              │                     │                     │                           │
              │                     └──────────┬──────────┘                           │
              │                                │                                      │
              │                Request         │                                      │
              │                preparado       │                                      │
              │                                ▼                                      │
              │                     ┌─────────────────────┐                           │
              │                     │                     │         HTTP Error        │
              │                     │     REQUESTING      │──────────────────┐        │
              │                     │                     │                  │        │
              │                     └──────────┬──────────┘                  │        │
              │                                │                             │        │
              │                 Response       │                             │        │
              │                 recibido       │                             │        │
              │                                ▼                             │        │
              │                     ┌─────────────────────┐                  │        │
              │      ┌─────────────►│                     │                  │        │
              │      │              │     STREAMING       │                  ▼        │
              │      │  Chunk       │                     │──────────► ┌───────────┐  │
              │      │  recibido    └──────────┬──────────┘  Error en  │           │  │
              │      │                         │             chunk     │   ERROR   │──┘
              │      └─────────────────────────┤                       │           │
              │                                │                       └───────────┘
              │                     Done = true│                            │
              │                                ▼                            │
              │                     ┌─────────────────────┐                 │
              │                     │                     │                 │
              │    Retornar nil     │     COMPLETED       │                 │
              └─────────────────────│                     │◄────────────────┘
                                    └─────────────────────┘   Retornar error
```

## Estructura de Datos

### Modelo de Dominio

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           ESTRUCTURAS DE DATOS                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────┐       ┌─────────────────────┐                      │
│  │       CONFIG        │       │   GENERATE_REQUEST  │                      │
│  ├─────────────────────┤       ├─────────────────────┤                      │
│  │ Model: string       │──────►│ Model: string       │                      │
│  │ OllamaURL: string   │       │ Prompt: string      │                      │
│  └─────────────────────┘       │ System: string      │                      │
│                                └──────────┬──────────┘                      │
│                                           │                                 │
│                                           │ se transforma en                │
│                                           ▼                                 │
│  ┌─────────────────────┐       ┌─────────────────────┐                      │
│  │   CONTEXT_RESULT    │       │   OLLAMA_REQUEST    │                      │
│  ├─────────────────────┤       ├─────────────────────┤                      │
│  │ Name: string        │       │ Model: string       │                      │
│  │ Content: string     │       │ Prompt: string      │                      │
│  │ Error: error        │       │ System: string      │                      │
│  └─────────────────────┘       │ Stream: bool        │                      │
│                                └──────────┬──────────┘                      │
│                                           │                                 │
│                                           │ genera                          │
│                                           ▼                                 │
│                                ┌─────────────────────┐                      │
│                                │   OLLAMA_RESPONSE   │                      │
│                                ├─────────────────────┤                      │
│                                │ Response: string    │                      │
│                                │ Done: bool          │                      │
│                                │ Error: string       │                      │
│                                └─────────────────────┘                      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Patrones de Diseño Utilizados

### 1. Strategy Pattern (Proveedores de Contexto)

```
                    ┌─────────────────────────────────┐
                    │              App                │
                    └───────────────┬─────────────────┘
                                    │
                                    │ usa
                                    ▼
                    ┌─────────────────────────────────┐
                    │    <<interface>>                │
                    │    ContextProvider              │
                    └───────────────┬─────────────────┘
                                    │
              ┌─────────────────────┼─────────────────────┐
              │                     │                     │
              ▼                     ▼                     ▼
┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────┐
│ FilesystemProvider  │ │    GitProvider      │ │  FutureProvider...  │
└─────────────────────┘ └─────────────────────┘ └─────────────────────┘
```

### 2. Factory Pattern (Cliente LLM)

```
                    ┌─────────────────────────────────┐
                    │              App                │
                    └───────────────┬─────────────────┘
                                    │
                                    │ usa
                                    ▼
                    ┌─────────────────────────────────┐
                    │    <<interface>>                │
                    │         Client                  │
                    └───────────────┬─────────────────┘
                                    │
              ┌─────────────────────┴─────────────────────┐
              │                                           │
              ▼                                           ▼
┌─────────────────────────────┐           ┌─────────────────────────────┐
│       OllamaClient          │           │      FutureClient...        │
└─────────────────────────────┘           └─────────────────────────────┘
```

### 3. Builder Pattern (Construcción de Prompts)

```
┌───────────────────────────────────────────────────────────────────────┐
│                            Builder                                    │
└───────────────────────────────┬───────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        │                       │                       │
        ▼                       ▼                       ▼
┌───────────────┐       ┌───────────────┐       ┌───────────────┐
│  Sistema      │       │   Contexto    │       │   Tarea del   │
│  Prompt Base  │       │  (FS + Git)   │       │    Usuario    │
└───────┬───────┘       └───────┬───────┘       └───────┬───────┘
        │                       │                       │
        └───────────────────────┼───────────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │     Prompt Final      │
                    └───────────────────────┘
```

## Puntos de Extensibilidad

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         PUNTOS EXTENSIBLES                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  EXTENSIÓN                      INTERFAZ                 EJEMPLO FUTURO     │
│  ──────────────────────────────────────────────────────────────────────────│
│                                                                             │
│  Nuevos Proveedores    ──────►  ContextProvider   ──────►  DockerProvider   │
│  de Contexto                                               NPMProvider      │
│                                                            EnvProvider      │
│                                                                             │
│  Nuevos Backends       ──────►  Client            ──────►  ClaudeClient     │
│  LLM                                                       OpenAIClient     │
│                                                            LocalLlamaClient │
│                                                                             │
│  Nuevos Formatos       ──────►  Callback func     ──────►  JSONFormatter    │
│  de Salida                                                 MarkdownFormatter│
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Decisiones de Arquitectura

### ADR-001: Uso Exclusivo de Librería Estándar

| Aspecto | Detalle |
|---------|---------|
| **Contexto** | Necesidad de una herramienta ligera y portable |
| **Decisión** | Usar solo la librería estándar de Go, sin dependencias externas |
| **Consecuencias (+)** | Sin gestión de dependencias, binario pequeño, compilación rápida |
| **Consecuencias (-)** | Más código para funcionalidades comunes |

### ADR-002: Streaming de Respuestas

| Aspecto | Detalle |
|---------|---------|
| **Contexto** | Mejorar la experiencia de usuario durante respuestas largas |
| **Decisión** | Implementar streaming NDJSON con callbacks |
| **Consecuencias (+)** | Feedback inmediato al usuario, menor uso de memoria |
| **Consecuencias (-)** | Mayor complejidad en manejo de errores |

### ADR-003: Proveedores de Contexto Modulares

| Aspecto | Detalle |
|---------|---------|
| **Contexto** | Necesidad de extender las fuentes de información |
| **Decisión** | Usar interface `ContextProvider` con registro dinámico |
| **Consecuencias (+)** | Fácil agregar nuevas fuentes, testeable, separación de responsabilidades |
| **Consecuencias (-)** | Overhead de abstracción |

## Métricas de Calidad

| Métrica | Valor | Objetivo |
|---------|-------|----------|
| Dependencias externas | 0 | 0 |
| Cobertura de tests | TBD | >80% |
| Complejidad ciclomática (avg) | Baja | <10 |
| Líneas de código | ~500 | <1000 |
| Tiempo de compilación | <2s | <5s |
