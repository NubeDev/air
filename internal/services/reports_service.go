package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"gorm.io/gorm"
)

// ReportsService handles report-related business logic
type ReportsService struct {
	registry *datasource.Registry
	db       *gorm.DB
}

// NewReportsService creates a new reports service
func NewReportsService(registry *datasource.Registry, db *gorm.DB) *ReportsService {
	return &ReportsService{
		registry: registry,
		db:       db,
	}
}

// CreateScope creates a new scope
func (s *ReportsService) CreateScope(req store.CreateScopeRequest) (*store.Scope, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceREST, "Creating scope", map[string]interface{}{
		"name": req.Name,
	})

	// Create scope
	scope := &store.Scope{
		Name:      req.Name,
		Status:    "draft",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(scope).Error; err != nil {
		logger.LogError(logger.ServiceREST, "Failed to create scope", err, map[string]interface{}{
			"name": req.Name,
		})
		return nil, fmt.Errorf("failed to create scope: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceREST, "Scope created successfully", map[string]interface{}{
		"scope_id": scope.ID,
		"name":     scope.Name,
		"duration": duration.String(),
	})

	return scope, nil
}

// GetScope retrieves a scope by ID
func (s *ReportsService) GetScope(id uint) (*store.Scope, error) {
	logger.LogInfo(logger.ServiceREST, "Retrieving scope", map[string]interface{}{
		"scope_id": id,
	})

	var scope store.Scope
	if err := s.db.First(&scope, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.LogWarn(logger.ServiceREST, "Scope not found", map[string]interface{}{
				"scope_id": id,
			})
			return nil, fmt.Errorf("scope not found")
		}
		logger.LogError(logger.ServiceREST, "Failed to retrieve scope", err, map[string]interface{}{
			"scope_id": id,
		})
		return nil, fmt.Errorf("failed to retrieve scope: %w", err)
	}

	logger.LogInfo(logger.ServiceREST, "Scope retrieved successfully", map[string]interface{}{
		"scope_id": scope.ID,
		"name":     scope.Name,
		"status":   scope.Status,
	})

	return &scope, nil
}

// CreateScopeVersion creates a new version of a scope
func (s *ReportsService) CreateScopeVersion(scopeID uint, req store.CreateScopeVersionRequest) (*store.ScopeVersion, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceREST, "Creating scope version", map[string]interface{}{
		"scope_id": scopeID,
	})

	// Check if scope exists
	var scope store.Scope
	if err := s.db.First(&scope, scopeID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("scope not found")
		}
		return nil, fmt.Errorf("failed to find scope: %w", err)
	}

	// Get next version number
	var maxVersion int
	if err := s.db.Model(&store.ScopeVersion{}).
		Where("scope_id = ?", scopeID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to get max version: %w", err)
	}

	// Create scope version
	scopeVersion := &store.ScopeVersion{
		ScopeID:   scopeID,
		Version:   maxVersion + 1,
		ScopeMD:   req.ScopeMD,
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(scopeVersion).Error; err != nil {
		logger.LogError(logger.ServiceREST, "Failed to create scope version", err, map[string]interface{}{
			"scope_id": scopeID,
			"version":  scopeVersion.Version,
		})
		return nil, fmt.Errorf("failed to create scope version: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceREST, "Scope version created successfully", map[string]interface{}{
		"scope_id": scopeID,
		"version":  scopeVersion.Version,
		"duration": duration.String(),
	})

	return scopeVersion, nil
}

// CreateReport creates a new report
func (s *ReportsService) CreateReport(req store.CreateReportRequest) (*store.Report, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceREST, "Creating report", map[string]interface{}{
		"key":   req.Key,
		"title": req.Title,
		"owner": req.Owner,
	})

	// Check if report key already exists
	var existingReport store.Report
	if err := s.db.Where("key = ?", req.Key).First(&existingReport).Error; err == nil {
		logger.LogWarn(logger.ServiceREST, "Report key already exists", map[string]interface{}{
			"key": req.Key,
		})
		return nil, fmt.Errorf("report key already exists")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing report: %w", err)
	}

	// Create report
	report := &store.Report{
		Key:       req.Key,
		Title:     req.Title,
		Owner:     req.Owner,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.db.Create(report).Error; err != nil {
		logger.LogError(logger.ServiceREST, "Failed to create report", err, map[string]interface{}{
			"key":   req.Key,
			"title": req.Title,
		})
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceREST, "Report created successfully", map[string]interface{}{
		"report_id": report.ID,
		"key":       report.Key,
		"title":     report.Title,
		"duration":  duration.String(),
	})

	return report, nil
}

