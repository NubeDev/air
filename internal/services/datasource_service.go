package services

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
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
	// Get datasource connector
	connector, err := s.registry.GetDatasource(req.DatasourceID)
	if err != nil {
		return fmt.Errorf("datasource not found: %w", err)
	}

	// Get DSN from database
	var datasource store.Datasource
	if err := s.db.Where("id = ?", req.DatasourceID).First(&datasource).Error; err != nil {
		return fmt.Errorf("failed to get datasource DSN: %w", err)
	}

	// Map connector kind to driver name
	driverName := connector.Kind
	if connector.Kind == "sqlite" {
		driverName = "sqlite3"
	}

	// Connect to the datasource
	db, err := sql.Open(driverName, datasource.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to datasource: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping datasource: %w", err)
	}

	// Introspect tables and views
	schemaNotes, err := s.introspectSchema(db, req.DatasourceID, connector.Kind, req.Schemas)
	if err != nil {
		return fmt.Errorf("failed to introspect schema: %w", err)
	}

	// Store schema notes in database
	for _, note := range schemaNotes {
		if err := s.db.Create(&note).Error; err != nil {
			// Log error but continue with other notes
			logger.LogError(logger.ServiceDB, "Failed to store schema note", err)
		}
	}

	return nil
}

// GetSchema returns schema information for a datasource
func (s *DatasourceService) GetSchema(datasourceID string) ([]store.SchemaNote, error) {
	var schemaNotes []store.SchemaNote
	if err := s.db.Where("datasource_id = ?", datasourceID).Find(&schemaNotes).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve schema notes: %w", err)
	}
	return schemaNotes, nil
}

// introspectSchema introspects database schema and returns schema notes
func (s *DatasourceService) introspectSchema(db *sql.DB, datasourceID, dbKind string, schemas []string) ([]store.SchemaNote, error) {
	var schemaNotes []store.SchemaNote

	// Get list of tables and views
	tables, err := s.getTablesAndViews(db, dbKind, schemas)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables and views: %w", err)
	}

	// Introspect each table/view
	for _, table := range tables {
		columns, err := s.getTableColumns(db, dbKind, table)
		if err != nil {
			logger.LogError(logger.ServiceDB, "Failed to get columns for table", err)
			continue
		}

		// Generate markdown description
		md := s.generateTableMarkdown(table, columns)
		mdHash := fmt.Sprintf("%x", md5.Sum([]byte(md)))

		// Create schema note
		note := store.SchemaNote{
			DatasourceID: datasourceID,
			Object:       table,
			Chunk:        0,
			MD:           md,
			MDHash:       mdHash,
			CreatedAt:    time.Now(),
		}

		schemaNotes = append(schemaNotes, note)
	}

	return schemaNotes, nil
}

// getTablesAndViews returns list of tables and views in the database
func (s *DatasourceService) getTablesAndViews(db *sql.DB, dbKind string, schemas []string) ([]string, error) {
	var query string
	var args []interface{}

	switch strings.ToLower(dbKind) {
	case "sqlite", "sqlite3":
		query = "SELECT name FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%'"
	case "postgres", "postgresql", "timescaledb":
		if len(schemas) == 0 {
			schemas = []string{"public"}
		}
		placeholders := strings.Repeat("?,", len(schemas))
		placeholders = placeholders[:len(placeholders)-1]
		query = fmt.Sprintf("SELECT tablename FROM pg_tables WHERE schemaname IN (%s) UNION SELECT viewname FROM pg_views WHERE schemaname IN (%s)", placeholders, placeholders)
		for _, schema := range schemas {
			args = append(args, schema, schema)
		}
	case "mysql":
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type IN ('BASE TABLE', 'VIEW')"
	default:
		return nil, fmt.Errorf("unsupported database kind: %s", dbKind)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// getTableColumns returns column information for a table
func (s *DatasourceService) getTableColumns(db *sql.DB, dbKind, tableName string) ([]ColumnInfo, error) {
	var query string
	var args []interface{}

	switch strings.ToLower(dbKind) {
	case "sqlite", "sqlite3":
		query = "PRAGMA table_info(?)"
		args = []interface{}{tableName}
	case "postgres", "postgresql", "timescaledb":
		query = `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_name = $1 
			ORDER BY ordinal_position`
		args = []interface{}{tableName}
	case "mysql":
		query = `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns 
			WHERE table_name = ? AND table_schema = DATABASE()
			ORDER BY ordinal_position`
		args = []interface{}{tableName}
	default:
		return nil, fmt.Errorf("unsupported database kind: %s", dbKind)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		if err := rows.Scan(&col.Name, &col.Type, &col.Nullable, &col.Default); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}
		columns = append(columns, col)
	}

	return columns, nil
}

// generateTableMarkdown generates markdown description of a table
func (s *DatasourceService) generateTableMarkdown(tableName string, columns []ColumnInfo) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# Table: %s\n\n", tableName))
	md.WriteString(fmt.Sprintf("**Columns:** %d\n\n", len(columns)))

	md.WriteString("| Column | Type | Nullable | Default |\n")
	md.WriteString("|--------|------|----------|--------|\n")

	for _, col := range columns {
		nullable := "No"
		if col.Nullable == "YES" || col.Nullable == "1" {
			nullable = "Yes"
		}

		defaultVal := col.Default
		if defaultVal == "" {
			defaultVal = "-"
		}

		md.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			col.Name, col.Type, nullable, defaultVal))
	}

	return md.String()
}

// ColumnInfo represents column information
type ColumnInfo struct {
	Name     string
	Type     string
	Nullable string
	Default  string
}
