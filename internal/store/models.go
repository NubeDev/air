package store

import (
	"time"

	"gorm.io/gorm"
)

// Datasource represents a registered analytics database connection
type Datasource struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Kind        string    `gorm:"not null" json:"kind"` // "postgres", "timescaledb", "mysql"
	DSN         string    `gorm:"not null" json:"dsn"`
	DisplayName string    `gorm:"not null" json:"display_name"`
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Scope represents a business question scope
type Scope struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Status    string    `gorm:"default:'draft'" json:"status"` // "draft", "approved", "archived"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ScopeVersion represents a versioned scope with Markdown content and IR
type ScopeVersion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ScopeID   uint      `gorm:"not null" json:"scope_id"`
	Version   int       `gorm:"not null" json:"version"`
	ScopeMD   string    `gorm:"type:text" json:"scope_md"`
	IRJSON    string    `gorm:"type:text" json:"ir_json"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Scope Scope `gorm:"foreignKey:ScopeID" json:"scope,omitempty"`
}

// SchemaNote represents learned schema information from a datasource
type SchemaNote struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DatasourceID string    `gorm:"not null" json:"datasource_id"`
	Object       string    `gorm:"not null" json:"object"`  // table name, view name, etc.
	Chunk        int       `gorm:"not null" json:"chunk"`   // chunk number for large schemas
	MD           string    `gorm:"type:text" json:"md"`     // markdown content
	MDHash       string    `gorm:"not null" json:"md_hash"` // hash for deduplication
	CreatedAt    time.Time `json:"created_at"`

	// Relationships
	Datasource Datasource `gorm:"foreignKey:DatasourceID" json:"datasource,omitempty"`
}

// Report represents a saved report definition
type Report struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"uniqueIndex;not null" json:"key"`
	Title     string    `gorm:"not null" json:"title"`
	Owner     string    `json:"owner"`
	Archived  bool      `gorm:"default:false" json:"archived"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ReportVersion represents a versioned report definition
type ReportVersion struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ReportID       uint      `gorm:"not null" json:"report_id"`
	Version        int       `gorm:"not null" json:"version"`
	ScopeVersionID uint      `gorm:"not null" json:"scope_version_id"`
	DatasourceID   *string   `json:"datasource_id"` // null for portable reports
	DefJSON        string    `gorm:"type:text" json:"def_json"`
	Checksum       string    `gorm:"not null" json:"checksum"`
	Status         string    `gorm:"default:'draft'" json:"status"` // "draft", "active", "archived"
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	Report       Report       `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	ScopeVersion ScopeVersion `gorm:"foreignKey:ScopeVersionID" json:"scope_version,omitempty"`
	Datasource   *Datasource  `gorm:"foreignKey:DatasourceID" json:"datasource,omitempty"`
}

// ReportRun represents an execution of a report
type ReportRun struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	ReportID        uint       `gorm:"not null" json:"report_id"`
	ReportVersionID uint       `gorm:"not null" json:"report_version_id"`
	DatasourceID    string     `gorm:"not null" json:"datasource_id"`
	ParamsJSON      string     `gorm:"type:text" json:"params_json"`
	SQLText         string     `gorm:"type:text" json:"sql_text"`
	RowCount        int        `json:"row_count"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	Status          string     `gorm:"default:'running'" json:"status"` // "running", "completed", "failed"
	ErrorText       string     `gorm:"type:text" json:"error_text"`

	// Relationships
	Report        Report        `gorm:"foreignKey:ReportID" json:"report,omitempty"`
	ReportVersion ReportVersion `gorm:"foreignKey:ReportVersionID" json:"report_version,omitempty"`
	Datasource    Datasource    `gorm:"foreignKey:DatasourceID" json:"datasource,omitempty"`
}

// ReportSample represents sample rows from a report run
type ReportSample struct {
	RunID   uint   `gorm:"primaryKey" json:"run_id"`
	Seq     int    `gorm:"primaryKey" json:"seq"`
	RowJSON string `gorm:"type:text" json:"row_json"`

	// Relationships
	Run ReportRun `gorm:"foreignKey:RunID" json:"run,omitempty"`
}

// ReportAnalysis represents AI analysis of a report run
type ReportAnalysis struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	RunID         uint      `gorm:"not null" json:"run_id"`
	ModelUsed     string    `gorm:"not null" json:"model_used"`
	RubricVersion string    `gorm:"not null" json:"rubric_version"`
	VerdictJSON   string    `gorm:"type:text" json:"verdict_json"`
	AnalysisMD    string    `gorm:"type:text" json:"analysis_md"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Run ReportRun `gorm:"foreignKey:RunID" json:"run,omitempty"`
}

