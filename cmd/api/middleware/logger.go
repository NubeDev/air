package middleware

import (
	"time"

	"github.com/NubeDev/air/internal/logger"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware creates a custom Gin logger middleware using zerolog
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log the request using structured logging
		logger.LogRequest(
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
		)

		// Return empty string since we're handling logging ourselves
		return ""
	})
}

// RecoveryMiddleware creates a custom recovery middleware with structured logging
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.LogError(logger.ServiceREST, "Panic recovered", nil, map[string]interface{}{
			"panic":     recovered,
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"client_ip": c.ClientIP(),
		})

		c.AbortWithStatus(500)
	})
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
