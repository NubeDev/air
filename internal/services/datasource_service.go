package services

import (
	"fmt"
	"time"

	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"gorm.io/gorm"
)

// DatasourceService handles datasource business logic
type DatasourceService struct {
	registry *datasource.Registry
	db       *gorm.DB
}

// NewDatasourceService creates a new datasource service
func NewDatasourceService(registry *datasource.Registry, db *gorm.DB) *DatasourceService {
	return &DatasourceService{
		registry: registry,
		db:       db,
	}
}

// ListDatasources returns all registered datasources with health status
func (s *DatasourceService) ListDatasources() ([]store.DatasourceResponse, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceDB, "Listing datasources")

	connectors := s.registry.ListDatasources()

	datasources := make([]store.DatasourceResponse, len(connectors))
	for i, connector := range connectors {
		datasources[i] = store.DatasourceResponse{
			ID:           connector.ID,
			Kind:         connector.Kind,
			DisplayName:  connector.DisplayName,
			IsDefault:    connector.IsDefault,
			HealthStatus: connector.HealthStatus,
			LastHealth:   connector.LastHealth,
		}
		if connector.Error != nil {
			datasources[i].Error = connector.Error.Error()
		}
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceDB, "Datasources listed successfully", map[string]interface{}{
		"count":    len(datasources),
		"duration": duration.String(),
	})

	return datasources, nil
}

// CreateDatasource creates a new datasource
func (s *DatasourceService) CreateDatasource(req store.CreateDatasourceRequest) error {
	start := time.Now()

	logger.LogInfo(logger.ServiceDB, "Creating datasource", map[string]interface{}{
		"id":           req.ID,
		"kind":         req.Kind,
		"display_name": req.DisplayName,
		"is_default":   req.IsDefault,
	})

	err := s.registry.AddDatasource(req.ID, req.Kind, req.DSN, req.DisplayName, req.IsDefault)

	duration := time.Since(start)
	if err != nil {
		logger.LogError(logger.ServiceDB, "Failed to create datasource", err, map[string]interface{}{
			"id":       req.ID,
			"duration": duration.String(),
		})
		return err
	}

	logger.LogInfo(logger.ServiceDB, "Datasource created successfully", map[string]interface{}{
		"id":       req.ID,
		"duration": duration.String(),
	})

	return nil
}

// GetDatasourceHealth checks the health of a specific datasource
func (s *DatasourceService) GetDatasourceHealth(id string) (store.HealthCheckResponse, error) {
	connector, err := s.registry.GetDatasource(id)
	if err != nil {
		return store.HealthCheckResponse{}, fmt.Errorf("datasource not found: %w", err)
	}

	if err := connector.TestConnection(); err != nil {
		return store.HealthCheckResponse{
			Status: "unhealthy",
			Error:  err.Error(),
		}, nil
	}

	return store.HealthCheckResponse{
		Status: "healthy",
	}, nil
}

// DeleteDatasource removes a datasource
func (s *DatasourceService) DeleteDatasource(id string) error {
	return s.registry.RemoveDatasource(id)
}

// LearnDatasource learns schema from a datasource
func (s *DatasourceService) LearnDatasource(req store.LearnDatasourceRequest) error {
	// TODO: Implement actual learning logic
	return fmt.Errorf("not implemented")
}

// GetSchema returns schema information for a datasource
func (s *DatasourceService) GetSchema(datasourceID string) ([]store.SchemaNote, error) {
	// TODO: Implement schema retrieval logic
	return nil, fmt.Errorf("not implemented")
}
