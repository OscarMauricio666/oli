package config

// ============================================================================
// CONFIGURACIÓN DE OLI - Modifica estos valores según tus necesidades
// ============================================================================

// Modelo por defecto de Ollama
var Model = "qwen2.5-coder:14b"

// URL del servidor Ollama
var OllamaURL = "http://localhost:11434"

// Máximo de archivos a leer (contenido completo)
var MaxFiles = 30

// Profundidad máxima de carpetas a explorar
var MaxDepth = 4

// ============================================================================
// NOTA: Los prompts ahora están en internal/prompts/*.txt
// Para agregar un nuevo prompt, crea un archivo .txt en esa carpeta
// ============================================================================
