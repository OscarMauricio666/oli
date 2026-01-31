# Documento de Avance - oli

## Información del Proyecto

| Campo | Valor |
|-------|-------|
| Nombre | oli |
| Repositorio | https://github.com/OscarMauricio666/oli |
| Lenguaje | Go 1.22 |
| Fecha inicio | 2026-01-31 |

---

## Estructura de Ramas

```
main                        ← Rama principal (desarrollo activo)
│
├── v1-stable               ← Versión estable funcional (punto de rollback)
│   └── Commit: c5a72ac     "Initial commit"
│
└── feature/github-integration  ← Feature en desarrollo (experimental)
    └── Commit: 3952fda     "WIP: GitHub integration"
```

### Cómo hacer rollback

```bash
# Volver a la versión estable
git checkout v1-stable
make build && cp ./bin/oli ~/.local/bin/

# Volver a main
git checkout main
make build && cp ./bin/oli ~/.local/bin/
```

---

## Estado Actual

### ✅ Funcionalidades Completadas y Funcionando

| Feature | Estado | Descripción |
|---------|--------|-------------|
| Modo interactivo | ✅ Funciona | Conversación continua sin cerrar |
| Comando corto `oli` | ✅ Funciona | En lugar de `ollama-cli` |
| Sin comillas | ✅ Funciona | `oli que hace esto` funciona |
| Leer archivos | ✅ Funciona | `read archivo.go` muestra contenido |
| Listar directorio | ✅ Funciona | `ls` y `ls carpeta/` |
| Escribir archivos | ✅ Funciona | `write archivo.txt` con confirmación |
| Detección automática | ✅ Funciona | Pregunta si leer archivos mencionados |
| Guardar código sugerido | ✅ Funciona | Detecta bloques de código y ofrece guardar |
| Prompts personalizables | ✅ Funciona | En `internal/config/config.go` |
| Navegación | ✅ Funciona | `cd`, `pwd` funcionan |
| Streaming | ✅ Funciona | Respuestas en tiempo real |
| Contexto Git | ✅ Funciona | Muestra rama, commits, cambios |
| Contexto filesystem | ✅ Funciona | Lista archivos del proyecto |

### 🔄 En Desarrollo (rama feature/github-integration)

| Feature | Estado | Descripción |
|---------|--------|-------------|
| Listar repos | 🔄 En pruebas | `repos` - lista repositorios del usuario |
| Analizar repo | 🔄 En pruebas | `repo usuario/repo` - analiza repo remoto |
| Ver issues | 🔄 En pruebas | `issues usuario/repo` |
| Ver PRs | 🔄 En pruebas | `prs usuario/repo` |
| Clonar repo | 🔄 En pruebas | `clone usuario/repo` |

### ❌ Pendiente / Ideas Futuras

| Feature | Prioridad | Descripción |
|---------|-----------|-------------|
| Tests unitarios | Alta | Agregar tests para cada módulo |
| Leer archivo de repo remoto | Media | `repo user/repo read archivo.go` |
| Historial de conversación | Media | Memoria entre preguntas |
| Múltiples LLMs | Baja | Soporte para OpenAI, Claude API |
| Modo offline | Baja | Cache de respuestas |
| Plugins | Baja | Sistema de extensiones |

---

## Archivos del Proyecto

```
ollama-cli/
├── cmd/ollama-cli/
│   └── main.go              # Punto de entrada, comandos interactivos
├── internal/
│   ├── cli/
│   │   └── app.go           # Orquestador principal
│   ├── config/
│   │   └── config.go        # ⭐ CONFIGURACIÓN (modelo, prompts)
│   ├── llm/
│   │   ├── client.go        # Interface del cliente LLM
│   │   └── ollama.go        # Implementación Ollama
│   ├── mcp/
│   │   ├── provider.go      # Interface de proveedores
│   │   ├── filesystem.go    # Contexto del sistema de archivos
│   │   ├── git.go           # Contexto de Git
│   │   └── registry.go      # Registro de proveedores
│   ├── prompt/
│   │   └── builder.go       # Constructor de prompts
│   └── tools/
│       ├── tools.go         # Herramientas (read, write, list)
│       └── github.go        # 🔄 Herramientas GitHub (en desarrollo)
├── docs/
│   ├── ARCHITECTURE.md      # Documentación de arquitectura
│   ├── TECHNICAL.md         # Documentación técnica
│   └── ANALYSIS.md          # Análisis del proyecto
├── Makefile                 # Comandos de compilación
├── go.mod                   # Definición del módulo
├── README.md                # Documentación principal
└── PROGRESS.md              # ⭐ Este documento
```

---

## Configuración Actual

**Archivo: `internal/config/config.go`**

```go
var Model = "qwen2.5-coder:14b"           // Modelo de Ollama
var OllamaURL = "http://localhost:11434"   // URL del servidor
var MaxFiles = 50                          // Máx archivos en contexto
var MaxDepth = 3                           // Profundidad de carpetas
```

---

## Comandos Disponibles (v1-stable)

### Modo Interactivo
```bash
oli                    # Iniciar modo interactivo
```

### Dentro del modo interactivo
```
help                   # Ayuda
ls [dir]               # Listar archivos
read <archivo>         # Leer archivo
write <archivo>        # Escribir archivo (con confirmación)
pwd                    # Directorio actual
cd <dir>               # Cambiar directorio
prompts                # Ver prompts disponibles
salir                  # Salir
```

### Modo Directo
```bash
oli que hace este proyecto
oli explica el archivo main.go
oli crea un archivo hello.py con hola mundo
```

---

## Cómo Compilar e Instalar

```bash
# Compilar
make build

# Instalar en ~/.local/bin/
cp ./bin/oli ~/.local/bin/

# O instalar globalmente (requiere sudo)
sudo make install
```

---

## Próximos Pasos

1. **Probar integración GitHub** - Verificar que `repos`, `repo`, `issues`, `prs` funcionen
2. **Agregar tests** - Crear tests unitarios para módulos críticos
3. **Documentar API** - Completar documentación de funciones internas
4. **Optimizar prompts** - Mejorar respuestas del modelo

---

## Notas de Desarrollo

- El proyecto usa solo la librería estándar de Go (sin dependencias externas)
- Los prompts se pueden personalizar en `internal/config/config.go`
- Para cambiar el modelo, modificar `config.Model` y recompilar
- La rama `v1-stable` es el punto seguro de rollback

---

*Última actualización: 2026-01-31*