// GetReport retrieves a report by key
func (s *ReportsService) GetReport(key string) (*store.Report, error) {
	logger.LogInfo(logger.ServiceREST, "Retrieving report", map[string]interface{}{
		"key": key,
	})

	var report store.Report
	if err := s.db.Where("key = ?", key).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.LogWarn(logger.ServiceREST, "Report not found", map[string]interface{}{
				"key": key,
			})
			return nil, fmt.Errorf("report not found")
		}
		logger.LogError(logger.ServiceREST, "Failed to retrieve report", err, map[string]interface{}{
			"key": key,
		})
		return nil, fmt.Errorf("failed to retrieve report: %w", err)
	}

	logger.LogInfo(logger.ServiceREST, "Report retrieved successfully", map[string]interface{}{
		"report_id": report.ID,
		"key":       report.Key,
		"title":     report.Title,
	})

	return &report, nil
}

// CreateReportVersion creates a new version of a report
func (s *ReportsService) CreateReportVersion(reportKey string, req store.CreateReportVersionRequest) (*store.ReportVersion, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceREST, "Creating report version", map[string]interface{}{
		"report_key":       reportKey,
		"scope_version_id": req.ScopeVersionID,
		"datasource_id":    req.DatasourceID,
	})

	// Check if report exists
	var report store.Report
	if err := s.db.Where("key = ?", reportKey).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		return nil, fmt.Errorf("failed to find report: %w", err)
	}

	// Check if scope version exists
	var scopeVersion store.ScopeVersion
	if err := s.db.First(&scopeVersion, req.ScopeVersionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("scope version not found")
		}
		return nil, fmt.Errorf("failed to find scope version: %w", err)
	}

	// Get next version number
	var maxVersion int
	if err := s.db.Model(&store.ReportVersion{}).
		Where("report_id = ?", report.ID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error; err != nil {
		return nil, fmt.Errorf("failed to get max version: %w", err)
	}

	// Create report version
	reportVersion := &store.ReportVersion{
		ReportID:       report.ID,
		ScopeVersionID: req.ScopeVersionID,
		DatasourceID:   req.DatasourceID,
		Version:        maxVersion + 1,
		DefJSON:        req.DefJSON,
		CreatedAt:      time.Now(),
	}

	if err := s.db.Create(reportVersion).Error; err != nil {
		logger.LogError(logger.ServiceREST, "Failed to create report version", err, map[string]interface{}{
			"report_id": report.ID,
			"version":   reportVersion.Version,
		})
		return nil, fmt.Errorf("failed to create report version: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceREST, "Report version created successfully", map[string]interface{}{
		"report_id": report.ID,
		"version":   reportVersion.Version,
		"duration":  duration.String(),
	})

	return reportVersion, nil
}

