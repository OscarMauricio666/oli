package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"ollama-cli/internal/cli"
	"ollama-cli/internal/config"
	"ollama-cli/internal/tools"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Verificar comandos especiales
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "prompts":
			showPrompts()
			return
		case "help", "--help", "-h":
			showHelp()
			return
		case "read":
			if len(os.Args) >= 3 {
				readFileCmd(os.Args[2])
			} else {
				fmt.Println("Uso: oli read <archivo>")
			}
			return
		case "ls":
			path := "."
			if len(os.Args) >= 3 {
				path = os.Args[2]
			}
			listDirCmd(path)
			return
		case "repos":
			listReposCmd()
			return
		case "repo":
			if len(os.Args) >= 3 {
				reviewRepoCmd(ctx, os.Args[2], strings.Join(os.Args[3:], " "))
			} else {
				fmt.Println("Uso: oli repo <usuario/repo> [pregunta]")
			}
			return
		}
	}

	// Crear app
	var app *cli.App
	promptName := os.Getenv("OLI_PROMPT")
	if promptName != "" {
		app = cli.NewWithPrompt(promptName)
	} else {
		app = cli.New()
	}

	// Si hay argumentos, ejecutar una sola vez
	if len(os.Args) >= 2 {
		task := strings.Join(os.Args[1:], " ")
		if err := app.Run(ctx, task); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Modo interactivo
	runInteractive(ctx, app)
}

func runInteractive(ctx context.Context, app *cli.App) {
	fmt.Printf("\n oli (%s)\n", app.GetModel())
	fmt.Println(" Comandos: help | repos | repo <nombre> | ls | read | write | salir\n")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		select {
		case <-ctx.Done():
			fmt.Println("\n Hasta luego!")
			return
		default:
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "" {
			continue
		}

		// Procesar comandos especiales
		if handled := handleCommand(ctx, app, input); handled {
			continue
		}

		// Enviar al modelo
		if err := app.Run(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		fmt.Println()
	}
}

func handleCommand(ctx context.Context, app *cli.App, input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := parts[0]

	switch cmd {
	case "salir", "exit", "quit":
		fmt.Println(" Hasta luego!")
		os.Exit(0)
		return true

	case "help":
		showHelp()
		return true

	case "prompts":
		showPrompts()
		return true

	case "ls":
		path := "."
		if len(parts) >= 2 {
			path = parts[1]
		}
		listDirCmd(path)
		return true

	case "read":
		if len(parts) >= 2 {
			readFileCmd(parts[1])
		} else {
			fmt.Println(" Uso: read <archivo>")
		}
		return true

	case "write":
		if len(parts) >= 2 {
			writeFileCmd(parts[1])
		} else {
			fmt.Println(" Uso: write <archivo>")
		}
		return true

	case "pwd":
		wd, _ := os.Getwd()
		fmt.Println(wd)
		return true

	case "cd":
		if len(parts) >= 2 {
			if err := os.Chdir(parts[1]); err != nil {
				fmt.Printf(" Error: %v\n", err)
			} else {
				wd, _ := os.Getwd()
				fmt.Println(wd)
			}
		}
		return true

	case "repos":
		listReposCmd()
		return true

	case "repo":
		if len(parts) >= 2 {
			question := ""
			if len(parts) >= 3 {
				question = strings.Join(parts[2:], " ")
			}
			reviewRepoCmd(ctx, parts[1], question)
		} else {
			fmt.Println(" Uso: repo <usuario/repo> [pregunta]")
		}
		return true

	case "clone":
		if len(parts) >= 2 {
			cloneRepoCmd(parts[1])
		} else {
			fmt.Println(" Uso: clone <usuario/repo>")
		}
		return true

	case "issues":
		if len(parts) >= 2 {
			showIssuesCmd(parts[1])
		} else {
			fmt.Println(" Uso: issues <usuario/repo>")
		}
		return true

	case "prs":
		if len(parts) >= 2 {
			showPRsCmd(parts[1])
		} else {
			fmt.Println(" Uso: prs <usuario/repo>")
		}
		return true
	}

	return false
}

func readFileCmd(path string) {
	content, err := tools.ReadFile(path)
	if err != nil {
		fmt.Printf(" Error leyendo archivo: %v\n", err)
		return
	}
	fmt.Println("\n────────────────────────────────")
	fmt.Println(content)
	fmt.Println("────────────────────────────────")
}

func listDirCmd(path string) {
	files, err := tools.ListDir(path)
	if err != nil {
		fmt.Printf(" Error listando directorio: %v\n", err)
		return
	}
	fmt.Println()
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}
	fmt.Println()
}

