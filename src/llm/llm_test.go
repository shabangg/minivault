package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLLM(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "Valid Ollama config",
			config: Config{
				Type:  "ollama",
				URL:   "http://localhost:11434",
				Model: "test-model",
			},
			wantError: false,
		},
		{
			name: "Missing Ollama URL",
			config: Config{
				Type:  "ollama",
				Model: "test-model",
			},
			wantError: true,
		},
		{
			name: "Missing Ollama model",
			config: Config{
				Type: "ollama",
				URL:  "http://localhost:11434",
			},
			wantError: true,
		},
		{
			name: "Valid stub config",
			config: Config{
				Type: "stub",
			},
			wantError: false,
		},
		{
			name: "Invalid type",
			config: Config{
				Type: "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm, err := NewLLM(tt.config)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, llm)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, llm)

				switch tt.config.Type {
				case "ollama":
					_, ok := llm.(*OllamaLLM)
					assert.True(t, ok, "Expected OllamaLLM type")
				case "stub":
					_, ok := llm.(*StubLLM)
					assert.True(t, ok, "Expected StubLLM type")
				}
			}
		})
	}
}
