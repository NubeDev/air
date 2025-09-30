package routes

import (
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

func SetupChatAPIRoutes(router *gin.RouterGroup, aiService *services.AIService, reportsService *services.ReportsService, datasourceService *services.DatasourceService) {
	chatGroup := router.Group("/chat")
	{
		chatGroup.POST("/message", func(c *gin.Context) {
			// This endpoint is deprecated - use WebSocket for chat
			c.JSON(501, gin.H{
				"error": "This endpoint is deprecated. Please use WebSocket for chat functionality.",
			})
		})

		chatGroup.POST("/query-data", func(c *gin.Context) {
			// This endpoint is deprecated - use WebSocket for data queries
			c.JSON(501, gin.H{
				"error": "This endpoint is deprecated. Please use WebSocket for data query functionality.",
			})
		})

		chatGroup.POST("/create-report", func(c *gin.Context) {
			// This endpoint is deprecated - use WebSocket for report creation
			c.JSON(501, gin.H{
				"error": "This endpoint is deprecated. Please use WebSocket for report creation functionality.",
			})
		})

	}
}
