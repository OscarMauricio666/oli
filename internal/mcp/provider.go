package mcp

import "context"

// ContextProvider abstracts any source of context for the LLM.
// Implementations can wrap external CLIs, read files, or query services.
type ContextProvider interface {
	// Name returns a human-readable identifier for this provider.
	Name() string

	// Gather collects context relevant to the given working directory.
	// Returns structured text suitable for prompt injection.
	Gather(ctx context.Context, workDir string) (ContextResult, error)
}

// ContextResult holds the output from a context provider.
type ContextResult struct {
	// Provider is the name of the provider that generated this result.
	Provider string

	// Content is the context text to inject into the prompt.
	Content string

	// Error is set if the provider failed (Content may still be partial).
	Error string
}
