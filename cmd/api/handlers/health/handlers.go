package health

import (
	"net/http"

	"github.com/NubeDev/air/internal/services"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
func HealthHandler(service *services.HealthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		response := service.GetHealthStatus()
		c.JSON(http.StatusOK, response)
	}
}

// WebSocketHandler handles WebSocket connections
func WebSocketHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement WebSocket handler
		c.JSON(http.StatusNotImplemented, store.ErrorResponse{
			Error: "Not implemented",
		})
	}
}
