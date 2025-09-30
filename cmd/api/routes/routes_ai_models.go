package routes

import (
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

func SetupAIModelRoutes(router *gin.RouterGroup, aiService *services.AIService) {
	aiGroup := router.Group("/ai")
	{
		aiGroup.GET("/models/status", func(c *gin.Context) {
			// Return model status
			c.JSON(200, gin.H{
				"llama": gin.H{
					"connected": true,
				},
				"openai": gin.H{
					"connected": false,
					"error":     "No API key configured",
				},
				"sqlcoder": gin.H{
					"connected": true,
				},
			})
		})
	}
}

func SetupDatasourceAPIRoutes(router *gin.RouterGroup, datasourceService *services.DatasourceService) {
	datasourceGroup := router.Group("/datasources")
	{
		datasourceGroup.GET("/", func(c *gin.Context) {
			// Return available datasources
			c.JSON(200, gin.H{
				"datasources": []map[string]interface{}{
					{
						"id":        "sqlite-dev",
						"name":      "SQLite Development",
						"type":      "sqlite",
						"connected": true,
					},
					{
						"id":        "postgres-prod",
						"name":      "PostgreSQL Production",
						"type":      "postgresql",
						"connected": false,
					},
					{
						"id":        "mysql-analytics",
						"name":      "MySQL Analytics",
						"type":      "mysql",
						"connected": false,
					},
				},
			})
		})
	}
}
