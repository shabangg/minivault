package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggingService(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	tests := []struct {
		name    string
		logPath string
		llmType string
		wantErr bool
	}{
		{
			name:    "Valid path and type",
			logPath: logPath,
			llmType: "stub",
			wantErr: false,
		},
		{
			name:    "Invalid path",
			logPath: "/nonexistent/directory/test.log",
			llmType: "stub",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLoggingService(tt.logPath, tt.llmType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				assert.NoError(t, logger.Close())
			}
		})
	}
}

func TestLoggingService_LogInteraction(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	logger, err := NewLoggingService(logPath, "stub")
	assert.NoError(t, err)
	defer logger.Close()

	// Test logging interaction
	prompt := "test prompt"
	response := "test response"
	streaming := false

	err = logger.LogInteraction(prompt, response, streaming)
	assert.NoError(t, err)

	// Read log file and verify content
	logData, err := os.ReadFile(logPath)
	assert.NoError(t, err)

	var entry LogEntry
	err = json.Unmarshal(logData, &entry)
	assert.NoError(t, err)

	assert.Equal(t, prompt, entry.Prompt)
	assert.Equal(t, response, entry.Response)
	assert.Equal(t, streaming, entry.Streaming)
	assert.Equal(t, "stub", entry.LLMType)
	assert.True(t, entry.Success)
}

func TestLoggingService_LogError(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	logger, err := NewLoggingService(logPath, "stub")
	assert.NoError(t, err)
	defer logger.Close()

	// Test logging error
	prompt := "test prompt"
	testErr := errors.New("test error")
	streaming := false

	err = logger.LogError(prompt, testErr, streaming)
	assert.NoError(t, err)

	// Read log file and verify content
	logData, err := os.ReadFile(logPath)
	assert.NoError(t, err)

	var entry LogEntry
	err = json.Unmarshal(logData, &entry)
	assert.NoError(t, err)

	assert.Equal(t, prompt, entry.Prompt)
	assert.Equal(t, testErr.Error(), entry.ErrorMessage)
	assert.Equal(t, streaming, entry.Streaming)
	assert.Equal(t, "stub", entry.LLMType)
	assert.False(t, entry.Success)
}

func TestLoggingService_Close(t *testing.T) {
	// Create temporary directory for test logs
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create logger
	logger, err := NewLoggingService(logPath, "stub")
	assert.NoError(t, err)

	// Test closing
	assert.NoError(t, logger.Close())

	// Test double close (should not error)
	assert.NoError(t, logger.Close())
}
