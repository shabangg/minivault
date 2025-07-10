package api

import (
	"fmt"
	"minivault-api/src/service"
	"minivault-api/src/types"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests
type Handler struct {
	generator service.Generator
	logger    service.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(generator service.Generator, logger service.Logger) *Handler {
	return &Handler{
		generator: generator,
		logger:    logger,
	}
}

// @Summary Generate text
// @Description Generate text from a prompt
// @Tags generation
// @Accept json
// @Produce json
// @Param request body types.Request true "Prompt for text generation"
// @Success 200 {object} types.Response
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /generate [post]
func (h *Handler) HandleGenerate(c *gin.Context) {
	var req types.Request
	if err := c.BindJSON(&req); err != nil {
		h.logger.LogError(req.Prompt, err, false)
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Prompt == "" {
		err := fmt.Errorf("prompt cannot be empty")
		h.logger.LogError(req.Prompt, err, false)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Generate response
	responseText, err := h.generator.Generate(c.Request.Context(), req.Prompt)
	if err != nil {
		h.logger.LogError(req.Prompt, err, false)
		c.JSON(500, gin.H{"error": "Failed to generate response"})
		return
	}

	// Log the interaction
	if err := h.logger.LogInteraction(req.Prompt, responseText, false); err != nil {
		// Don't fail the request if logging fails
		c.JSON(200, types.Response{Response: responseText})
		return
	}

	// Return response
	c.JSON(200, types.Response{Response: responseText})
}

// @Summary Generate text with streaming
// @Description Generate text from a prompt with streaming response
// @Tags generation
// @Accept json
// @Produce json
// @Param request body types.Request true "Prompt for text generation"
// @Success 200 {string} string "Streamed response as newline-delimited JSON"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /generate/stream [post]
func (h *Handler) HandleGenerateStream(c *gin.Context) {
	var req types.Request
	if err := c.BindJSON(&req); err != nil {
		h.logger.LogError(req.Prompt, err, true)
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Prompt == "" {
		err := fmt.Errorf("prompt cannot be empty")
		h.logger.LogError(req.Prompt, err, true)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Create a channel to capture the full response for logging
	fullResponse := make(chan string, 1)
	responseBuilder := ""

	// Create chunked writer
	writer := service.NewChunkedWriter(c.Writer, func(text string) {
		responseBuilder += text
	})

	// Stream the response
	if err := h.generator.GenerateStream(c.Request.Context(), req.Prompt, writer); err != nil {
		h.logger.LogError(req.Prompt, err, true)
		c.JSON(500, gin.H{"error": "Failed to generate response"})
		return
	}

	// Log the complete interaction
	if err := h.logger.LogInteraction(req.Prompt, responseBuilder, true); err != nil {
		// Don't fail the request if logging fails
		return
	}

	fullResponse <- responseBuilder
}
