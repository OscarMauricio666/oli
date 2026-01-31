package prompts

import (
	"embed"
	"strings"
)

//go:embed *.txt
var promptFiles embed.FS

// Prompts disponibles (cargados de archivos .txt)
var Prompts map[string]string

// SystemPrompt es el prompt por defecto
var SystemPrompt string

func init() {
	Prompts = make(map[string]string)

	// Cargar todos los archivos .txt
	entries, err := promptFiles.ReadDir(".")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".txt") {
			continue
		}

		content, err := promptFiles.ReadFile(name)
		if err != nil {
			continue
		}

		// Nombre sin extensión
		promptName := strings.TrimSuffix(name, ".txt")
		Prompts[promptName] = string(content)
	}

	// El prompt por defecto
	if p, ok := Prompts["default"]; ok {
		SystemPrompt = p
	}
}

// Get obtiene un prompt por nombre, retorna el default si no existe
func Get(name string) string {
	if p, ok := Prompts[name]; ok {
		return p
	}
	return SystemPrompt
}

// List retorna los nombres de todos los prompts disponibles
func List() []string {
	names := make([]string, 0, len(Prompts))
	for name := range Prompts {
		names = append(names, name)
	}
	return names
}
