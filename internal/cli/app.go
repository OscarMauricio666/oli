package cli

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"ollama-cli/internal/config"
	"ollama-cli/internal/llm"
	"ollama-cli/internal/mcp"
	"ollama-cli/internal/prompt"
	"ollama-cli/internal/tools"
)

type App struct {
	model     string
	client    llm.Client
	providers []mcp.ContextProvider
	builder   *prompt.Builder
}

func New() *App {
	model := config.Model
	if env := os.Getenv("OLLAMA_MODEL"); env != "" {
		model = env
	}

	ollamaURL := config.OllamaURL
	if env := os.Getenv("OLLAMA_URL"); env != "" {
		ollamaURL = env
	}

	return &App{
		model:  model,
		client: llm.NewOllamaClient(ollamaURL),
		providers: []mcp.ContextProvider{
			mcp.NewFilesystemProvider(config.MaxFiles, config.MaxDepth),
			mcp.NewGitProvider(),
		},
		builder: prompt.NewBuilder(config.SystemPrompt),
	}
}

func NewWithPrompt(promptName string) *App {
	app := New()
	if p, ok := config.Prompts[promptName]; ok {
		app.builder = prompt.NewBuilder(p)
	}
	return app
}

func (a *App) Run(ctx context.Context, task string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	// 1. Detectar si el usuario menciona archivos específicos
	fileContents := a.extractFileContents(task)

	// 2. Gather context from all providers
	fmt.Fprintln(os.Stderr, "Recopilando contexto...")
	contexts := a.gatherContext(ctx, workDir)

	// 3. Agregar contenido de archivos mencionados
	if len(fileContents) > 0 {
		contexts = append(contexts, mcp.ContextResult{
			Provider: "archivos-leidos",
			Content:  fileContents,
		})
	}

	// 4. Build prompt
	system, user := a.builder.Build(contexts, task)

	// 5. Stream response y capturar para detectar código
	fmt.Fprintln(os.Stderr, "---")
	var fullResponse strings.Builder

	err = a.client.Generate(ctx, llm.GenerateRequest{
		Model:  a.model,
		System: system,
		Prompt: user,
	}, func(chunk string) {
		fmt.Print(chunk)
		fullResponse.WriteString(chunk)
	})
	fmt.Println()

	if err != nil {
		return err
	}

	// 6. Detectar bloques de código y ofrecer guardar
	a.offerToSaveCodeBlocks(fullResponse.String())

	return nil
}

// offerToSaveCodeBlocks detecta bloques de código en la respuesta y ofrece guardarlos
func (a *App) offerToSaveCodeBlocks(response string) {
	// Buscar bloques de código con nombre de archivo
	// Patrones como: ```go:filename.go o <!-- filename.go --> o // filename.go
	patterns := []*regexp.Regexp{
		// ```lang:filename.ext
		regexp.MustCompile("```\\w*:([\\w\\-./]+\\.[\\w]+)\\s*\\n([\\s\\S]*?)```"),
		// ```lang filename.ext
		regexp.MustCompile("```\\w*\\s+([\\w\\-./]+\\.[\\w]+)\\s*\\n([\\s\\S]*?)```"),
		// **filename.ext** seguido de bloque de código
		regexp.MustCompile("\\*\\*([\\w\\-./]+\\.[\\w]+)\\*\\*[:\\s]*\\n```\\w*\\n([\\s\\S]*?)```"),
		// `filename.ext`: seguido de bloque de código
		regexp.MustCompile("`([\\w\\-./]+\\.[\\w]+)`[:\\s]*\\n```\\w*\\n([\\s\\S]*?)```"),
		// Archivo: filename.ext seguido de bloque
		regexp.MustCompile("(?i)(?:archivo|file)[:\\s]+([\\w\\-./]+\\.[\\w]+)\\s*\\n```\\w*\\n([\\s\\S]*?)```"),
	}

	savedFiles := make(map[string]bool)

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(response, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				filename := match[1]
				content := strings.TrimSpace(match[2])

				// Evitar duplicados
				if savedFiles[filename] {
					continue
				}

				fmt.Printf("\n Código detectado para: %s\n", filename)
				if tools.AskConfirmation(fmt.Sprintf("¿Guardar archivo '%s'?", filename)) {
					if err := tools.WriteFileDirectly(filename, content); err != nil {
						fmt.Printf(" Error guardando: %v\n", err)
					} else {
						fmt.Printf(" Guardado: %s\n", filename)
						savedFiles[filename] = true
					}
				}
			}
		}
	}
}

// extractFileContents detecta menciones a archivos y pregunta si leerlos
func (a *App) extractFileContents(task string) string {
	patterns := []string{
		`[\w\-./]+\.(go|js|ts|py|java|c|cpp|h|rs|rb|php|html|css|json|yaml|yml|md|txt|sh|sql)`,
	}

	var files []string
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(task, -1)
		files = append(files, matches...)
	}

	if len(files) == 0 {
		return ""
	}

	// Eliminar duplicados
	seen := make(map[string]bool)
	var unique []string
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	var contents []string
	for _, file := range unique {
		if tools.AskConfirmation(fmt.Sprintf("¿Leer archivo '%s'?", file)) {
			content, err := tools.ReadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, " No se pudo leer %s: %v\n", file, err)
				continue
			}
			contents = append(contents, fmt.Sprintf("### Archivo: %s\n```\n%s\n```", file, content))
		}
	}

	return strings.Join(contents, "\n\n")
}

func (a *App) GetModel() string {
	return a.model
}

func (a *App) gatherContext(ctx context.Context, workDir string) []mcp.ContextResult {
	var results []mcp.ContextResult
	for _, p := range a.providers {
		result, err := p.Gather(ctx, workDir)
		if err != nil {
			result = mcp.ContextResult{
				Provider: p.Name(),
				Error:    err.Error(),
			}
		}
		results = append(results, result)
	}
	return results
}
