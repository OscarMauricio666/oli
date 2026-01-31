package prompt

import (
	"fmt"
	"strings"

	"ollama-cli/internal/mcp"
)

type Builder struct {
	systemPrompt string
}

func NewBuilder(systemPrompt string) *Builder {
	return &Builder{systemPrompt: systemPrompt}
}

func (b *Builder) Build(contexts []mcp.ContextResult, task string) (system, user string) {
	system = b.systemPrompt

	var parts []string

	// Add context sections
	for _, ctx := range contexts {
		if ctx.Error != "" {
			parts = append(parts, fmt.Sprintf("## Context: %s\n[Error: %s]", ctx.Provider, ctx.Error))
		} else if ctx.Content != "" {
			parts = append(parts, fmt.Sprintf("## Context: %s\n%s", ctx.Provider, ctx.Content))
		}
	}

	// Add user task
	parts = append(parts, fmt.Sprintf("## Task\n%s", task))

	user = strings.Join(parts, "\n\n")
	return system, user
}
