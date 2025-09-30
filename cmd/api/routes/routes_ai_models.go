package routes

import (
	"context"
	"time"

	"github.com/NubeDev/air/internal/llm"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

func SetupAIModelRoutes(router *gin.RouterGroup, aiService *services.AIService) {
	aiGroup := router.Group("/ai")
	{
		aiGroup.GET("/models/status", func(c *gin.Context) {
			// Check model status dynamically
			status := make(map[string]interface{})

			// Check OpenAI
			if aiService.Config.Models.ChatPrimary == "openai" && aiService.Config.Models.OpenAI.APIKey != "" {
				// Try to create OpenAI client and check health
				openaiClient, err := llm.NewOpenAIClient(aiService.Config.Models.OpenAI)
				if err == nil {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					err = openaiClient.Health(ctx)
					cancel()
				}

				if err == nil {
					status["openai"] = gin.H{"connected": true}
				} else {
					status["openai"] = gin.H{
						"connected": false,
						"error":     err.Error(),
					}
				}
			} else {
				status["openai"] = gin.H{
					"connected": false,
					"error":     "No API key configured",
				}
			}

			// Check Ollama (always try to check)
			ollamaClient, err := llm.NewOllamaClient(aiService.Config.Models.Ollama)
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				err = ollamaClient.Health(ctx)
				cancel()
			}

			if err == nil {
				status["llama"] = gin.H{"connected": true}
			} else {
				status["llama"] = gin.H{
					"connected": false,
					"error":     err.Error(),
				}
			}

			// SQLCoder is typically the same as Ollama
			status["sqlcoder"] = status["llama"]

			c.JSON(200, status)
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