// RunReport executes a report with parameters
func (s *ReportsService) RunReport(reportKey string, req store.RunReportRequest) (*store.ReportRun, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceREST, "Running report", map[string]interface{}{
		"report_key":    reportKey,
		"datasource_id": req.DatasourceID,
	})

	// Get report
	var report store.Report
	if err := s.db.Where("key = ?", reportKey).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		return nil, fmt.Errorf("failed to find report: %w", err)
	}

	// Get latest report version
	var reportVersion store.ReportVersion
	if err := s.db.Where("report_id = ?", report.ID).Order("version DESC").First(&reportVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no report version found")
		}
		return nil, fmt.Errorf("failed to find report version: %w", err)
	}

	// Determine datasource
	datasourceID := reportVersion.DatasourceID
	if req.DatasourceID != nil {
		datasourceID = req.DatasourceID
	}
	if datasourceID == nil || *datasourceID == "" {
		return nil, fmt.Errorf("no datasource specified")
	}

	// Get datasource connector
	connector, err := s.registry.GetDatasource(*datasourceID)
	if err != nil {
		return nil, fmt.Errorf("datasource not found: %w", err)
	}

	// Extract SQL from def_json (expects a JSON with {"sql": "..."})
	sqlText := extractSQLFromDef(reportVersion.DefJSON)
	logger.LogInfo(logger.ServiceREST, "Extracted SQL from report version", map[string]interface{}{
		"report_id":  report.ID,
		"version_id": reportVersion.ID,
		"has_sql":    sqlText != "",
	})
	if sqlText == "" {
		return nil, fmt.Errorf("report version def_json does not contain sql")
	}

	// Replace simple placeholders {{param}} with provided params (dev only)
	sqlPrepared := replacePlaceholders(sqlText, req.Params)

	// Execute SQL and get results
	results, rowCount, execErr := executeAndGetResults(connector.DB, sqlPrepared)
	if execErr != nil {
		logger.LogError(logger.ServiceREST, "Report SQL execution failed", execErr, map[string]interface{}{
			"report_id":  report.ID,
			"version_id": reportVersion.ID,
			"sql":        sqlPrepared,
		})
	} else {
		logger.LogInfo(logger.ServiceREST, "Report SQL executed", map[string]interface{}{
			"rows": rowCount,
			"sql":  sqlPrepared,
		})
	}
	status := "completed"
	var errText string
	if execErr != nil {
		status = "failed"
		errText = execErr.Error()
	}

	finished := time.Now()
	reportRun := &store.ReportRun{
		ReportID:        report.ID,
		ReportVersionID: reportVersion.ID,
		DatasourceID:    *datasourceID,
		ParamsJSON:      fmt.Sprintf(`{"params": %v}`, req.Params),
		SQLText:         sqlPrepared,
		RowCount:        rowCount,
		Results:         results,
		StartedAt:       start,
		FinishedAt:      &finished,
		Status:          status,
		ErrorText:       errText,
	}

	if err := s.db.Create(reportRun).Error; err != nil {
		logger.LogError(logger.ServiceREST, "Failed to create report run", err, map[string]interface{}{
			"report_id": report.ID,
		})
		return nil, fmt.Errorf("failed to create report run: %w", err)
	}

	// Manually populate the relationships
	populatedReportRun := *reportRun

	// Load Report
	logger.LogInfo(logger.ServiceREST, "Loading report", map[string]interface{}{
		"report_id": report.ID,
		"run_id":    reportRun.ID,
	})
	if err := s.db.First(&populatedReportRun.Report, report.ID).Error; err != nil {
		logger.LogWarn(logger.ServiceREST, "Failed to load report", map[string]interface{}{
			"report_id": report.ID,
			"error":     err.Error(),
		})
	} else {
		logger.LogInfo(logger.ServiceREST, "Successfully loaded report", map[string]interface{}{
			"report_id": populatedReportRun.Report.ID,
			"key":       populatedReportRun.Report.Key,
			"title":     populatedReportRun.Report.Title,
		})
	}

	// Load ReportVersion
	if err := s.db.First(&populatedReportRun.ReportVersion, reportVersion.ID).Error; err != nil {
		logger.LogWarn(logger.ServiceREST, "Failed to load report version", map[string]interface{}{
			"version_id": reportVersion.ID,
			"error":      err.Error(),
		})
	} else {
		logger.LogInfo(logger.ServiceREST, "Successfully loaded report version", map[string]interface{}{
			"version_id": populatedReportRun.ReportVersion.ID,
			"version":    populatedReportRun.ReportVersion.Version,
		})
	}

	// Load Datasource - check if it exists in database first
	if err := s.db.Where("id = ?", *datasourceID).First(&populatedReportRun.Datasource).Error; err != nil {
		logger.LogWarn(logger.ServiceREST, "Failed to load datasource from database", map[string]interface{}{
			"datasource_id": *datasourceID,
			"error":         err.Error(),
		})
		// Try to get from registry instead
		if connector, err := s.registry.GetDatasource(*datasourceID); err == nil {
			populatedReportRun.Datasource = store.Datasource{
				ID:          *datasourceID,
				Kind:        connector.Kind,
				DisplayName: connector.DisplayName,
				DSN:         "", // DSN not available in connector
				IsDefault:   connector.IsDefault,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			logger.LogInfo(logger.ServiceREST, "Loaded datasource from registry", map[string]interface{}{
				"datasource_id": *datasourceID,
				"kind":          connector.Kind,
			})
		}
	} else {
		logger.LogInfo(logger.ServiceREST, "Successfully loaded datasource from database", map[string]interface{}{
			"datasource_id": populatedReportRun.Datasource.ID,
			"kind":          populatedReportRun.Datasource.Kind,
		})
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceREST, "Report run finished", map[string]interface{}{
		"run_id":    populatedReportRun.ID,
		"report_id": populatedReportRun.ReportID,
		"status":    status,
		"rows":      rowCount,
		"duration":  duration.String(),
	})

	return &populatedReportRun, nil
}

