package routes

import (
	"fmt"

	"github.com/NubeDev/air/cmd/api/handlers/fastapi"
	"github.com/NubeDev/air/cmd/api/handlers/health"
	"github.com/NubeDev/air/internal/auth"
	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/redis"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, cfg *config.Config, db *gorm.DB, registry *datasource.Registry, jwtManager *auth.JWTManager, redisClient *redis.Client) {
	// Initialize services
	datasourceService := services.NewDatasourceService(registry, db)
	aiService, err := services.NewAIService(registry, db, cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize AI service: %v", err))
	}
	reportsService := services.NewReportsService(registry, db)
	healthService := services.NewHealthService(cfg, registry)
	fastapiHandler := fastapi.NewFastAPIHandler("http://localhost:9001")

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
		SetupChatRoutes(v1, aiService, authMiddleware)
		SetupSessionRoutes(v1, db, authMiddleware)
		SetupGeneratedReportRoutes(v1, db, authMiddleware)

		// FastAPI integration routes
		fastapiGroup := v1.Group("/fastapi")
		{
			fastapiGroup.GET("/health", fastapiHandler.Health)
			fastapiGroup.POST("/test/energy", fastapiHandler.TestEnergyData)
			fastapiGroup.POST("/test/discover", fastapiHandler.TestDiscoverFiles)
		}
	}

	// WebSocket routes
	if cfg.Server.WSEnabled {
		SetupWebSocketRoutes(router, redisClient, &cfg.WebSocket)
	}
}
