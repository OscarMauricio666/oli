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
	fmt.Println(" Comandos: salir | help | ls [dir] | read <archivo> | write <archivo>\n")

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
		if handled := handleCommand(input); handled {
			continue
		}

		// Enviar al modelo
		if err := app.Run(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		fmt.Println()
	}
}

func handleCommand(input string) bool {
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
   oli                     Inicia modo interactivo

 COMANDOS EN MODO INTERACTIVO:
   help                    Esta ayuda
   prompts                 Ver prompts disponibles
   ls [dir]                Listar archivos
   read <archivo>          Leer contenido de archivo
   write <archivo>         Escribir archivo (con confirmación)
   pwd                     Directorio actual
   cd <dir>                Cambiar directorio
   salir                   Salir

 MODO DIRECTO:
   oli <pregunta>          Pregunta única
   oli read <archivo>      Leer archivo
   oli ls [dir]            Listar directorio

 CONFIGURACIÓN:
   Editar: internal/config/config.go

 VARIABLES DE ENTORNO:
   OLLAMA_MODEL            Modelo a usar
   OLLAMA_URL              URL de Ollama
   OLI_PROMPT              Prompt (default, code-review, etc.)

 EJEMPLOS:
   oli que hace este proyecto
   oli read main.go
   OLI_PROMPT=code-review oli
`)
}
