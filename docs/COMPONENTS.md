# Componentes del Sistema - oli

Este documento detalla todos los componentes del sistema y sus relaciones.

---

## 1. Mapa de Componentes

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              OLI - ARQUITECTURA                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                                 CAPA CLI                                    │
│                            cmd/ollama-cli/                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                           main.go                                   │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │                                                                     │   │
│   │   RESPONSABILIDADES:                                                │   │
│   │   • Punto de entrada del programa                                   │   │
│   │   • Parsear argumentos de línea de comandos                         │   │
│   │   • Manejar modo directo vs modo interactivo                        │   │
│   │   • Procesar comandos especiales (help, ls, read, write, etc.)      │   │
│   │   • Manejar señales (Ctrl+C)                                        │   │
│   │                                                                     │   │
│   │   DEPENDENCIAS:                                                     │   │
│   │   • internal/cli.App                                                │   │
│   │   • internal/config                                                 │   │
│   │   • internal/tools                                                  │   │
│   │                                                                     │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ usa
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CAPA CORE                                      │
│                              internal/                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌───────────────────────────────────────────────────────────────────┐     │
│   │                         cli/app.go                                │     │
│   ├───────────────────────────────────────────────────────────────────┤     │
│   │                                                                   │     │
│   │   STRUCT App {                                                    │     │
│   │       model     string                                            │     │
│   │       client    llm.Client                                        │     │
│   │       providers []mcp.ContextProvider                             │     │
│   │       builder   *prompt.Builder                                   │     │
│   │   }                                                               │     │
│   │                                                                   │     │
│   │   MÉTODOS:                                                        │     │
│   │   • New() *App                                                    │     │
│   │   • NewWithPrompt(name) *App                                      │     │
│   │   • Run(ctx, task) error                                          │     │
│   │   • GetModel() string                                             │     │
│   │   • gatherContext() []ContextResult                               │     │
│   │   • offerToSaveCodeBlocks(response)                               │     │
│   │                                                                   │     │
│   │   RESPONSABILIDADES:                                              │     │
│   │   • Orquestar el flujo completo                                   │     │
│   │   • Coordinar providers, builder y client                         │     │
│   │   • Post-procesar respuestas (detectar código)                    │     │
│   │                                                                   │     │
│   └───────────────────────────────────────────────────────────────────┘     │
│                                      │                                      │
│            ┌─────────────────────────┼─────────────────────────┐            │
│            │                         │                         │            │
│            ▼                         ▼                         ▼            │
│   ┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐     │
│   │  config/        │      │    llm/         │      │    mcp/         │     │
│   │  config.go      │      │                 │      │                 │     │
│   ├─────────────────┤      ├─────────────────┤      ├─────────────────┤     │
│   │                 │      │ ┌─────────────┐ │      │ ┌─────────────┐ │     │
│   │ VARIABLES:      │      │ │ client.go   │ │      │ │ provider.go │ │     │
│   │ • Model         │      │ │ (interface) │ │      │ │ (interface) │ │     │
│   │ • OllamaURL     │      │ └─────────────┘ │      │ └─────────────┘ │     │
│   │ • MaxFiles      │      │        │        │      │        │        │     │
│   │ • MaxDepth      │      │        ▼        │      │        ▼        │     │
│   │ • SystemPrompt  │      │ ┌─────────────┐ │      │ ┌─────────────┐ │     │
│   │ • Prompts{}     │      │ │ ollama.go   │ │      │ │filesystem.go│ │     │
│   │                 │      │ │(implementa) │ │      │ │(implementa) │ │     │
│   │ RESPONSABILIDAD:│      │ └─────────────┘ │      │ └─────────────┘ │     │
│   │ Configuración   │      │                 │      │        │        │     │
│   │ centralizada    │      │ RESPONSABILIDAD:│      │        ▼        │     │
│   │                 │      │ Comunicación    │      │ ┌─────────────┐ │     │
│   └─────────────────┘      │ con Ollama API  │      │ │  git.go     │ │     │
│                            │                 │      │ │(implementa) │ │     │
│                            └─────────────────┘      │ └─────────────┘ │     │
│                                                     │                 │     │
│                                                     │ RESPONSABILIDAD:│     │
│                                                     │ Recopilar       │     │
│                                                     │ contexto        │     │
│                                                     └─────────────────┘     │
│                                                                             │
│   ┌─────────────────┐                              ┌─────────────────┐      │
│   │  prompt/        │                              │   tools/        │      │
│   │  builder.go     │                              │                 │      │
│   ├─────────────────┤                              ├─────────────────┤      │
│   │                 │                              │ ┌─────────────┐ │      │
│   │ STRUCT Builder  │                              │ │ tools.go    │ │      │
│   │                 │                              │ ├─────────────┤ │      │
│   │ MÉTODOS:        │                              │ │ • ReadFile  │ │      │
│   │ • NewBuilder()  │                              │ │ • WriteFile │ │      │
│   │ • Build()       │                              │ │ • ListDir   │ │      │
│   │                 │                              │ │ • AskConfirm│ │      │
│   │ RESPONSABILIDAD:│                              │ └─────────────┘ │      │
│   │ Construir       │                              │                 │      │
│   │ prompts         │                              │ RESPONSABILIDAD:│      │
│   │ estructurados   │                              │ Operaciones de  │      │
│   │                 │                              │ archivos        │      │
│   └─────────────────┘                              └─────────────────┘      │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ comunica con
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SERVICIOS EXTERNOS                                │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐           │
│   │    OLLAMA       │   │      GIT        │   │  FILESYSTEM     │           │
│   │    SERVER       │   │                 │   │                 │           │
│   ├─────────────────┤   ├─────────────────┤   ├─────────────────┤           │
│   │                 │   │                 │   │                 │           │
│   │ localhost:11434 │   │ exec.Command()  │   │ os.ReadFile()   │           │
│   │                 │   │                 │   │ os.WriteFile()  │           │
│   │ POST /api/      │   │ • branch        │   │ filepath.Walk() │           │
│   │   generate      │   │ • log           │   │                 │           │
│   │                 │   │ • status        │   │                 │           │
│   │ Modelo LLM:     │   │                 │   │                 │           │
│   │ qwen2.5-coder   │   │                 │   │                 │           │
│   │                 │   │                 │   │                 │           │
│   └─────────────────┘   └─────────────────┘   └─────────────────┘           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Interfaces del Sistema

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INTERFACES                                     │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                         llm.Client (interface)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   type Client interface {                                                   │
│       Generate(ctx context.Context,                                         │
│                req GenerateRequest,                                         │
│                onChunk func(string)) error                                  │
│   }                                                                         │
│                                                                             │
│   type GenerateRequest struct {                                             │
│       Model  string                                                         │
│       Prompt string                                                         │
│       System string                                                         │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│   IMPLEMENTACIONES:                                                         │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  OllamaClient                                                       │   │
│   │  • baseURL: string                                                  │   │
│   │  • httpClient: *http.Client                                         │   │
│   │  • Generate() → HTTP POST + NDJSON streaming                        │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   FUTURAS IMPLEMENTACIONES:                                                 │
│   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐             │
│   │  OpenAIClient   │  │  ClaudeClient   │  │  LocalLLMClient │             │
│   └─────────────────┘  └─────────────────┘  └─────────────────┘             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                      mcp.ContextProvider (interface)                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   type ContextProvider interface {                                          │
│       Name() string                                                         │
│       Gather(ctx context.Context, workDir string) (ContextResult, error)    │
│   }                                                                         │
│                                                                             │
│   type ContextResult struct {                                               │
│       Provider string                                                       │
│       Content  string                                                       │
│       Error    string                                                       │
│   }                                                                         │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│   IMPLEMENTACIONES:                                                         │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  FilesystemProvider                                                 │   │
│   │  • maxFiles: int                                                    │   │
│   │  • maxDepth: int                                                    │   │
│   │  • Gather() → Lee archivos del proyecto automáticamente             │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  GitProvider                                                        │   │
│   │  • Gather() → Ejecuta comandos git y parsea output                  │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   FUTURAS IMPLEMENTACIONES:                                                 │
│   ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐             │
│   │  DockerProvider │  │  EnvProvider    │  │  GitHubProvider │             │
│   └─────────────────┘  └─────────────────┘  └─────────────────┘             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Estructura de Archivos Detallada

```
ollama-cli/
│
├── cmd/
│   └── ollama-cli/
│       └── main.go                 # 150 líneas
│           │
│           ├── main()              # Punto de entrada
│           ├── runInteractive()    # Loop del modo interactivo
│           ├── handleCommand()     # Router de comandos
│           ├── readFileCmd()       # Comando read
│           ├── listDirCmd()        # Comando ls
│           ├── writeFileCmd()      # Comando write
│           └── showHelp()          # Comando help
│
├── internal/
│   │
│   ├── cli/
│   │   └── app.go                  # 130 líneas
│   │       │
│   │       ├── App struct          # Estructura principal
│   │       ├── New()               # Constructor
│   │       ├── Run()               # Ejecutar pregunta
│   │       ├── gatherContext()     # Recopilar contexto
│   │       └── offerToSaveCodeBlocks()  # Guardar código
│   │
│   ├── config/
│   │   └── config.go               # 60 líneas
│   │       │
│   │       ├── Model               # "qwen2.5-coder:14b"
│   │       ├── OllamaURL           # "http://localhost:11434"
│   │       ├── MaxFiles            # 30
│   │       ├── MaxDepth            # 4
│   │       ├── SystemPrompt        # Prompt principal
│   │       └── Prompts{}           # Mapa de prompts
│   │
│   ├── llm/
│   │   ├── client.go               # 20 líneas
│   │   │   │
│   │   │   ├── Client interface
│   │   │   └── GenerateRequest struct
│   │   │
│   │   └── ollama.go               # 80 líneas
│   │       │
│   │       ├── OllamaClient struct
│   │       ├── NewOllamaClient()
│   │       └── Generate()          # HTTP + streaming
│   │
│   ├── mcp/
│   │   ├── provider.go             # 15 líneas
│   │   │   │
│   │   │   ├── ContextProvider interface
│   │   │   └── ContextResult struct
│   │   │
│   │   ├── filesystem.go           # 200 líneas
│   │   │   │
│   │   │   ├── FilesystemProvider struct
│   │   │   ├── NewFilesystemProvider()
│   │   │   ├── Gather()            # Lee archivos
│   │   │   ├── readableExtensions  # Extensiones permitidas
│   │   │   ├── ignoredDirs         # Directorios ignorados
│   │   │   └── ignoredFiles        # Archivos ignorados
│   │   │
│   │   ├── git.go                  # 70 líneas
│   │   │   │
│   │   │   ├── GitProvider struct
│   │   │   └── Gather()            # Ejecuta git commands
│   │   │
│   │   └── registry.go             # 30 líneas (no usado aún)
│   │
│   ├── prompt/
│   │   └── builder.go              # 40 líneas
│   │       │
│   │       ├── Builder struct
│   │       ├── NewBuilder()
│   │       └── Build()             # Construye prompt
│   │
│   └── tools/
│       └── tools.go                # 80 líneas
│           │
│           ├── AskConfirmation()   # Preguntar s/n
│           ├── ReadFile()          # Leer archivo
│           ├── WriteFile()         # Escribir con confirmación
│           ├── WriteFileDirectly() # Escribir sin confirmación
│           └── ListDir()           # Listar directorio
│
├── docs/
│   ├── ARCHITECTURE.md             # Arquitectura general
│   ├── TECHNICAL.md                # Detalles técnicos
│   ├── ANALYSIS.md                 # Análisis FODA
│   ├── FLOW.md                     # Flujos de información
│   └── COMPONENTS.md               # Este documento
│
├── Makefile                        # build, install, clean
├── go.mod                          # Módulo Go
├── README.md                       # Documentación principal
└── PROGRESS.md                     # Estado del proyecto
```

---

## 4. Dependencias entre Componentes

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                     GRAFO DE DEPENDENCIAS                                   │
└─────────────────────────────────────────────────────────────────────────────┘

                              main.go
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
                    ▼            ▼            ▼
               cli/app.go    config.go    tools.go
                    │
       ┌────────────┼────────────┬────────────┐
       │            │            │            │
       ▼            ▼            ▼            ▼
   llm/client   mcp/provider  prompt/      tools/
       │            │         builder      tools.go
       ▼            │
   llm/ollama       ├─────────┬─────────┐
                    │         │         │
                    ▼         ▼         ▼
              filesystem    git      registry
                 .go        .go        .go


LEYENDA:
────────
→ depende de / importa

NOTAS:
──────
• config.go no tiene dependencias internas
• tools.go no tiene dependencias internas
• Las interfaces (client.go, provider.go) no tienen dependencias
• ollama.go depende de client.go
• filesystem.go y git.go dependen de provider.go
```

---

## 5. Ciclo de Vida de una Petición

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    CICLO DE VIDA: "oli explica main.go"                     │
└─────────────────────────────────────────────────────────────────────────────┘

TIEMPO ──────────────────────────────────────────────────────────────────────►

│ T0        │ T1          │ T2          │ T3          │ T4          │ T5
│           │             │             │             │             │
│ ENTRADA   │ CONTEXTO    │ PROMPT      │ LLM         │ RESPUESTA   │ POST
│           │             │             │             │             │
▼           ▼             ▼             ▼             ▼             ▼
┌─────┐   ┌─────┐       ┌─────┐       ┌─────┐       ┌─────┐       ┌─────┐
│Parse│──►│Gather│─────►│Build│─────►│Send │─────►│Stream│─────►│Save │
│Args │   │Context      │Prompt│      │to   │      │Response     │Code │
└─────┘   └─────┘       └─────┘       │Ollama       └─────┘       └─────┘
                                      └─────┘

DETALLES POR FASE:

T0 - ENTRADA (1ms)
├── Parsear os.Args
├── Detectar modo (directo/interactivo)
└── Crear App con configuración

T1 - CONTEXTO (50-200ms)
├── FilesystemProvider.Gather()
│   ├── WalkDir del proyecto
│   ├── Filtrar por extensión
│   ├── Leer contenido de archivos
│   └── Respetar límites (30 archivos, 200KB)
└── GitProvider.Gather()
    ├── git branch
    ├── git log -5
    └── git status

T2 - PROMPT (1ms)
├── Combinar SystemPrompt
├── Agregar contextos
└── Agregar tarea del usuario

T3 - LLM (100ms - 30s)
├── HTTP POST a Ollama
├── Esperar inicio de respuesta
└── Modelo procesa prompt

T4 - RESPUESTA (1s - 60s)
├── Recibir chunks NDJSON
├── Imprimir cada chunk
└── Acumular respuesta completa

T5 - POST-PROCESO (10ms)
├── Buscar bloques de código
├── Extraer nombre de archivo
├── Preguntar confirmación
└── Escribir archivo si confirma
```
