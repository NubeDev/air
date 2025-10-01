package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupSessionRoutes configures session management routes
func SetupSessionRoutes(rg *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	sessionGroup := rg.Group("/sessions")
	sessionGroup.Use(authMiddleware)
	{
		sessionGroup.POST("/start", sessions.StartSession(db))
		sessionGroup.GET("", sessions.ListSessions(db))
		sessionGroup.GET("/:id", sessions.GetSession(db))
		sessionGroup.GET("/:id/status", sessions.GetSessionStatus(db))
		sessionGroup.DELETE("/:id", sessions.EndSession(db))

		// Scope management for sessions
		sessionGroup.GET("/:id/scope/versions", sessions.GetSessionScopeVersions(db))
		sessionGroup.POST("/:id/scope/versions", sessions.CreateSessionScopeVersion(db))
		sessionGroup.GET("/:id/scope/current", sessions.GetCurrentSessionScope(db))
		sessionGroup.POST("/:id/scope/current", sessions.UpdateCurrentSessionScope(db))
	}
}
