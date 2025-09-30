package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/health"
	"github.com/NubeDev/air/internal/auth"
	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, cfg *config.Config, db *gorm.DB, registry *datasource.Registry, jwtManager *auth.JWTManager) {
	// Initialize services
	datasourceService := services.NewDatasourceService(registry, db)
	aiService := services.NewAIService(registry, db)
	reportsService := services.NewReportsService(registry, db)
	healthService := services.NewHealthService(cfg, registry)

	// Health check endpoint
	router.GET("/health", health.HealthHandler(healthService))

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Authentication middleware
		var authMiddleware gin.HandlerFunc
		if cfg.Server.Auth.Enabled && jwtManager != nil {
			authMiddleware = auth.AuthMiddleware(jwtManager, true)
		} else {
			authMiddleware = func(c *gin.Context) { c.Next() }
		}

		// Setup API groups
		SetupDatasourceRoutes(v1, datasourceService, authMiddleware)
		SetupLearnRoutes(v1, datasourceService, authMiddleware)
		SetupSchemaRoutes(v1, datasourceService, authMiddleware)
		SetupScopeRoutes(v1, reportsService, authMiddleware)
		SetupIRRoutes(v1, aiService, authMiddleware)
		SetupSQLRoutes(v1, aiService, authMiddleware)
		SetupReportRoutes(v1, reportsService, authMiddleware)
		SetupAnalysisRoutes(v1, aiService, authMiddleware)
		SetupAIToolsRoutes(v1, aiService, authMiddleware)
	}

	// WebSocket endpoint
	if cfg.Server.WSEnabled {
		router.GET("/v1/ws", health.WebSocketHandler())
	}
}
