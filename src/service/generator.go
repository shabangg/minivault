package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"minivault-api/src/llm"
)

// Generator interface defines the contract for text generation services
type Generator interface {
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string, writer io.Writer) error
}

// GeneratorService provides text generation with automatic fallback
type GeneratorService struct {
	llmService llm.LLM
}

// NewGeneratorService creates a new generator service
func NewGeneratorService(llmType string) *GeneratorService {
	config := llm.Config{
		Type:  llmType,
		URL:   os.Getenv("OLLAMA_HOST"),
		Model: os.Getenv("OLLAMA_MODEL"),
	}

	// Try to create LLM service, fallback to stub if fails
	llmService, err := llm.NewLLM(config)
	if err != nil {
		llmService, _ = llm.NewLLM(llm.Config{Type: "stub"})
	}

	return &GeneratorService{
		llmService: llmService,
	}
}

// Generate returns a response from the LLM
func (g *GeneratorService) Generate(ctx context.Context, prompt string) (string, error) {
	return g.llmService.Generate(ctx, prompt)
}

// GenerateStream streams responses from the LLM
func (g *GeneratorService) GenerateStream(ctx context.Context, prompt string, writer io.Writer) error {
	return g.llmService.GenerateStream(ctx, prompt, writer)
}

// ChunkedWriter implements io.Writer for chunked transfer encoding
type ChunkedWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
	onWrite func(string)
}

// TokenResponse represents a single token in the stream
type TokenResponse struct {
	Token string `json:"token"`
}

// NewChunkedWriter creates a new chunked transfer writer
func NewChunkedWriter(w http.ResponseWriter, onWrite func(string)) *ChunkedWriter {
	w.Header().Set("Content-Type", "application/json")
	// Content-Length is intentionally not set to enable chunked transfer

	return &ChunkedWriter{
		w:       w,
		flusher: w.(http.Flusher),
		onWrite: onWrite,
	}
}

// Write implements io.Writer
func (w *ChunkedWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if w.onWrite != nil {
		w.onWrite(data)
	}

	// Send token as newline-delimited JSON
	response := TokenResponse{Token: data}
	jsonData, err := json.Marshal(response)
	if err != nil {
		return 0, err
	}

	if _, err := fmt.Fprintf(w.w, "%s\n", jsonData); err != nil {
		return 0, err
	}
	w.flusher.Flush()
	return len(p), nil
}
