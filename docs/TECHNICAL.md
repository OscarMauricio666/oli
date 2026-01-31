# Documentación Técnica

Este documento proporciona detalles técnicos sobre la implementación, APIs internas y guías para desarrolladores.

## Tabla de Contenidos

1. [APIs Internas](#apis-internas)
2. [Implementación de Módulos](#implementación-de-módulos)
3. [Manejo de Errores](#manejo-de-errores)
4. [Guía de Extensión](#guía-de-extensión)
5. [Testing](#testing)
6. [Rendimiento](#rendimiento)

---

## APIs Internas

### Módulo LLM (`internal/llm`)

#### Interface Client

```go
type Client interface {
    Generate(ctx context.Context, req GenerateRequest, onChunk func(string)) error
}
```

| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| `ctx` | `context.Context` | Contexto para cancelación y timeouts |
| `req` | `GenerateRequest` | Configuración de la solicitud |
| `onChunk` | `func(string)` | Callback para cada fragmento de respuesta |

#### GenerateRequest

```go
type GenerateRequest struct {
    Model  string  // Nombre del modelo (ej: "llama3.2")
    Prompt string  // Prompt del usuario
    System string  // Prompt del sistema
}
```

#### OllamaClient

```go
type OllamaClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient
func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest, onChunk func(string)) error
```

**Protocolo de comunicación:**

```
POST /api/generate HTTP/1.1
Content-Type: application/json

{
    "model": "llama3.2",
    "prompt": "...",
    "system": "...",
    "stream": true
}
```

**Formato de respuesta (NDJSON):**

```json
{"response": "Hola", "done": false}
{"response": " mundo", "done": false}
{"response": "", "done": true}
```

---

### Módulo MCP (`internal/mcp`)

#### Interface ContextProvider

```go
type ContextProvider interface {
    Name() string
    Gather(ctx context.Context, workDir string) (ContextResult, error)
}
```

| Método | Retorno | Descripción |
|--------|---------|-------------|
| `Name()` | `string` | Identificador único del proveedor |
| `Gather()` | `ContextResult, error` | Recopila contexto del directorio |

#### ContextResult

```go
type ContextResult struct {
    Name    string  // Nombre del proveedor
    Content string  // Contenido recopilado
    Error   error   // Error si falló la recopilación
}
```

#### FilesystemProvider

**Configuración interna:**

| Constante | Valor | Descripción |
|-----------|-------|-------------|
| `maxFiles` | 50 | Máximo número de archivos a listar |
| `maxDepth` | 3 | Profundidad máxima de directorio |

**Directorios ignorados:**

```go
var ignoredDirs = map[string]bool{
    "node_modules": true,
    "vendor":       true,
    "__pycache__":  true,
    "dist":         true,
    "build":        true,
    ".git":         true,
    "bin":          true,
    "obj":          true,
    "target":       true,
    ".idea":        true,
    ".vscode":      true,
    "coverage":     true,
    ".next":        true,
    ".nuxt":        true,
    "venv":         true,
}
```

**Algoritmo de recopilación:**

```
1. Iniciar WalkDir desde workDir
2. Para cada entrada:
   a. Calcular profundidad relativa
   b. Si profundidad > maxDepth → SkipDir
   c. Si es directorio ignorado → SkipDir
   d. Si empieza con '.' → Continuar
   e. Si es archivo y total < maxFiles → Agregar
   f. Si total >= maxFiles → Detener
3. Retornar lista formateada
```

#### GitProvider

**Comandos ejecutados:**

| Comando | Propósito |
|---------|-----------|
| `git rev-parse --git-dir` | Verificar si es repositorio Git |
| `git branch --show-current` | Obtener rama actual |
| `git log --oneline -5` | Últimos 5 commits |
| `git status --short` | Estado de archivos modificados |

**Formato de salida:**

```
Current branch: main

Recent commits:
abc1234 Fix bug in parser
def5678 Add new feature
...

Changed files:
M  file1.go
A  file2.go
?? file3.txt
```

---

### Módulo Prompt (`internal/prompt`)

#### Builder

```go
type Builder struct{}

func (b *Builder) Build(contexts []mcp.ContextResult, task string) (system, user string)
```

**Estructura del prompt del sistema:**

```
You are a helpful coding assistant that analyzes codebases.
You have access to the following context about the current project.

IMPORTANT: You are in READ-ONLY mode. You can analyze and suggest changes,
but you cannot directly modify files. Always propose changes clearly
so the user can implement them.

Be concise and specific in your responses.
```

**Estructura del prompt del usuario:**

```
## Context: filesystem
[lista de archivos]

## Context: git
[información de git]

## Task
[pregunta del usuario]
```

---

### Módulo CLI (`internal/cli`)

#### App

```go
type App struct {
    config    Config
    client    llm.Client
    providers []mcp.ContextProvider
    builder   *prompt.Builder
}

func NewApp(config Config) *App
func (a *App) Run(ctx context.Context, task string) error
```

#### Config

```go
type Config struct {
    Model     string  // Modelo a usar
    OllamaURL string  // URL del servicio Ollama
}
```

**Flujo de Run():**

```
┌────────────────┐
│      Run       │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│   os.Getwd     │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ gatherContext  │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│ builder.Build  │
└───────┬────────┘
        │
        ▼
┌────────────────┐
│client.Generate │
└───────┬────────┘
        │
        ▼
   ┌─────────┐
   │ Error?  │
   └────┬────┘
        │
   ┌────┴────┐
   │         │
   ▼         ▼
┌──────┐  ┌──────┐
│  Sí  │  │  No  │
└──┬───┘  └──┬───┘
   │         │
   ▼         ▼
┌────────┐ ┌────────┐
│ return │ │ return │
│ error  │ │  nil   │
└────────┘ └────────┘
```

---

## Implementación de Módulos

### Streaming NDJSON

El cliente Ollama implementa streaming mediante NDJSON (Newline Delimited JSON):

```go
func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest, onChunk func(string)) error {
    // 1. Preparar request
    ollamaReq := ollamaRequest{
        Model:  req.Model,
        Prompt: req.Prompt,
        System: req.System,
        Stream: true,
    }

    body, _ := json.Marshal(ollamaReq)

    // 2. Crear HTTP request con contexto
    httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")

    // 3. Ejecutar request
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // 4. Leer stream línea por línea
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        var chunk ollamaResponse
        if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
            continue // Ignorar líneas malformadas
        }

        if chunk.Error != "" {
            return fmt.Errorf("ollama error: %s", chunk.Error)
        }

        if chunk.Response != "" {
            onChunk(chunk.Response)
        }

        if chunk.Done {
            break
        }
    }

    return scanner.Err()
}
```

### Context Cancellation

Todos los módulos respetan la cancelación de contexto:

```go
// En FilesystemProvider
func (p *FilesystemProvider) Gather(ctx context.Context, workDir string) (ContextResult, error) {
    var files []string

    err := filepath.WalkDir(workDir, func(path string, d fs.DirEntry, err error) error {
        // Verificar cancelación en cada iteración
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        // ... resto de la lógica
    })

    // ...
}
```

---

## Manejo de Errores

### Estrategia de Degradación Graceful

```
                    ┌─────────────────────┐
                    │  Error en Provider  │
                    └──────────┬──────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │   Tipo de Error     │
                    └──────────┬──────────┘
                               │
         ┌─────────────────────┼─────────────────────┐
         │                     │                     │
         ▼                     ▼                     ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ No es repo Git  │  │Archivo inaccesib│  │Contexto cancelad│
└────────┬────────┘  └────────┬────────┘  └────────┬────────┘
         │                    │                    │
         ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│Retornar mensaje │  │ Saltar archivo  │  │ Propagar error  │
│  informativo    │  │   y continuar   │  │   al caller     │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

### Propagación de Errores

| Capa | Comportamiento |
|------|---------------|
| Provider | Retorna error en ContextResult.Error |
| Builder | Incluye error como texto en prompt |
| Client | Propaga error al App |
| App | Propaga error a main |
| main | Exit 1 + mensaje stderr |

### Errores Comunes

| Error | Causa | Solución |
|-------|-------|----------|
| `connection refused` | Ollama no está corriendo | Iniciar Ollama |
| `model not found` | Modelo no descargado | `ollama pull <modelo>` |
| `context deadline exceeded` | Timeout de request | Verificar conectividad |

---

## Guía de Extensión

### Agregar un Nuevo Provider de Contexto

1. **Crear archivo en `internal/mcp/`:**

```go
// docker.go
package mcp

import (
    "context"
    "os/exec"
)

type DockerProvider struct{}

func (p *DockerProvider) Name() string {
    return "docker"
}

func (p *DockerProvider) Gather(ctx context.Context, workDir string) (ContextResult, error) {
    result := ContextResult{Name: p.Name()}

    // Verificar si existe docker-compose.yml
    // Ejecutar docker ps
    // ...

    return result, nil
}
```

2. **Registrar en App:**

```go
// En cli/app.go
func NewApp(config Config) *App {
    return &App{
        config:  config,
        client:  llm.NewOllamaClient(config.OllamaURL),
        providers: []mcp.ContextProvider{
            &mcp.FilesystemProvider{},
            &mcp.GitProvider{},
            &mcp.DockerProvider{}, // Nuevo
        },
        builder: &prompt.Builder{},
    }
}
```

### Agregar un Nuevo Backend LLM

1. **Implementar interface Client:**

```go
// anthropic.go
package llm

import (
    "context"
    "net/http"
)

type AnthropicClient struct {
    apiKey     string
    httpClient *http.Client
}

func NewAnthropicClient(apiKey string) *AnthropicClient {
    return &AnthropicClient{
        apiKey:     apiKey,
        httpClient: &http.Client{},
    }
}

func (c *AnthropicClient) Generate(ctx context.Context, req GenerateRequest, onChunk func(string)) error {
    // Implementar llamada a API de Anthropic
    // ...
    return nil
}
```

2. **Configurar en main.go:**

```go
var client llm.Client
if os.Getenv("USE_ANTHROPIC") != "" {
    client = llm.NewAnthropicClient(os.Getenv("ANTHROPIC_API_KEY"))
} else {
    client = llm.NewOllamaClient(ollamaURL)
}
```

---

## Testing

### Estructura de Tests

```
internal/
├── cli/
│   └── app_test.go
├── llm/
│   ├── client_test.go
│   └── ollama_test.go
├── mcp/
│   ├── filesystem_test.go
│   └── git_test.go
└── prompt/
    └── builder_test.go
```

### Mocking Interfaces

```go
// Mock para Client
type MockClient struct {
    GenerateFunc func(ctx context.Context, req GenerateRequest, onChunk func(string)) error
}

func (m *MockClient) Generate(ctx context.Context, req GenerateRequest, onChunk func(string)) error {
    return m.GenerateFunc(ctx, req, onChunk)
}

// Mock para ContextProvider
type MockProvider struct {
    name   string
    result ContextResult
}

func (m *MockProvider) Name() string { return m.name }
func (m *MockProvider) Gather(ctx context.Context, workDir string) (ContextResult, error) {
    return m.result, nil
}
```

### Ejemplo de Test

```go
func TestApp_Run(t *testing.T) {
    mockClient := &MockClient{
        GenerateFunc: func(ctx context.Context, req GenerateRequest, onChunk func(string)) error {
            onChunk("Test response")
            return nil
        },
    }

    app := &App{
        config:    Config{Model: "test"},
        client:    mockClient,
        providers: []mcp.ContextProvider{},
        builder:   &prompt.Builder{},
    }

    err := app.Run(context.Background(), "test task")
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### Ejecutar Tests

```bash
# Todos los tests
go test ./...

# Con cobertura
go test -cover ./...

# Test específico
go test ./internal/llm/... -v

# Con race detector
go test -race ./...
```

---

## Rendimiento

### Limitaciones Configuradas

| Recurso | Límite | Razón |
|---------|--------|-------|
| Archivos listados | 50 | Evitar prompts excesivamente largos |
| Profundidad directorio | 3 | Balance entre contexto y ruido |
| Commits mostrados | 5 | Información reciente relevante |

### Optimizaciones Implementadas

1. **Early termination**: El WalkDir se detiene al alcanzar 50 archivos
2. **Directory skipping**: Directorios ignorados no se recorren
3. **Streaming**: Las respuestas se muestran mientras llegan
4. **Context cancellation**: Operaciones cancelables para respuesta rápida a Ctrl+C

### Métricas de Uso de Memoria

| Operación | Memoria Estimada |
|-----------|------------------|
| Lista de archivos | O(n) donde n = num archivos |
| Buffer de respuesta | Streaming, no acumula |
| HTTP client | Pool de conexiones reutilizable |

### Recomendaciones de Uso

- Para proyectos grandes, ejecutar desde subdirectorios específicos
- Usar modelos más pequeños para respuestas más rápidas
- Considerar aumentar `maxFiles` para proyectos con estructura plana

---

## Variables de Entorno

| Variable | Tipo | Default | Descripción |
|----------|------|---------|-------------|
| `OLLAMA_MODEL` | string | `llama3.2` | Modelo LLM a utilizar |
| `OLLAMA_URL` | string | `http://localhost:11434` | URL del servicio Ollama |

## Códigos de Salida

| Código | Significado |
|--------|-------------|
| 0 | Ejecución exitosa |
| 1 | Error general |
| 2 | Argumentos inválidos (sin tarea) |

## Dependencias del Sistema

| Dependencia | Requerida | Uso |
|-------------|-----------|-----|
| Ollama | Sí | Backend LLM |
| Git | No | Contexto de repositorio |
| Go 1.22+ | Solo compilación | Compilar el proyecto |
