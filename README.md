# Ollama CLI

Una herramienta de línea de comandos que utiliza modelos LLM locales (via Ollama) para analizar y comprender codebases. Actúa como un asistente de programación que puede explicar código, proporcionar sugerencias y responder preguntas sobre proyectos.

## Características

- **Análisis de código contextual**: Recopila automáticamente información del sistema de archivos y Git
- **Streaming de respuestas**: Respuestas en tiempo real desde el modelo
- **Sin dependencias externas**: Solo utiliza la librería estándar de Go
- **Arquitectura modular**: Fácil de extender con nuevos proveedores de contexto
- **Modo de solo lectura**: Propone cambios sin ejecutarlos directamente

## Requisitos

- Go 1.22 o superior
- [Ollama](https://ollama.ai/) instalado y ejecutándose localmente
- Un modelo LLM descargado (por defecto: `llama3.2`)

## Instalación

### Desde código fuente

```bash
# Clonar el repositorio
git clone https://github.com/tu-usuario/ollama-cli.git
cd ollama-cli

# Compilar
make build

# Instalar globalmente (opcional)
make install
```

### Verificar instalación de Ollama

```bash
# Verificar que Ollama está corriendo
curl http://localhost:11434/api/tags

# Descargar un modelo (si no tienes uno)
ollama pull llama3.2
```

## Uso

```bash
# Uso básico
ollama-cli "explica este código"

# Hacer preguntas sobre el proyecto
ollama-cli "¿cuál es la estructura de este proyecto?"

# Solicitar sugerencias
ollama-cli "sugiere mejoras para el manejo de errores"
```

## Configuración

La herramienta se configura mediante variables de entorno:

| Variable | Valor por defecto | Descripción |
|----------|-------------------|-------------|
| `OLLAMA_MODEL` | `llama3.2` | Modelo LLM a utilizar |
| `OLLAMA_URL` | `http://localhost:11434` | URL del servicio Ollama |

### Ejemplos de configuración

```bash
# Usar un modelo diferente
OLLAMA_MODEL=codellama ollama-cli "analiza este código"

# Conectar a Ollama remoto
OLLAMA_URL=http://servidor:11434 ollama-cli "explica la arquitectura"
```

## Estructura del Proyecto

```
ollama-cli/
├── cmd/
│   └── ollama-cli/
│       └── main.go              # Punto de entrada
├── internal/
│   ├── cli/
│   │   └── app.go               # Orquestador principal
│   ├── llm/
│   │   ├── client.go            # Interfaz del cliente LLM
│   │   └── ollama.go            # Implementación de Ollama
│   ├── mcp/
│   │   ├── provider.go          # Interfaz de proveedores de contexto
│   │   ├── registry.go          # Registro de proveedores
│   │   ├── filesystem.go        # Proveedor de contexto del sistema de archivos
│   │   └── git.go               # Proveedor de contexto de Git
│   └── prompt/
│       └── builder.go           # Constructor de prompts
├── docs/
│   ├── ARCHITECTURE.md          # Documentación de arquitectura
│   └── TECHNICAL.md             # Documentación técnica
├── Makefile                     # Configuración de compilación
└── go.mod                       # Definición del módulo
```

## Comandos Make

| Comando | Descripción |
|---------|-------------|
| `make build` | Compila el binario en `./bin/ollama-cli` |
| `make run ARGS="tu pregunta"` | Compila y ejecuta con argumentos |
| `make clean` | Elimina artefactos de compilación |
| `make install` | Instala el binario en `/usr/local/bin/` |

## Documentación

- [Arquitectura del Sistema](docs/ARCHITECTURE.md) - Diagramas y diseño del sistema
- [Documentación Técnica](docs/TECHNICAL.md) - Detalles de implementación y APIs

## Cómo Funciona

1. **Recopilación de Contexto**: Al ejecutarse, la herramienta recopila información del directorio actual:
   - Lista de archivos (hasta 50 archivos, 3 niveles de profundidad)
   - Información de Git (rama actual, últimos commits, archivos modificados)

2. **Construcción del Prompt**: Combina el contexto recopilado con la pregunta del usuario en un prompt estructurado.

3. **Generación de Respuesta**: Envía el prompt a Ollama y transmite la respuesta en tiempo real.

```
┌─────────┐    Pregunta    ┌─────┐    ┌───────────────────┐    ┌─────────────────┐
│ Usuario │ ─────────────► │ CLI │ ──►│ Recopilar Contexto│ ──►│ Construir Prompt│
└─────────┘                └─────┘    └───────────────────┘    └────────┬────────┘
     ▲                                                                  │
     │                                                                  ▼
     │  Streaming      ┌───────────┐                          ┌────────────────┐
     └─────────────────│ Respuesta │◄─────────────────────────│  Ollama API    │
                       └───────────┘                          └────────────────┘
```

## Limitaciones

- Máximo 50 archivos analizados por ejecución
- Profundidad máxima de 3 niveles de directorio
- Directorios ignorados: `node_modules`, `vendor`, `__pycache__`, `dist`, `build`, `.git`, etc.
- Requiere Ollama ejecutándose localmente (o accesible por red)

## Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -am 'Agrega nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Abre un Pull Request

## Licencia

MIT License - ver [LICENSE](LICENSE) para más detalles.
