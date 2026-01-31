package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	var files []string
	count := 0

	err := filepath.WalkDir(workDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Skip hidden directories and common noise
		name := d.Name()
		if strings.HasPrefix(name, ".") || isIgnoredDir(name) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Respect depth limit
		rel, _ := filepath.Rel(workDir, path)
		depth := strings.Count(rel, string(os.PathSeparator))
		if depth > p.maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			files = append(files, rel)
			count++
			if count >= p.maxFiles {
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil {
		return ContextResult{Provider: p.Name()}, err
	}

	content := fmt.Sprintf("Files in working directory (%d shown):\n%s",
		len(files), strings.Join(files, "\n"))

	return ContextResult{
		Provider: p.Name(),
		Content:  content,
	}, nil
}

func isIgnoredDir(name string) bool {
	ignored := []string{
		"node_modules",
		"vendor",
		"__pycache__",
		"dist",
		"build",
		".git",
		"bin",
		"obj",
		"target",
	}
	for _, d := range ignored {
		if name == d {
			return true
		}
	}
	return false
}