func writeFileCmd(path string) {
	fmt.Println(" Escribe el contenido (termina con una línea que solo tenga 'FIN'):")
	fmt.Println("────────────────────────────────")

	scanner := bufio.NewScanner(os.Stdin)
	var lines []string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "FIN" {
			break
		}
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")

	if err := tools.WriteFile(path, content); err != nil {
		fmt.Printf(" Error: %v\n", err)
		return
	}
	fmt.Printf(" Archivo guardado: %s\n", path)
}

func listReposCmd() {
	fmt.Println("\n Obteniendo repositorios...")
	repos, err := tools.ListMyRepos(20)
	if err != nil {
		fmt.Printf(" Error: %v\n", err)
		return
	}
	fmt.Println("\n Tus repositorios:")
	fmt.Println("────────────────────────────────")
	fmt.Println(repos)
}

func reviewRepoCmd(ctx context.Context, repo string, question string) {
	fmt.Printf("\n Obteniendo info de %s...\n", repo)

	info, err := tools.GetRepoInfo(repo)
	if err != nil {
		fmt.Printf(" Error: %v\n", err)
		return
	}

	fmt.Println("────────────────────────────────")
	fmt.Println(tools.FormatRepoInfo(info))
	fmt.Println("────────────────────────────────")

	// Si hay pregunta, enviar al modelo con contexto del repo
	if question == "" {
		question = "analiza este repositorio y dame un resumen de qué hace, su estructura y sugerencias de mejora"
	}

	// Crear app con contexto del repo
	app := cli.New()
	repoContext := tools.FormatRepoInfo(info)
	if info.Readme != "" {
		repoContext += "\n\nREADME:\n" + info.Readme
	}

	fullQuestion := fmt.Sprintf("Contexto del repositorio GitHub:\n%s\n\nPregunta: %s", repoContext, question)

	if err := app.RunWithoutLocalContext(ctx, fullQuestion); err != nil {
		fmt.Printf(" Error: %v\n", err)
	}
	fmt.Println()
}

func cloneRepoCmd(repo string) {
	fmt.Printf("\n Clonando %s...\n", repo)
	if err := tools.CloneRepo(repo, ""); err != nil {
		fmt.Printf(" Error: %v\n", err)
		return
	}
	fmt.Printf(" Repositorio clonado exitosamente\n")
}

func showIssuesCmd(repo string) {
	fmt.Printf("\n Issues de %s:\n", repo)
	fmt.Println("────────────────────────────────")
	issues, err := tools.GetRepoIssues(repo)
	if err != nil {
		fmt.Printf(" Error: %v\n", err)
		return
	}
	if issues == "" {
		fmt.Println(" No hay issues abiertos")
	} else {
		fmt.Println(issues)
	}
}

func showPRsCmd(repo string) {
	fmt.Printf("\n Pull Requests de %s:\n", repo)
	fmt.Println("────────────────────────────────")
	prs, err := tools.GetRepoPRs(repo)
	if err != nil {
		fmt.Printf(" Error: %v\n", err)
		return
	}
	if prs == "" {
		fmt.Println(" No hay PRs abiertos")
	} else {
		fmt.Println(prs)
	}
}

func showPrompts() {
	fmt.Println("\n Prompts disponibles:")
	fmt.Println(" ────────────────────")
	for name := range config.Prompts {
		fmt.Printf("   • %s\n", name)
	}
	fmt.Println("\n Uso: OLI_PROMPT=code-review oli")
	fmt.Println(" Editar: internal/config/config.go\n")
}

func showHelp() {
	fmt.Println(`
 oli - Asistente de código con Ollama

 MODO INTERACTIVO:
   oli                       Inicia modo interactivo

 COMANDOS LOCALES:
   help                      Esta ayuda
   ls [dir]                  Listar archivos
   read <archivo>            Leer archivo
   write <archivo>           Escribir archivo
   pwd                       Directorio actual
   cd <dir>                  Cambiar directorio

 COMANDOS GITHUB:
   repos                     Listar mis repositorios
   repo <user/repo>          Analizar repositorio
   repo <user/repo> <pregunta>  Preguntar sobre repo
   clone <user/repo>         Clonar repositorio
   issues <user/repo>        Ver issues abiertos
   prs <user/repo>           Ver PRs abiertos

 MODO DIRECTO:
   oli <pregunta>            Pregunta única
   oli repo <user/repo>      Analizar repo directamente

 CONFIGURACIÓN:
   Editar: internal/config/config.go

 VARIABLES DE ENTORNO:
   OLLAMA_MODEL              Modelo a usar
   OLLAMA_URL                URL de Ollama
   OLI_PROMPT                Prompt a usar

 EJEMPLOS:
   oli que hace este proyecto
   oli repo OscarMauricio666/oli
   oli repo facebook/react que tecnologias usa
`)
}
