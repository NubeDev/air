package services

import (
	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/store"
)

// HealthService handles health check business logic
type HealthService struct {
	config   *config.Config
	registry *datasource.Registry
}

// NewHealthService creates a new health service
func NewHealthService(cfg *config.Config, registry *datasource.Registry) *HealthService {
	return &HealthService{
		config:   cfg,
		registry: registry,
	}
}

// GetHealthStatus returns the overall health status
func (s *HealthService) GetHealthStatus() store.HealthResponse {
	return store.HealthResponse{
		Status:      "healthy",
		AuthEnabled: s.config.Server.Auth.Enabled,
		Datasources: len(s.registry.ListDatasources()),
	}
}
