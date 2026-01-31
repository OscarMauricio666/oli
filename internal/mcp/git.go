package mcp

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type GitProvider struct{}

func NewGitProvider() *GitProvider {
	return &GitProvider{}
}

func (p *GitProvider) Name() string {
	return "git"
}

func (p *GitProvider) Gather(ctx context.Context, workDir string) (ContextResult, error) {
	// Check if this is a git repo
	if _, err := p.runGit(ctx, workDir, "rev-parse", "--git-dir"); err != nil {
		return ContextResult{
			Provider: p.Name(),
			Content:  "Not a git repository.",
		}, nil
	}

	var sections []string

	// Current branch
	if branch, err := p.runGit(ctx, workDir, "branch", "--show-current"); err == nil {
		sections = append(sections, fmt.Sprintf("Branch: %s", strings.TrimSpace(branch)))
	}

	// Recent commits (last 5)
	if log, err := p.runGit(ctx, workDir, "log", "--oneline", "-5"); err == nil {
		log = strings.TrimSpace(log)
		if log != "" {
			sections = append(sections, fmt.Sprintf("Recent commits:\n%s", log))
		}
	}

	// Status summary
	if status, err := p.runGit(ctx, workDir, "status", "--short"); err == nil {
		if status = strings.TrimSpace(status); status != "" {
			sections = append(sections, fmt.Sprintf("Changed files:\n%s", status))
		} else {
			sections = append(sections, "Working tree clean.")
		}
	}

	return ContextResult{
		Provider: p.Name(),
		Content:  strings.Join(sections, "\n\n"),
	}, nil
}

func (p *GitProvider) runGit(ctx context.Context, workDir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
