package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Configuración de límites
const (
	maxFileSize    = 50000  // 50KB máximo por archivo
	maxTotalSize   = 200000 // 200KB máximo total de contenido
)

// Extensiones de archivos que se leen automáticamente
var readableExtensions = map[string]bool{
	".go":     true,
	".js":     true,
	".ts":     true,
	".jsx":    true,
	".tsx":    true,
	".py":     true,
	".java":   true,
	".c":      true,
	".cpp":    true,
	".h":      true,
	".hpp":    true,
	".rs":     true,
	".rb":     true,
	".php":    true,
	".swift":  true,
	".kt":     true,
	".scala":  true,
	".cs":     true,
	".html":   true,
	".css":    true,
	".scss":   true,
	".json":   true,
	".yaml":   true,
	".yml":    true,
	".toml":   true,
	".xml":    true,
	".md":     true,
	".txt":    true,
	".sh":     true,
	".bash":   true,
	".zsh":    true,
	".sql":    true,
	".graphql": true,
	".proto":  true,
	".env.example": true,
	".gitignore": true,
	".dockerignore": true,
	"Makefile": true,
	"Dockerfile": true,
	"Gemfile":  true,
	"Rakefile": true,
}

// Directorios a ignorar
var ignoredDirs = map[string]bool{
	"node_modules":    true,
	"vendor":          true,
	"__pycache__":     true,
	"dist":            true,
	"build":           true,
	".git":            true,
	"bin":             true,
	"obj":             true,
	"target":          true,
	".idea":           true,
	".vscode":         true,
	"coverage":        true,
	".next":           true,
	".nuxt":           true,
	"venv":            true,
	".venv":           true,
	"env":             true,
	".env":            true,
	"__snapshots__":   true,
	".cache":          true,
	".parcel-cache":   true,
	".turbo":          true,
	"tmp":             true,
	"temp":            true,
	"logs":            true,
	".pytest_cache":   true,
	".mypy_cache":     true,
	".tox":            true,
	"htmlcov":         true,
	".coverage":       true,
	"eggs":            true,
	".eggs":           true,
	"wheels":          true,
	"pip-wheel-metadata": true,
	"*.egg-info":      true,
	".installed.cfg":  true,
	"lib":             true,
	"lib64":           true,
	"parts":           true,
	"sdist":           true,
	"var":             true,
	".sass-cache":     true,
	"bower_components": true,
	"jspm_packages":   true,
	".npm":            true,
	".yarn":           true,
	".pnp":            true,
}

// Archivos a ignorar
var ignoredFiles = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"Gemfile.lock":      true,
	"Cargo.lock":        true,
	"poetry.lock":       true,
	"composer.lock":     true,
	"go.sum":            true,
	".DS_Store":         true,
	"Thumbs.db":         true,
}

type FilesystemProvider struct {
	maxFiles int
	maxDepth int
}

func NewFilesystemProvider(maxFiles, maxDepth int) *FilesystemProvider {
	return &FilesystemProvider{
		maxFiles: maxFiles,
		maxDepth: maxDepth,
	}
}

func (p *FilesystemProvider) Name() string {
	return "filesystem"
}

func (p *FilesystemProvider) Gather(ctx context.Context, workDir string) (ContextResult, error) {
	var fileContents []string
	var fileList []string
	count := 0
	totalSize := 0

	err := filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		name := d.Name()

		// Ignorar directorios
		if d.IsDir() {
			if strings.HasPrefix(name, ".") || ignoredDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		// Ignorar archivos ocultos y de lock
		if strings.HasPrefix(name, ".") || ignoredFiles[name] {
			return nil
		}

		// Verificar profundidad
		rel, _ := filepath.Rel(workDir, path)
		depth := strings.Count(rel, string(os.PathSeparator))
		if depth > p.maxDepth {
			return nil
		}

		// Verificar si alcanzamos el límite de archivos
		if count >= p.maxFiles {
			return filepath.SkipAll
		}

		// Verificar si es un archivo legible
		ext := filepath.Ext(name)
		isReadable := readableExtensions[ext] || readableExtensions[name]

		if isReadable {
			// Verificar tamaño del archivo
			info, err := d.Info()
			if err != nil {
				return nil
			}

			fileSize := int(info.Size())
			if fileSize > maxFileSize {
				fileList = append(fileList, fmt.Sprintf("%s (muy grande: %dKB)", rel, fileSize/1024))
				count++
				return nil
			}

			// Verificar límite total
			if totalSize+fileSize > maxTotalSize {
				fileList = append(fileList, fmt.Sprintf("%s (omitido por límite de contexto)", rel))
				count++
				return nil
			}

			// Leer contenido
			content, err := os.ReadFile(path)
			if err != nil {
				fileList = append(fileList, fmt.Sprintf("%s (error al leer)", rel))
				count++
				return nil
			}

			fileContents = append(fileContents, fmt.Sprintf("### %s\n```\n%s\n```", rel, string(content)))
			totalSize += fileSize
			count++
		} else {
			// Solo listar el archivo sin leer contenido
			fileList = append(fileList, rel)
			count++
		}

		return nil
	})

	if err != nil {
		return ContextResult{Provider: p.Name()}, err
	}

	// Construir resultado
	var sb strings.Builder

	if len(fileContents) > 0 {
		sb.WriteString("## Contenido de archivos del proyecto\n\n")
		sb.WriteString(strings.Join(fileContents, "\n\n"))
	}

	if len(fileList) > 0 {
		sb.WriteString("\n\n## Otros archivos (sin contenido)\n")
		sb.WriteString(strings.Join(fileList, "\n"))
	}

	return ContextResult{
		Provider: p.Name(),
		Content:  sb.String(),
	}, nil
}
