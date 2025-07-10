package api

import (
	_ "minivault-api/docs" // This is required for swagger

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter sets up the Gin router with all routes and middleware
func SetupRouter(handler *Handler) *gin.Engine {
	// Set Gin to release mode
	gin.SetMode(gin.ReleaseMode)

	// Initialize router
	router := gin.Default()

	// Register routes
	router.POST("/generate", handler.HandleGenerate)
	router.POST("/generate/stream", handler.HandleGenerateStream)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
