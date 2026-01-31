package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AskConfirmation pregunta al usuario y espera confirmación
func AskConfirmation(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n %s (s/n): ", question)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "s" || response == "si" || response == "y" || response == "yes"
}

// ReadFile lee el contenido de un archivo
func ReadFile(path string) (string, error) {
	// Convertir a ruta absoluta si es relativa
	if !filepath.IsAbs(path) {
		wd, _ := os.Getwd()
		path = filepath.Join(wd, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// WriteFile escribe contenido a un archivo (con confirmación)
func WriteFile(path string, content string) error {
	// Convertir a ruta absoluta si es relativa
	if !filepath.IsAbs(path) {
		wd, _ := os.Getwd()
		path = filepath.Join(wd, path)
	}

	// Verificar si existe
	exists := false
	if _, err := os.Stat(path); err == nil {
		exists = true
	}

	action := "crear"
	if exists {
		action = "sobrescribir"
	}

	if !AskConfirmation(fmt.Sprintf("¿%s archivo %s?", action, path)) {
		return fmt.Errorf("operación cancelada por el usuario")
	}

	// Crear directorio si no existe
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// WriteFileDirectly escribe sin pedir confirmación (ya se pidió antes)
func WriteFileDirectly(path string, content string) error {
	if !filepath.IsAbs(path) {
		wd, _ := os.Getwd()
		path = filepath.Join(wd, path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// ListDir lista archivos en un directorio
func ListDir(path string) ([]string, error) {
	if path == "" || path == "." {
		path, _ = os.Getwd()
	}

	if !filepath.IsAbs(path) {
		wd, _ := os.Getwd()
		path = filepath.Join(wd, path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		files = append(files, name)
	}
	return files, nil
}
