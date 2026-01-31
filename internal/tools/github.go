package tools

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitHubInfo contiene información de un repositorio
type GitHubInfo struct {
	Name        string
	Description string
	Language    string
	Stars       string
	Forks       string
	Issues      string
	URL         string
	Files       string
	Readme      string
}

// GetRepoInfo obtiene información de un repositorio de GitHub
func GetRepoInfo(repo string) (*GitHubInfo, error) {
	info := &GitHubInfo{}

	// Obtener info básica del repo
	cmd := exec.Command("gh", "repo", "view", repo, "--json", "name,description,primaryLanguage,stargazerCount,forkCount,url,openIssuesCount")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo info del repo: %v", err)
	}

	// Parsear JSON manualmente (sin dependencias)
	s := string(output)
	info.Name = extractJSON(s, "name")
	info.Description = extractJSON(s, "description")
	info.Language = extractJSON(s, "primaryLanguage")
	info.Stars = extractJSON(s, "stargazerCount")
	info.Forks = extractJSON(s, "forkCount")
	info.Issues = extractJSON(s, "openIssuesCount")
	info.URL = extractJSON(s, "url")

	// Obtener lista de archivos
	cmd = exec.Command("gh", "api", fmt.Sprintf("repos/%s/git/trees/HEAD", repo), "--jq", ".tree[].path")
	if output, err := cmd.Output(); err == nil {
		files := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(files) > 30 {
			files = append(files[:30], fmt.Sprintf("... y %d archivos más", len(files)-30))
		}
		info.Files = strings.Join(files, "\n")
	}

	// Obtener README usando gh repo view
	cmd = exec.Command("gh", "repo", "view", repo)
	if output, err := cmd.Output(); err == nil {
		info.Readme = string(output)
	}

	return info, nil
}

// ListMyRepos lista los repositorios del usuario
func ListMyRepos(limit int) (string, error) {
	cmd := exec.Command("gh", "repo", "list", "--limit", fmt.Sprintf("%d", limit))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetRepoContent obtiene el contenido de un archivo del repo
func GetRepoContent(repo, path string) (string, error) {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/%s/contents/%s", repo, path), "--jq", ".content")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Decodificar base64
	content := strings.TrimSpace(string(output))
	cmd = exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | base64 -d", content))
	decoded, err := cmd.Output()
	if err != nil {
		return content, nil
	}
	return string(decoded), nil
}

// GetRepoPRs obtiene los PRs abiertos
func GetRepoPRs(repo string) (string, error) {
	cmd := exec.Command("gh", "pr", "list", "--repo", repo)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetRepoIssues obtiene los issues abiertos
func GetRepoIssues(repo string) (string, error) {
	cmd := exec.Command("gh", "issue", "list", "--repo", repo)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// CloneRepo clona un repositorio
func CloneRepo(repo, dir string) error {
	args := []string{"repo", "clone", repo}
	if dir != "" {
		args = append(args, dir)
	}
	cmd := exec.Command("gh", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// extractJSON extrae un valor de un JSON simple
func extractJSON(json, key string) string {
	// Buscar "key": o "key":
	patterns := []string{
		fmt.Sprintf(`"%s":"`, key),
		fmt.Sprintf(`"%s": "`, key),
		fmt.Sprintf(`"%s":`, key),
		fmt.Sprintf(`"%s": `, key),
	}

	for _, pattern := range patterns {
		idx := strings.Index(json, pattern)
		if idx == -1 {
			continue
		}

		start := idx + len(pattern)
		rest := json[start:]

		// Si empieza con comilla, buscar el cierre
		if len(rest) > 0 && rest[0] == '"' {
			rest = rest[1:]
			end := strings.Index(rest, `"`)
			if end != -1 {
				return rest[:end]
			}
		}

		// Si es un número o null
		end := strings.IndexAny(rest, ",}")
		if end != -1 {
			val := strings.TrimSpace(rest[:end])
			val = strings.Trim(val, `"`)
			if val == "null" {
				return ""
			}
			return val
		}
	}

	return ""
}

// FormatRepoInfo formatea la info del repo para el contexto
func FormatRepoInfo(info *GitHubInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Repositorio: %s\n", info.Name))
	sb.WriteString(fmt.Sprintf("URL: %s\n", info.URL))
	if info.Description != "" {
		sb.WriteString(fmt.Sprintf("Descripción: %s\n", info.Description))
	}
	if info.Language != "" {
		sb.WriteString(fmt.Sprintf("Lenguaje: %s\n", info.Language))
	}
	sb.WriteString(fmt.Sprintf("Stars: %s | Forks: %s | Issues: %s\n", info.Stars, info.Forks, info.Issues))

	if info.Files != "" {
		sb.WriteString(fmt.Sprintf("\nArchivos:\n%s\n", info.Files))
	}

	return sb.String()
}