// ============================================================================
// API Request/Response Models
// ============================================================================

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string `json:"status"`
	AuthEnabled bool   `json:"auth_enabled"`
	Datasources int    `json:"datasources"`
}

// DatasourceResponse represents a datasource in API responses
type DatasourceResponse struct {
	ID           string    `json:"id"`
	Kind         string    `json:"kind"`
	DisplayName  string    `json:"display_name"`
	IsDefault    bool      `json:"is_default"`
	HealthStatus string    `json:"health_status"`
	LastHealth   time.Time `json:"last_health"`
	Error        string    `json:"error,omitempty"`
}

// DatasourcesResponse represents the list datasources response
type DatasourcesResponse struct {
	Datasources []DatasourceResponse `json:"datasources"`
}

// HealthCheckResponse represents a datasource health check response
type HealthCheckResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// ============================================================================
// API Request Models
// ============================================================================

// CreateDatasourceRequest represents the request to create a new datasource
type CreateDatasourceRequest struct {
	ID          string `json:"id" binding:"required"`
	Kind        string `json:"kind" binding:"required,oneof=postgres timescaledb mysql"`
	DSN         string `json:"dsn" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	IsDefault   bool   `json:"is_default"`
}

// LearnDatasourceRequest represents the request to learn from a datasource
type LearnDatasourceRequest struct {
	DatasourceID string   `json:"datasource_id" binding:"required"`
	Schemas      []string `json:"schemas,omitempty"`
}

// CreateScopeRequest represents the request to create a new scope
type CreateScopeRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateScopeVersionRequest represents the request to create a new scope version
type CreateScopeVersionRequest struct {
	ScopeMD string `json:"scope_md" binding:"required"`
}

// BuildIRRequest represents the request to build IR from scope
type BuildIRRequest struct {
	ScopeVersionID uint `json:"scope_version_id" binding:"required"`
}

// GenerateSQLRequest represents the request to generate SQL
type GenerateSQLRequest struct {
	IR           map[string]interface{} `json:"ir" binding:"required"`
	DatasourceID string                 `json:"datasource_id" binding:"required"`
}

// CreateReportRequest represents the request to create a new report
type CreateReportRequest struct {
	Key   string `json:"key" binding:"required"`
	Title string `json:"title" binding:"required"`
	Owner string `json:"owner,omitempty"`
}

// CreateReportVersionRequest represents the request to create a new report version
type CreateReportVersionRequest struct {
	ScopeVersionID uint    `json:"scope_version_id" binding:"required"`
	DatasourceID   *string `json:"datasource_id,omitempty"`
	DefJSON        string  `json:"def_json" binding:"required"`
}

// RunReportRequest represents the request to run a report
type RunReportRequest struct {
	Params       map[string]interface{} `json:"params" binding:"required"`
	DatasourceID *string                `json:"datasource_id,omitempty"`
}

// AnalyzeRunRequest represents the request to analyze a report run
type AnalyzeRunRequest struct {
	ModelUsed     string `json:"model_used,omitempty"`
	RubricVersion string `json:"rubric_version,omitempty"`
}

// ============================================================================
// Database Migration
// ============================================================================

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Datasource{},
		&Scope{},
		&ScopeVersion{},
		&SchemaNote{},
		&Report{},
		&ReportVersion{},
		&ReportRun{},
		&ReportSample{},
		&ReportAnalysis{},
	)
}
