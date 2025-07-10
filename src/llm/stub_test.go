package llm

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubLLM_Generate(t *testing.T) {
	llm := NewStubLLM()
	ctx := context.Background()
	prompt := "test prompt"

	response, err := llm.Generate(ctx, prompt)
	assert.NoError(t, err)
	assert.Contains(t, response, prompt)
}

func TestStubLLM_GenerateStream(t *testing.T) {
	llm := NewStubLLM()
	ctx := context.Background()
	prompt := "test prompt"
	var buf bytes.Buffer

	err := llm.GenerateStream(ctx, prompt, &buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), prompt)
}
