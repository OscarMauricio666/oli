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

	// 1. Recopilar contexto automáticamente (lee archivos del proyecto)
	fmt.Fprintln(os.Stderr, "Leyendo proyecto...")
	contexts := a.gatherContext(ctx, workDir)

	// 2. Construir prompt
	system, user := a.builder.Build(contexts, task)

	// 3. Stream response y capturar para detectar código
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

	// 4. Detectar bloques de código y ofrecer guardar
	a.offerToSaveCodeBlocks(fullResponse.String())

	return nil
}

// offerToSaveCodeBlocks detecta bloques de código en la respuesta y ofrece guardarlos
func (a *App) offerToSaveCodeBlocks(response string) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile("```\\w*:([\\w\\-./]+\\.[\\w]+)\\s*\\n([\\s\\S]*?)```"),
		regexp.MustCompile("```\\w*\\s+([\\w\\-./]+\\.[\\w]+)\\s*\\n([\\s\\S]*?)```"),
		regexp.MustCompile("\\*\\*([\\w\\-./]+\\.[\\w]+)\\*\\*[:\\s]*\\n```\\w*\\n([\\s\\S]*?)```"),
		regexp.MustCompile("`([\\w\\-./]+\\.[\\w]+)`[:\\s]*\\n```\\w*\\n([\\s\\S]*?)```"),
		regexp.MustCompile("(?i)(?:archivo|file)[:\\s]+([\\w\\-./]+\\.[\\w]+)\\s*\\n```\\w*\\n([\\s\\S]*?)```"),
	}

	savedFiles := make(map[string]bool)

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(response, -1)
		for _, match := range matches {
			if len(match) >= 3 {
				filename := match[1]
				content := strings.TrimSpace(match[2])

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