// GetLatestReportRun retrieves the most recent report run for a given report ID
func (s *ReportsService) GetLatestReportRun(reportID uint) (*store.ReportRun, error) {
	var reportRun store.ReportRun
	if err := s.db.Where("report_id = ?", reportID).Order("started_at DESC").First(&reportRun).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no report runs found for report ID %d", reportID)
		}
		return nil, fmt.Errorf("failed to retrieve report run: %w", err)
	}

	// Load relationships
	if err := s.db.Preload("Report").Preload("ReportVersion").Preload("Datasource").First(&reportRun, reportRun.ID).Error; err != nil {
		logger.LogWarn(logger.ServiceREST, "Failed to load report run relationships", map[string]interface{}{
			"run_id":    reportRun.ID,
			"report_id": reportID,
			"error":     err.Error(),
		})
		// Return the basic report run if preload fails
		return &reportRun, nil
	}

	return &reportRun, nil
}

func extractSQLFromDef(defJSON string) string {
	// Case 1: defJSON contains a JSON object
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(defJSON), &obj); err == nil {
		if v, ok := obj["sql"].(string); ok {
			return v
		}
	} else {
		// Case 2: defJSON is a JSON-encoded string containing an object
		var raw string
		if err2 := json.Unmarshal([]byte(defJSON), &raw); err2 == nil {
			var inner map[string]interface{}
			if err3 := json.Unmarshal([]byte(raw), &inner); err3 == nil {
				if v, ok := inner["sql"].(string); ok {
					return v
				}
			}
		}
	}
	// Fallback: regex extraction
	re := regexp.MustCompile(`"sql"\s*:\s*"([\s\S]*?)"`)
	if m := re.FindStringSubmatch(defJSON); len(m) == 2 {
		s := strings.ReplaceAll(m[1], `\n`, "\n")
		s = strings.ReplaceAll(s, `\t`, "\t")
		s = strings.ReplaceAll(s, `\"`, `"`)
		return s
	}
	return ""
}

func replacePlaceholders(sqlText string, params map[string]interface{}) string {
	if params == nil {
		return sqlText
	}
	out := sqlText
	for k, v := range params {
		raw := fmt.Sprintf("%v", v)
		quoted := "'" + strings.ReplaceAll(raw, "'", "''") + "'"

		// First handle quoted placeholders like '{{param}}'
		quotedPlaceholder := "'{{" + k + "}}'"
		if strings.Contains(out, quotedPlaceholder) {
			out = strings.ReplaceAll(out, quotedPlaceholder, quoted)
		} else {
			// Then handle unquoted placeholders like {{param}}
			placeholder := "{{" + k + "}}"
			if strings.Contains(out, placeholder) {
				out = strings.ReplaceAll(out, placeholder, quoted)
			}
		}
	}
	return out
}

