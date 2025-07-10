package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaLLM_Generate(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Parse request body
		var req ollamaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-model", req.Model)
		assert.Equal(t, "test prompt", req.Prompt)
		assert.False(t, req.Stream)

		// Send response
		response := ollamaResponse{
			Response: "test response",
			Done:     true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create LLM with test server URL
	llm := NewOllamaLLM(server.URL, "test-model")
	ctx := context.Background()

	// Test generation
	response, err := llm.Generate(ctx, "test prompt")
	assert.NoError(t, err)
	assert.Equal(t, "test response", response)
}

func TestOllamaLLM_GenerateStream(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "/api/generate", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Parse request body
		var req ollamaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-model", req.Model)
		assert.Equal(t, "test prompt", req.Prompt)
		assert.True(t, req.Stream)

		// Send streamed responses
		responses := []ollamaResponse{
			{Response: "test", Done: false},
			{Response: " response", Done: true},
		}

		for _, resp := range responses {
			json.NewEncoder(w).Encode(resp)
			w.(http.Flusher).Flush()
		}
	}))
	defer server.Close()

	// Create LLM with test server URL
	llm := NewOllamaLLM(server.URL, "test-model")
	ctx := context.Background()

	// Test streaming
	var buf bytes.Buffer
	err := llm.GenerateStream(ctx, "test prompt", &buf)
	assert.NoError(t, err)
	assert.Equal(t, "test response", buf.String())
}

func TestOllamaLLM_GenerateError(t *testing.T) {
	// Create test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("test error"))
	}))
	defer server.Close()

	// Create LLM with test server URL
	llm := NewOllamaLLM(server.URL, "test-model")
	ctx := context.Background()

	// Test generation error
	_, err := llm.Generate(ctx, "test prompt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 500")

	// Test streaming error
	var buf bytes.Buffer
	err = llm.GenerateStream(ctx, "test prompt", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code: 500")
}
