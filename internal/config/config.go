package config

// ============================================================================
// CONFIGURACIÓN DE OLI - Modifica estos valores según tus necesidades
// ============================================================================

// Modelo por defecto de Ollama
var Model = "qwen2.5-coder:14b"

// URL del servidor Ollama
var OllamaURL = "http://localhost:11434"

// Máximo de archivos a mostrar en el contexto
var MaxFiles = 50

// Profundidad máxima de carpetas a explorar
var MaxDepth = 3

// ============================================================================
// PROMPTS DEL SISTEMA - Personaliza el comportamiento del asistente
// ============================================================================

// Prompt principal del sistema
var SystemPrompt = `Eres un asistente de programación experto. Analizas código y das sugerencias.

REGLAS:
- Cuando crees o modifiques archivos, usa este formato para que se puedan guardar automáticamente:

**nombre_archivo.ext**
` + "```" + `lenguaje
contenido del archivo
` + "```" + `

- Siempre muestra el contenido completo del archivo.
- Cuando sugieras comandos, explica qué hacen.

Sé conciso. Enfócate en la tarea específica del usuario.
Responde en español.`

// ============================================================================
// PROMPTS ADICIONALES - Puedes agregar más prompts para diferentes usos
// ============================================================================

var Prompts = map[string]string{
	"default": SystemPrompt,

	"code-review": `Eres un revisor de código experto. Tu trabajo es:
- Identificar bugs y problemas potenciales
- Sugerir mejoras de rendimiento
- Verificar mejores prácticas
- Detectar problemas de seguridad
Sé específico y muestra ejemplos de código corregido.
Responde en español.`,

	"explainer": `Eres un profesor de programación paciente. Tu trabajo es:
- Explicar código de forma clara y sencilla
- Usar analogías cuando sea útil
- Dividir conceptos complejos en partes simples
- Dar ejemplos prácticos
Responde en español.`,

	"architect": `Eres un arquitecto de software senior. Tu trabajo es:
- Analizar la estructura del proyecto
- Sugerir patrones de diseño apropiados
- Identificar problemas de arquitectura
- Proponer mejoras escalables
Responde en español.`,
}