func needsQuoting(val string) bool {
	// quote if contains non-digit and not starting with {{
	if val == "" {
		return true
	}
	for _, r := range val {
		if r < '0' || r > '9' {
			return true
		}
	}
	return false
}

func executeAndCountRows(db *sql.DB, query string) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("nil db connection")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}
	count := 0
	values := make([]interface{}, len(cols))
	scanArgs := make([]interface{}, len(cols))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return count, err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return count, err
	}
	return count, nil
}

// executeAndGetResults executes a query and returns both results and row count
func executeAndGetResults(db *sql.DB, query string) (string, int, error) {
	if db == nil {
		return "", 0, fmt.Errorf("nil db connection")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", 0, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", 0, err
	}

	var results []map[string]interface{}
	values := make([]interface{}, len(cols))
	scanArgs := make([]interface{}, len(cols))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			return "", 0, err
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return "", 0, err
	}

	// Convert results to JSON
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(resultsJSON), len(results), nil
}

// ExportReport exports a report in various formats
func (s *ReportsService) ExportReport(reportKey string, format string) ([]byte, error) {
	logger.LogInfo(logger.ServiceREST, "Exporting report", map[string]interface{}{
		"report_key": reportKey,
		"format":     format,
	})

	// Get report
	var report store.Report
	if err := s.db.Where("key = ?", reportKey).First(&report).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("report not found")
		}
		return nil, fmt.Errorf("failed to find report: %w", err)
	}

	// Get latest report version
	var reportVersion store.ReportVersion
	if err := s.db.Where("report_id = ?", report.ID).
		Order("version DESC").
		First(&reportVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no report version found")
		}
		return nil, fmt.Errorf("failed to find report version: %w", err)
	}

	// Export based on format
	switch format {
	case "json":
		exportData := map[string]interface{}{
			"report": map[string]interface{}{
				"id":    report.ID,
				"key":   report.Key,
				"title": report.Title,
				"owner": report.Owner,
			},
			"version": map[string]interface{}{
				"id":               reportVersion.ID,
				"version":          reportVersion.Version,
				"scope_version_id": reportVersion.ScopeVersionID,
				"datasource_id":    reportVersion.DatasourceID,
				"def_json":         reportVersion.DefJSON,
				"created_at":       reportVersion.CreatedAt,
			},
		}
		return []byte(fmt.Sprintf(`%v`, exportData)), nil
	case "yaml":
		// TODO: Implement YAML export
		return nil, fmt.Errorf("YAML export not implemented")
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ListReports returns all reports
func (s *ReportsService) ListReports() ([]store.Report, error) {
	var reports []store.Report
	if err := s.db.Order("created_at DESC").Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

// GetReportByID retrieves a report by numeric ID
func (s *ReportsService) GetReportByID(id uint) (*store.Report, error) {
	var report store.Report
	if err := s.db.First(&report, id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// DeleteReportByID deletes a report by ID
func (s *ReportsService) DeleteReportByID(id uint) error {
	return s.db.Delete(&store.Report{}, id).Error
}

// RunReportByID runs a report by numeric ID (uses latest version)
func (s *ReportsService) RunReportByID(id uint, req store.RunReportRequest) (*store.ReportRun, error) {
	var report store.Report
	if err := s.db.First(&report, id).Error; err != nil {
		return nil, err
	}
	return s.RunReport(report.Key, req)
}

// GetLatestReportSQLByReportID returns the SQL text from the latest version of a report
func (s *ReportsService) GetLatestReportSQLByReportID(reportID uint) (string, error) {
	// Find report
	var report store.Report
	if err := s.db.First(&report, reportID).Error; err != nil {
		return "", fmt.Errorf("failed to find report: %w", err)
	}

	// Get latest report version
	var reportVersion store.ReportVersion
	if err := s.db.Where("report_id = ?", report.ID).Order("version DESC").First(&reportVersion).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("no report version found")
		}
		return "", fmt.Errorf("failed to find report version: %w", err)
	}

	// Extract SQL from definition JSON
	sqlText := extractSQLFromDef(reportVersion.DefJSON)
	return sqlText, nil
}
