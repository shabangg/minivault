package main

import (
	"fmt"
	"log"
	"os"

	"minivault-api/src/api"
	"minivault-api/src/service"
)

// @title MiniVault API
// @version 1.0
// @description A lightweight local REST API that simulates ModelVault's prompt-response functionality.
// @host localhost:8080
// @BasePath /
func main() {
	// Get configuration from environment
	llmType := os.Getenv("LLM_TYPE")
	if llmType == "" {
		llmType = "ollama"
	}

	// Initialize services
	logger, err := service.NewLoggingService("logs/log.jsonl", llmType)
	if err != nil {
		log.Fatalf("Failed to initialize logging service: %v", err)
	}
	defer logger.Close()

	// Initialize generator service
	generator := service.NewGeneratorService(llmType)

	// Initialize handler
	handler := api.NewHandler(generator, logger)

	// Setup router
	router := api.SetupRouter(handler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Starting MiniVault API server on :%s...\n", port)
	fmt.Printf("Using LLM type: %s\n", llmType)

	fmt.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
