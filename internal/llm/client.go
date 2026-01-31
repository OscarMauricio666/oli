package llm

import "context"

// Client abstracts LLM interactions.
// Designed for Ollama but can support other backends.
type Client interface {
	// Generate sends a prompt and streams the response.
	// The callback is invoked for each chunk of text received.
	Generate(ctx context.Context, req GenerateRequest, onChunk func(chunk string)) error
}

// GenerateRequest contains the parameters for generation.
type GenerateRequest struct {
	Model  string
	Prompt string
	System string // Optional system prompt
}
