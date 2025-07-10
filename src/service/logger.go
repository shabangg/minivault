package service

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

// Logger defines the interface for logging operations
type Logger interface {
	LogInteraction(prompt, response string, streaming bool) error
	LogError(prompt string, err error, streaming bool) error
	Close() error
}

// LogEntry represents a single log entry with enhanced details
type LogEntry struct {
	// Request details
	ID        string    `json:"id"`          // Unique request ID
	Timestamp time.Time `json:"timestamp"`   // ISO 8601 timestamp
	Duration  int64     `json:"duration_ms"` // Request duration in milliseconds

	// Input details
	Prompt    string `json:"prompt"`
	LLMType   string `json:"llm_type"`  // "ollama" or "stub"
	LLMModel  string `json:"llm_model"` // Model name if using Ollama
	Streaming bool   `json:"streaming"` // Whether streaming was used

	// Response details
	Response     string `json:"response"`
	TokenCount   int    `json:"token_count"`   // Number of tokens in response
	ResponseSize int    `json:"response_size"` // Size of response in bytes

	// Status details
	Success      bool   `json:"success"`         // Whether the request succeeded
	ErrorMessage string `json:"error,omitempty"` // Error message if any

	// System details
	GoVersion  string `json:"go_version"`   // Go runtime version
	GoRoutines int    `json:"goroutines"`   // Number of active goroutines
	MemoryUsed int64  `json:"memory_bytes"` // Memory used in bytes
}

// LoggingService handles logging of interactions
type LoggingService struct {
	logFile *os.File
	llmType string
}

// NewLoggingService creates a new logging service
func NewLoggingService(logPath, llmType string) (*LoggingService, error) {
	// Create logs directory if it doesn't exist
	dir := "logs"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %v", err)
	}

	// Open log file
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	return &LoggingService{
		logFile: logFile,
		llmType: llmType,
	}, nil
}

// Close closes the log file
func (s *LoggingService) Close() error {
	if s.logFile == nil {
		return nil
	}
	err := s.logFile.Close()
	if err == nil {
		s.logFile = nil
	}
	return err
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
}

// getSystemStats returns current system statistics
func getSystemStats() (int, int64) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return runtime.NumGoroutine(), int64(memStats.Alloc)
}

// countTokens returns a simple approximation of token count
// In a real implementation, this would use a proper tokenizer
func countTokens(text string) int {
	// Simple word-based approximation
	if text == "" {
		return 0
	}
	words := 0
	inWord := false
	for _, r := range text {
		if r == ' ' || r == '\n' || r == '\t' {
			inWord = false
		} else if !inWord {
			words++
			inWord = true
		}
	}
	return words
}

// LogInteraction logs a prompt-response interaction with enhanced details
func (s *LoggingService) LogInteraction(prompt, response string, streaming bool) error {
	startTime := time.Now()
	goroutines, memUsed := getSystemStats()

	entry := LogEntry{
		// Request details
		ID:        generateRequestID(),
		Timestamp: startTime,
		Duration:  time.Since(startTime).Milliseconds(),

		// Input details
		Prompt:    prompt,
		LLMType:   s.llmType,
		Streaming: streaming,

		// Response details
		Response:     response,
		TokenCount:   countTokens(response),
		ResponseSize: len(response),

		// Status details
		Success:      true, // Set to false if there was an error
		ErrorMessage: "",   // Populated when there's an error

		// System details
		GoVersion:  runtime.Version(),
		GoRoutines: goroutines,
		MemoryUsed: memUsed,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %v", err)
	}

	if _, err := fmt.Fprintln(s.logFile, string(jsonData)); err != nil {
		return fmt.Errorf("failed to write to log file: %v", err)
	}

	return nil
}

// LogError logs an error with the interaction
func (s *LoggingService) LogError(prompt string, err error, streaming bool) error {
	startTime := time.Now()
	goroutines, memUsed := getSystemStats()

	entry := LogEntry{
		// Request details
		ID:        generateRequestID(),
		Timestamp: startTime,
		Duration:  time.Since(startTime).Milliseconds(),

		// Input details
		Prompt:    prompt,
		LLMType:   s.llmType,
		Streaming: streaming,

		// Response details
		Response:     "",
		TokenCount:   0,
		ResponseSize: 0,

		// Status details
		Success:      false,
		ErrorMessage: err.Error(),

		// System details
		GoVersion:  runtime.Version(),
		GoRoutines: goroutines,
		MemoryUsed: memUsed,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal error log entry: %v", err)
	}

	if _, err := fmt.Fprintln(s.logFile, string(jsonData)); err != nil {
		return fmt.Errorf("failed to write error log entry: %v", err)
	}

	return nil
}
