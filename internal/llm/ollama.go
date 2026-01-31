package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func (c *OllamaClient) Generate(ctx context.Context, req GenerateRequest, onChunk func(string)) error {
	body, err := json.Marshal(ollamaRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		System: req.System,
		Stream: true,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// Stream NDJSON response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunk ollamaResponse
		if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
			continue // Skip malformed lines
		}
		if chunk.Error != "" {
			return fmt.Errorf("ollama error: %s", chunk.Error)
		}
		if chunk.Response != "" {
			onChunk(chunk.Response)
		}
		if chunk.Done {
			break
		}
	}

	return scanner.Err()
}
