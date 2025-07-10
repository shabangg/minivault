package llm

import (
	"context"
	"fmt"
	"io"
)

// LLM defines the interface for language model interactions
type LLM interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string, writer io.Writer) error
}

// Config holds LLM configuration
type Config struct {
	Type  string // "ollama" or "stub"
	URL   string // base URL for API calls
	Model string // model name
}

// NewLLM creates a new LLM instance based on configuration
func NewLLM(config Config) (LLM, error) {
	switch config.Type {
	case "ollama":
		if config.URL == "" {
			return nil, fmt.Errorf("OLLAMA_HOST is not set")
		}
		if config.Model == "" {
			return nil, fmt.Errorf("OLLAMA_MODEL is not set")
		}
		return NewOllamaLLM(config.URL, config.Model), nil
	case "stub":
		return NewStubLLM(), nil
	default:
		return nil, fmt.Errorf("unsupported LLM type: %s", config.Type)
	}
}
