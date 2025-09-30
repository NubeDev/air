package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/store"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// Registry manages multiple datasource connections
type Registry struct {
	config      *config.Config
	db          *gorm.DB
	datasources map[string]*DatasourceConnector
	mu          sync.RWMutex
}

// DatasourceConnector represents a connection to a specific datasource
type DatasourceConnector struct {
	ID           string
	Kind         string
	DisplayName  string
	IsDefault    bool
	DB           *sql.DB
	LastHealth   time.Time
	HealthStatus string // "healthy", "unhealthy", "unknown"
	Error        error
}

// NewRegistry creates a new datasource registry
func NewRegistry(cfg *config.Config, db *gorm.DB) *Registry {
	registry := &Registry{
		config:      cfg,
		db:          db,
		datasources: make(map[string]*DatasourceConnector),
	}

	// Initialize datasources from config
	registry.initializeFromConfig()

	return registry
}

// initializeFromConfig initializes datasources from configuration
func (r *Registry) initializeFromConfig() {
	for _, sourceConfig := range r.config.AnalyticsSources {
		// Check if datasource already exists in database
		var existingDatasource store.Datasource
		result := r.db.Where("id = ?", sourceConfig.ID).First(&existingDatasource)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new datasource in database
			datasource := store.Datasource{
				ID:          sourceConfig.ID,
				Kind:        sourceConfig.Kind,
				DSN:         sourceConfig.DSN,
				DisplayName: sourceConfig.DisplayName,
				IsDefault:   sourceConfig.Default,
			}

			if err := r.db.Create(&datasource).Error; err != nil {
				fmt.Printf("Failed to create datasource %s: %v\n", sourceConfig.ID, err)
				continue
			}
		} else if result.Error != nil {
			fmt.Printf("Failed to check datasource %s: %v\n", sourceConfig.ID, result.Error)
			continue
		}

		// Create connector
		connector, err := r.createConnector(sourceConfig)
		if err != nil {
			fmt.Printf("Failed to create connector for %s: %v\n", sourceConfig.ID, err)
			continue
		}

		r.mu.Lock()
		r.datasources[sourceConfig.ID] = connector
		r.mu.Unlock()
	}
}

// createConnector creates a new datasource connector
func (r *Registry) createConnector(sourceConfig config.AnalyticsSourceConfig) (*DatasourceConnector, error) {
	db, err := r.openConnection(sourceConfig.Kind, sourceConfig.DSN)
	if err != nil {
		return &DatasourceConnector{
			ID:           sourceConfig.ID,
			Kind:         sourceConfig.Kind,
			DisplayName:  sourceConfig.DisplayName,
			IsDefault:    sourceConfig.Default,
			HealthStatus: "unhealthy",
			Error:        err,
		}, err
	}

	connector := &DatasourceConnector{
		ID:           sourceConfig.ID,
		Kind:         sourceConfig.Kind,
		DisplayName:  sourceConfig.DisplayName,
		IsDefault:    sourceConfig.Default,
		DB:           db,
		LastHealth:   time.Now(),
		HealthStatus: "healthy",
	}

	// Test connection
	if err := connector.TestConnection(); err != nil {
		connector.HealthStatus = "unhealthy"
		connector.Error = err
	}

	return connector, nil
}

// openConnection opens a database connection based on the kind
func (r *Registry) openConnection(kind, dsn string) (*sql.DB, error) {
	var driver string
	switch kind {
	case "postgres", "timescaledb":
		driver = "postgres"
	case "mysql":
		driver = "mysql"
	case "sqlite":
		driver = "sqlite3"
	default:
		return nil, fmt.Errorf("unsupported database kind: %s", kind)
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// TestConnection tests the datasource connection
func (c *DatasourceConnector) TestConnection() error {
	if c.DB == nil {
		return fmt.Errorf("no database connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.DB.PingContext(ctx)
}

// GetDatasource returns a datasource connector by ID
func (r *Registry) GetDatasource(id string) (*DatasourceConnector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connector, exists := r.datasources[id]
	if !exists {
		return nil, fmt.Errorf("datasource not found: %s", id)
	}

	return connector, nil
}

// GetDefaultDatasource returns the default datasource
func (r *Registry) GetDefaultDatasource() (*DatasourceConnector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, connector := range r.datasources {
		if connector.IsDefault {
			return connector, nil
		}
	}

	return nil, fmt.Errorf("no default datasource found")
}

// ListDatasources returns all registered datasources
func (r *Registry) ListDatasources() []*DatasourceConnector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	connectors := make([]*DatasourceConnector, 0, len(r.datasources))
	for _, connector := range r.datasources {
		connectors = append(connectors, connector)
	}

	return connectors
}

// AddDatasource adds a new datasource to the registry
func (r *Registry) AddDatasource(id, kind, dsn, displayName string, isDefault bool) error {
	// Create in database
	datasource := store.Datasource{
		ID:          id,
		Kind:        kind,
		DSN:         dsn,
		DisplayName: displayName,
		IsDefault:   isDefault,
	}

	if err := r.db.Create(&datasource).Error; err != nil {
		return fmt.Errorf("failed to create datasource in database: %w", err)
	}

	// Create connector
	connector, err := r.createConnector(config.AnalyticsSourceConfig{
		ID:          id,
		Kind:        kind,
		DSN:         dsn,
		DisplayName: displayName,
		Default:     isDefault,
	})
	if err != nil {
		// Clean up database record
		r.db.Delete(&datasource)
		return fmt.Errorf("failed to create connector: %w", err)
	}

	r.mu.Lock()
	r.datasources[id] = connector
	r.mu.Unlock()

	return nil
}

// RemoveDatasource removes a datasource from the registry
func (r *Registry) RemoveDatasource(id string) error {
	// Check if datasource is in use
	var count int64
	r.db.Model(&store.ReportRun{}).Where("datasource_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("cannot remove datasource %s: it is in use by %d report runs", id, count)
	}

	// Close connection
	r.mu.Lock()
	if connector, exists := r.datasources[id]; exists {
		if connector.DB != nil {
			connector.DB.Close()
		}
		delete(r.datasources, id)
	}
	r.mu.Unlock()

	// Remove from database
	if err := r.db.Where("id = ?", id).Delete(&store.Datasource{}).Error; err != nil {
		return fmt.Errorf("failed to remove datasource from database: %w", err)
	}

	return nil
}

// HealthCheck performs health checks on all datasources
func (r *Registry) HealthCheck() map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()

	results := make(map[string]string)

	for id, connector := range r.datasources {
		if err := connector.TestConnection(); err != nil {
			connector.HealthStatus = "unhealthy"
			connector.Error = err
			results[id] = fmt.Sprintf("unhealthy: %v", err)
		} else {
			connector.HealthStatus = "healthy"
			connector.Error = nil
			connector.LastHealth = time.Now()
			results[id] = "healthy"
		}
	}

	return results
}

// Close closes all datasource connections
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for _, connector := range r.datasources {
		if connector.DB != nil {
			if err := connector.DB.Close(); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}
