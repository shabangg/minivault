package service

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGeneratorService(t *testing.T) {
	tests := []struct {
		name    string
		llmType string
		envVars map[string]string
	}{
		{
			name:    "Create with stub type",
			llmType: "stub",
			envVars: map[string]string{},
		},
		{
			name:    "Create with ollama type",
			llmType: "ollama",
			envVars: map[string]string{
				"OLLAMA_HOST":  "http://localhost:11434",
				"OLLAMA_MODEL": "test-model",
			},
		},
		{
			name:    "Invalid type falls back to stub",
			llmType: "invalid",
			envVars: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Create service
			service := NewGeneratorService(tt.llmType)
			assert.NotNil(t, service)
			assert.NotNil(t, service.llmService)
		})
	}
}

type mockWriter struct {
	written []byte
	header  http.Header
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		header: make(http.Header),
	}
}

func (w *mockWriter) Header() http.Header {
	return w.header
}

func (w *mockWriter) Write(p []byte) (n int, err error) {
	w.written = append(w.written, p...)
	return len(p), nil
}

func (w *mockWriter) WriteHeader(statusCode int) {
	// No-op for testing
}

func (w *mockWriter) Flush() {
	// No-op for testing
}

func TestGeneratorService_Generate(t *testing.T) {
	// Create service with stub LLM
	service := NewGeneratorService("stub")

	// Test generation
	ctx := context.Background()
	response, err := service.Generate(ctx, "test prompt")
	assert.NoError(t, err)
	assert.Contains(t, response, "test prompt") // Stub should include the prompt in response
}

func TestGeneratorService_GenerateStream(t *testing.T) {
	// Create service with stub LLM
	service := NewGeneratorService("stub")

	// Create mock writer
	writer := newMockWriter()

	// Test streaming
	ctx := context.Background()
	err := service.GenerateStream(ctx, "test prompt", writer)
	assert.NoError(t, err)
	assert.Contains(t, string(writer.written), "test prompt") // Stub should include the prompt in response
}

func TestChunkedWriter(t *testing.T) {
	var captured string
	onWrite := func(text string) {
		captured += text
	}

	// Create a mock http.ResponseWriter
	mockWriter := newMockWriter()
	writer := NewChunkedWriter(mockWriter, onWrite)

	// Test writing multiple chunks
	testData := []string{
		"First chunk",
		"Second chunk",
		"Third chunk",
	}

	for _, chunk := range testData {
		n, err := writer.Write([]byte(chunk))
		assert.NoError(t, err)
		assert.Equal(t, len(chunk), n)
	}

	// Verify the captured text
	assert.Equal(t, strings.Join(testData, ""), captured)

	// Verify the written data contains JSON responses
	writtenStr := string(mockWriter.written)
	lines := strings.Split(strings.TrimSpace(writtenStr), "\n")
	assert.Equal(t, len(testData), len(lines))

	for i, line := range lines {
		var response struct {
			Token string `json:"token"`
		}
		err := json.Unmarshal([]byte(line), &response)
		assert.NoError(t, err)
		assert.Equal(t, testData[i], response.Token)
	}
}
