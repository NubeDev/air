package routes

import (
	"context"

	"github.com/NubeDev/air/cmd/api/handlers/websocket"
	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/redis"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupWebSocketRoutes sets up WebSocket routes
func SetupWebSocketRoutes(router *gin.Engine, redisClient *redis.Client, wsConfig *config.WebSocketConfig, aiService interface{}) {
	if !wsConfig.Enabled {
		logger.LogWarn(logger.ServiceWS, "WebSocket routes disabled")
		return
	}

	// Create WebSocket handler
	aiServiceTyped, ok := aiService.(*services.AIService)
	if !ok {
		logger.LogError(logger.ServiceWS, "Invalid AI service type", nil)
		return
	}
	wsHandler := websocket.NewHandler(redisClient, wsConfig, aiServiceTyped)

	// Start WebSocket hub
	ctx := context.Background()
	wsHandler.StartHub(ctx)

	// WebSocket group
	wsGroup := router.Group("/v1/ws")
	{
		// Main WebSocket endpoint
		wsGroup.GET("/", wsHandler.HandleWebSocket)

		// Chat-specific WebSocket endpoint
		wsGroup.GET("/chat", wsHandler.HandleChat)

		// Presence WebSocket endpoint
		wsGroup.GET("/presence", wsHandler.HandlePresence)
	}

	// WebSocket management endpoints
	wsAPI := router.Group("/v1/websocket")
	{
		// Get online users
		wsAPI.GET("/users", wsHandler.GetOnlineUsers)

		// Send message
		wsAPI.POST("/send", wsHandler.SendMessage)

		// Get hub statistics
		wsAPI.GET("/stats", wsHandler.GetHubStats)
	}

	logger.LogInfo(logger.ServiceWS, "WebSocket routes configured", map[string]interface{}{
		"enabled": wsConfig.Enabled,
		"endpoints": []string{
			"GET /v1/ws/",
			"GET /v1/ws/chat",
			"GET /v1/ws/presence",
			"GET /v1/websocket/users",
			"POST /v1/websocket/send",
			"GET /v1/websocket/stats",
		},
	})
}
