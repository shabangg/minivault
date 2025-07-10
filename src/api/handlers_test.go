package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"minivault/src/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGenerator mocks the Generator interface
type MockGenerator struct {
	mock.Mock
}

func (m *MockGenerator) Generate(ctx context.Context, prompt string) (string, error) {
	args := m.Called(ctx, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockGenerator) GenerateStream(ctx context.Context, prompt string, writer io.Writer) error {
	args := m.Called(ctx, prompt, writer)
	return args.Error(0)
}

// MockLogger mocks the LoggingService
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) LogInteraction(prompt, response string, streaming bool) error {
	args := m.Called(prompt, response, streaming)
	return args.Error(0)
}

func (m *MockLogger) LogError(prompt string, err error, streaming bool) error {
	args := m.Called(prompt, err, streaming)
	return args.Error(0)
}

func (m *MockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestHandler() (*Handler, *MockGenerator, *MockLogger) {
	gin.SetMode(gin.TestMode)
	mockGen := new(MockGenerator)
	mockLogger := new(MockLogger)
	handler := NewHandler(mockGen, mockLogger)
	return handler, mockGen, mockLogger
}

func TestHandleGenerate_Success(t *testing.T) {
	handler, mockGen, mockLogger := setupTestHandler()

	// Setup expectations
	expectedPrompt := "test prompt"
	expectedResponse := "test response"
	mockGen.On("Generate", mock.Anything, expectedPrompt).Return(expectedResponse, nil)
	mockLogger.On("LogInteraction", expectedPrompt, expectedResponse, false).Return(nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := types.Request{Prompt: expectedPrompt}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/generate", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute handler
	handler.HandleGenerate(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	var response types.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response.Response)

	// Verify mocks
	mockGen.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandleGenerate_EmptyPrompt(t *testing.T) {
	handler, _, mockLogger := setupTestHandler()

	// Setup expectations
	mockLogger.On("LogError", "", mock.Anything, false).Return(nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := types.Request{Prompt: ""}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/generate", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute handler
	handler.HandleGenerate(c)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request format")

	// Verify mocks
	mockLogger.AssertExpectations(t)
}

func TestHandleGenerate_GeneratorError(t *testing.T) {
	handler, mockGen, mockLogger := setupTestHandler()

	// Setup expectations
	expectedPrompt := "test prompt"
	expectedError := errors.New("generator error")
	mockGen.On("Generate", mock.Anything, expectedPrompt).Return("", expectedError)
	mockLogger.On("LogError", expectedPrompt, expectedError, false).Return(nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := types.Request{Prompt: expectedPrompt}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/generate", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute handler
	handler.HandleGenerate(c)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to generate response")

	// Verify mocks
	mockGen.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandleGenerateStream_Success(t *testing.T) {
	handler, mockGen, mockLogger := setupTestHandler()

	// Setup expectations
	expectedPrompt := "test prompt"
	mockGen.On("GenerateStream", mock.Anything, expectedPrompt, mock.Anything).Return(nil)
	mockLogger.On("LogInteraction", expectedPrompt, mock.Anything, true).Return(nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := types.Request{Prompt: expectedPrompt}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/generate/stream", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute handler
	handler.HandleGenerateStream(c)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify mocks
	mockGen.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandleGenerateStream_Error(t *testing.T) {
	handler, mockGen, mockLogger := setupTestHandler()

	// Setup expectations
	expectedPrompt := "test prompt"
	expectedError := errors.New("stream error")
	mockGen.On("GenerateStream", mock.Anything, expectedPrompt, mock.Anything).Return(expectedError)
	mockLogger.On("LogError", expectedPrompt, expectedError, true).Return(nil)

	// Create test request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := types.Request{Prompt: expectedPrompt}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/generate/stream", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute handler
	handler.HandleGenerateStream(c)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Failed to generate response")

	// Verify mocks
	mockGen.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
