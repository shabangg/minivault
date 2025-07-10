package llm

import (
	"context"
	"fmt"
	"io"
	"time"
)

type StubLLM struct{}

func NewStubLLM() *StubLLM {
	return &StubLLM{}
}

func (l *StubLLM) Generate(_ context.Context, prompt string) (string, error) {
	return fmt.Sprintf("This is a stubbed response to your prompt: %s", prompt), nil
}

func (l *StubLLM) GenerateStream(_ context.Context, prompt string, writer io.Writer) error {
	words := []string{"This", "is", "a", "stubbed", "streaming", "response", "to", "your", "prompt:", prompt}

	for _, word := range words {
		if _, err := fmt.Fprintf(writer, "%s\n", word); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond) // Simulate streaming delay
	}

	return nil
}
