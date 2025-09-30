package main

import (
	"fmt"
	"os"
	"time"

	"github.com/NubeDev/air/cmd/api/middleware"
	"github.com/NubeDev/air/cmd/api/routes"
	"github.com/NubeDev/air/internal/auth"
	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Server represents the AIR API server
type Server struct {
	config   *config.Config
	db       *gorm.DB
	registry *datasource.Registry
	jwtMgr   *auth.JWTManager
	router   *gin.Engine
}

// NewServer creates a new server instance
func NewServer(cfg *config.Config) (*Server, error) {
	// Setup logging
	loggerConfig := &logger.LoggerConfig{
		Level:      cfg.Telemetry.Level,
		Format:     cfg.Telemetry.Format,
		TimeFormat: cfg.Telemetry.TimeFormat,
		Color:      cfg.Telemetry.Color,
	}
	logger.SetupLogger(loggerConfig)

	logger.LogInfo(logger.ServiceServer, "Initializing AIR server")

	// Initialize database
	logger.LogInfo(logger.ServiceDB, "Initializing database connection")

	db, err := initDatabase(cfg)
	if err != nil {
		logger.LogError(logger.ServiceDB, "Failed to initialize database", err)
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	logger.LogInfo(logger.ServiceDB, "Database connected successfully", map[string]interface{}{
		"dsn": cfg.ControlPlane.DSN,
	})

	// Initialize datasource registry
	logger.LogInfo(logger.ServiceConfig, "Initializing datasource registry", map[string]interface{}{
		"sources": len(cfg.AnalyticsSources),
	})

	registry := datasource.NewRegistry(cfg, db)
	logger.LogInfo(logger.ServiceConfig, "Datasource registry initialized")

	// Initialize JWT manager
	var jwtManager *auth.JWTManager
	if cfg.Server.Auth.Enabled {
		logger.LogInfo(logger.ServiceAuth, "Initializing JWT manager", map[string]interface{}{
			"token_expiry": cfg.Server.Auth.TokenExpiry.String(),
		})
		jwtManager = auth.NewJWTManager(cfg.Server.Auth.JWTSecret, cfg.Server.Auth.TokenExpiry)
		logger.LogInfo(logger.ServiceAuth, "JWT manager initialized")
	} else {
		logger.LogWarn(logger.ServiceAuth, "Authentication disabled")
	}

	// Setup router
	logger.LogInfo(logger.ServiceREST, "Setting up HTTP router")

	router := setupRouter(cfg, db, registry, jwtManager)
	logger.LogInfo(logger.ServiceREST, "HTTP router setup complete")

	logger.LogInfo(logger.ServiceServer, "AIR server initialization complete")
	return &Server{
		config:   cfg,
		db:       db,
		registry: registry,
		jwtMgr:   jwtManager,
		router:   router,
	}, nil
}

// Start starts the server
func (s *Server) Start() error {
	addr := s.config.GetServerAddr()

	logger.LogInfo(logger.ServiceServer, "Starting AIR server", map[string]interface{}{
		"address":           addr,
		"auth_enabled":      s.config.Server.Auth.Enabled,
		"websocket_enabled": s.config.Server.WSEnabled,
	})

	return s.router.Run(addr)
}

// Close closes the server and cleans up resources
func (s *Server) Close() error {
	return s.registry.Close()
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	start := time.Now()

	// Open SQLite database
	logger.LogInfo(logger.ServiceDB, "Connecting to database", map[string]interface{}{
		"dsn": cfg.ControlPlane.DSN,
	})
	db, err := gorm.Open(sqlite.Open(cfg.ControlPlane.DSN), &gorm.Config{})
	if err != nil {
		logger.LogError(logger.ServiceDB, "Failed to connect to database", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate schema
	logger.LogInfo(logger.ServiceDB, "Running database migrations")
	if err := store.AutoMigrate(db); err != nil {
		logger.LogError(logger.ServiceDB, "Failed to migrate database", err)
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceDB, "Database initialization complete", map[string]interface{}{
		"duration": duration.String(),
	})

	return db, nil
}

func setupRouter(cfg *config.Config, db *gorm.DB, registry *datasource.Registry, jwtManager *auth.JWTManager) *gin.Engine {
	// Set Gin mode
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Use custom middleware with structured logging
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())

	// Setup middleware
	setupMiddleware(router)

	// Setup routes
	routes.SetupRoutes(router, cfg, db, registry, jwtManager)

	return router
}
