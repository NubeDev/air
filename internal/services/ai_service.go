package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/llm"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/ollama/ollama/api"
	"gorm.io/gorm"
)

// AIService handles AI-related business logic
type AIService struct {
	registry          *datasource.Registry
	db                *gorm.DB
	llmClient         llm.LLMClient
	sqlClient         llm.LLMClient
	Config            *config.Config
	datasourceService *DatasourceService
}

// NewAIService creates a new AI service
func NewAIService(registry *datasource.Registry, db *gorm.DB, cfg *config.Config, datasourceService *DatasourceService) (*AIService, error) {
	// Initialize LLM client for chat
	llmClient, err := llm.NewLLMClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	// Initialize SQL client (could be different from chat client)
	sqlClient, err := llm.NewSQLClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQL client: %w", err)
	}

	return &AIService{
		registry:          registry,
		db:                db,
		llmClient:         llmClient,
		sqlClient:         sqlClient,
		Config:            cfg,
		datasourceService: datasourceService,
	}, nil
}

// BuildIR builds Intermediate Representation from scope
func (s *AIService) BuildIR(req store.BuildIRRequest) (map[string]interface{}, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceAI, "Building Intermediate Representation", map[string]interface{}{
		"scope_version_id": req.ScopeVersionID,
	})

	// Load scope version
	var scopeVersion store.ScopeVersion
	if err := s.db.First(&scopeVersion, req.ScopeVersionID).Error; err != nil {
		logger.LogError(logger.ServiceAI, "Failed to load scope version", err, map[string]interface{}{
			"scope_version_id": req.ScopeVersionID,
		})
		return nil, fmt.Errorf("scope version not found")
	}

	// Get the datasource to understand the schema
	var datasource store.Datasource
	if err := s.db.Where("id = ?", req.DatasourceID).First(&datasource).Error; err != nil {
		logger.LogError(logger.ServiceAI, "Failed to load datasource", err, map[string]interface{}{
			"datasource_id": req.DatasourceID,
		})
		return nil, fmt.Errorf("datasource not found")
	}

	// Get schema information for the datasource
	schemaNotes, err := s.datasourceService.GetSchema(req.DatasourceID)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to get schema", err, map[string]interface{}{
			"datasource_id": req.DatasourceID,
		})
		// Continue without schema info
		schemaNotes = []store.SchemaNote{}
	}

	// Compose chat to convert scope markdown to IR JSON with schema context
	systemMsg := llm.Message{
		Role:    "system",
		Content: "You are an expert data analyst. Convert the user's scope (Markdown) into a compact JSON Intermediate Representation (IR) for analytics. Respond with ONLY valid JSON (no code fences, no commentary).\n\nIMPORTANT: \n- Use ONLY the actual column names from the schema information provided\n- If the goal mentions 'sum sales per customer name', you MUST include:\n  * select: [\"customer_name\", {\"SUM(total_amount)\": \"total_sales\"}]\n  * group_by: [\"customer_name\"]\n  * filters: [{\"field\": \"customer_name\", \"op\": \"=\", \"value\": \"{{customer_name}}\"}]\n- Always include proper aggregation functions (SUM, COUNT, AVG, etc.) when needed\n- Make filters parameterizable using {{param_name}} syntax\n- NEVER leave select array empty - always specify what to select\n\nIR schema: {\n  \"dataset\": string,                  // main table/view or dataset\n  \"select\": [string | object],        // columns or expressions to select (use actual column names)\n  \"filters\": [                        // simple filter list\n    {\n      \"field\": string,\n      \"op\": one of [=,!=,>,>=,<,<=,IN,NOT IN,LIKE,BETWEEN],\n      \"value\": any | [any, any] | \"{{param_name}}\"\n    }\n  ],\n  \"group_by\": [string],               // optional group by columns (use actual column names)\n  \"order_by\": [{\"field\": string, \"dir\": one of [ASC, DESC]}],\n  \"limit\": number                     // optional row limit\n}",
	}

	// Include schema information in the user message
	schemaInfo := ""
	if len(schemaNotes) > 0 {
		var schemaStrings []string
		for _, note := range schemaNotes {
			schemaStrings = append(schemaStrings, note.MD)
		}
		schemaInfo = fmt.Sprintf("\n\nAvailable schema information:\n%s", strings.Join(schemaStrings, "\n"))
	}

	userMsg := llm.Message{
		Role:    "user",
		Content: fmt.Sprintf("Scope Markdown:\n\n%s%s\n\nGenerate IR now.", scopeVersion.ScopeMD, schemaInfo),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	model := llm.GetModelName(s.Config, "chat")

	chatReq := llm.ChatRequest{
		Model:    model,
		Messages: []llm.Message{systemMsg, userMsg},
		Stream:   false,
		Options: &api.Options{
			Temperature: 0.2,
			TopP:        0.9,
		},
	}

	resp, err := s.llmClient.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to build IR: %w", err)
	}

	// Sanitize/parse JSON
	content := strings.TrimSpace(resp.Message.Content)
	jsonBytes := sanitizeModelJSONOutput(content)

	var ir map[string]interface{}
	if uErr := json.Unmarshal(jsonBytes, &ir); uErr != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse IR JSON", uErr, map[string]interface{}{
			"content_head": content[:min(200, len(content))],
		})
		return nil, fmt.Errorf("model did not return valid IR JSON: %w", uErr)
	}

	// Persist IR back to scope version
	irJSON, _ := json.Marshal(ir)
	scopeVersion.IRJSON = string(irJSON)
	if err := s.db.Save(&scopeVersion).Error; err != nil {
		logger.LogError(logger.ServiceAI, "Failed to save IR to scope version", err, map[string]interface{}{
			"scope_version_id": scopeVersion.ID,
		})
		return nil, fmt.Errorf("failed to save IR: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceAI, "IR building completed", map[string]interface{}{
		"scope_version_id": req.ScopeVersionID,
		"duration":         duration.String(),
	})

	return ir, nil
}

// GenerateSQLFromIR generates SQL from IR for a specific datasource
func (s *AIService) GenerateSQLFromIR(req store.GenerateSQLRequest) (string, map[string]interface{}, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceAI, "Generating SQL from IR (SQLCoder)", map[string]interface{}{
		"datasource_id": req.DatasourceID,
	})

	// Get datasource (to determine dialect label)
	connector, err := s.registry.GetDatasource(req.DatasourceID)
	if err != nil {
		return "", nil, fmt.Errorf("datasource not found: %w", err)
	}

	// Convert IR to natural language prompt for SQLCoder
	prompt, err := s.buildSQLCoderPromptFromIR(req.IR, connector.Kind)
	if err != nil {
		return "", nil, fmt.Errorf("failed to build SQLCoder prompt: %w", err)
	}

	// Get schema information for the datasource
	schema, err := s.getDatasourceSchema(req.DatasourceID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get datasource schema: %w", err)
	}

	// Use SQLCoder to generate SQL
	sql, err := s.GenerateSQL(prompt, schema)
	if err != nil {
		return "", nil, fmt.Errorf("SQLCoder generation failed: %w", err)
	}

	if sql == "" {
		return "", nil, fmt.Errorf("SQLCoder returned empty result")
	}

	// Use the SQLCoder-generated SQL directly
	// SQLCoder already knows the database type and generates appropriate SQL

	// Safety report (structure only; checks can be expanded later)
	safetyReport := map[string]interface{}{
		"read_only": true,
		"warnings":  []string{},
		"checks":    map[string]any{"generated_by": "sqlcoder"},
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceAI, "SQL generation completed", map[string]interface{}{
		"datasource_id": req.DatasourceID,
		"duration":      duration.String(),
	})

	return sql, safetyReport, nil
}

// generateDatabaseSpecificSQL generates SQL optimized for specific database types
func (s *AIService) generateDatabaseSpecificSQL(ir map[string]interface{}, dbKind string) string {
	switch strings.ToLower(dbKind) {
	case "sqlite", "sqlite3":
		return s.generateSQLiteSQL(ir)
	case "postgres", "postgresql", "timescaledb":
		return s.generatePostgreSQLSQL(ir)
	case "mysql":
		return s.generateMySQLSQL(ir)
	default:
		// Fallback to basic SQL generation
		return s.generateBasicSQL(ir)
	}
}

// generateSQLiteSQL generates SQL optimized for SQLite
func (s *AIService) generateSQLiteSQL(ir map[string]interface{}) string {
	var sql strings.Builder

	// Extract IR components
	dataset, _ := ir["dataset"].(string)
	selectFields, _ := ir["select"].([]interface{})
	groupBy, _ := ir["group_by"].([]interface{})
	orderBy, _ := ir["order_by"].([]interface{})
	limit, _ := ir["limit"].(float64)
	filters, _ := ir["filters"].([]interface{})

	// Build SELECT clause
	sql.WriteString("SELECT ")
	selectClause := s.buildSelectClause(selectFields, "sqlite")
	sql.WriteString(selectClause)

	// Build FROM clause
	sql.WriteString(fmt.Sprintf(" FROM %s", dataset))

	// Build WHERE clause with date range filtering
	whereClause := s.buildWhereClause(filters, "sqlite")
	if whereClause != "" {
		sql.WriteString(" WHERE ")
		sql.WriteString(whereClause)
	} else {
		// Add date range filtering for time-series data
		sql.WriteString(" WHERE ")
		sql.WriteString(s.buildDateRangeWhereClause("timestamp", "sqlite"))
	}

	// Build GROUP BY clause
	if len(groupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		groupClause := s.buildGroupByClause(groupBy, "sqlite")
		sql.WriteString(groupClause)
	}

	// Build ORDER BY clause
	if len(orderBy) > 0 {
		sql.WriteString(" ORDER BY ")
		orderClause := s.buildOrderByClause(orderBy, "sqlite")
		sql.WriteString(orderClause)
	}

	// Build LIMIT clause
	if limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", int(limit)))
	}

	return sql.String()
}

// generatePostgreSQLSQL generates SQL optimized for PostgreSQL
func (s *AIService) generatePostgreSQLSQL(ir map[string]interface{}) string {
	var sql strings.Builder

	// Extract IR components
	dataset, _ := ir["dataset"].(string)
	selectFields, _ := ir["select"].([]interface{})
	groupBy, _ := ir["group_by"].([]interface{})
	orderBy, _ := ir["order_by"].([]interface{})
	limit, _ := ir["limit"].(float64)
	filters, _ := ir["filters"].([]interface{})

	// Build SELECT clause
	sql.WriteString("SELECT ")
	selectClause := s.buildSelectClause(selectFields, "postgres")
	sql.WriteString(selectClause)

	// Build FROM clause
	sql.WriteString(fmt.Sprintf(" FROM %s", dataset))

	// Build WHERE clause
	whereClause := s.buildWhereClause(filters, "postgres")
	if whereClause != "" {
		sql.WriteString(" WHERE ")
		sql.WriteString(whereClause)
	}

	// Build GROUP BY clause
	if len(groupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		groupClause := s.buildGroupByClause(groupBy, "postgres")
		sql.WriteString(groupClause)
	}

	// Build ORDER BY clause
	if len(orderBy) > 0 {
		sql.WriteString(" ORDER BY ")
		orderClause := s.buildOrderByClause(orderBy, "postgres")
		sql.WriteString(orderClause)
	}

	// Build LIMIT clause
	if limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", int(limit)))
	}

	return sql.String()
}

// generateMySQLSQL generates SQL optimized for MySQL
func (s *AIService) generateMySQLSQL(ir map[string]interface{}) string {
	var sql strings.Builder

	// Extract IR components
	dataset, _ := ir["dataset"].(string)
	selectFields, _ := ir["select"].([]interface{})
	groupBy, _ := ir["group_by"].([]interface{})
	orderBy, _ := ir["order_by"].([]interface{})
	limit, _ := ir["limit"].(float64)
	filters, _ := ir["filters"].([]interface{})

	// Build SELECT clause
	sql.WriteString("SELECT ")
	selectClause := s.buildSelectClause(selectFields, "mysql")
	sql.WriteString(selectClause)

	// Build FROM clause
	sql.WriteString(fmt.Sprintf(" FROM %s", dataset))

	// Build WHERE clause
	whereClause := s.buildWhereClause(filters, "mysql")
	if whereClause != "" {
		sql.WriteString(" WHERE ")
		sql.WriteString(whereClause)
	}

	// Build GROUP BY clause
	if len(groupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		groupClause := s.buildGroupByClause(groupBy, "mysql")
		sql.WriteString(groupClause)
	}

	// Build ORDER BY clause
	if len(orderBy) > 0 {
		sql.WriteString(" ORDER BY ")
		orderClause := s.buildOrderByClause(orderBy, "mysql")
		sql.WriteString(orderClause)
	}

	// Build LIMIT clause
	if limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", int(limit)))
	}

	return sql.String()
}

// generateBasicSQL generates basic SQL as fallback
func (s *AIService) generateBasicSQL(ir map[string]interface{}) string {
	// Simple fallback - just return a basic query
	dataset, _ := ir["dataset"].(string)
	return fmt.Sprintf("SELECT * FROM %s LIMIT 100", dataset)
}

// buildSelectClause builds the SELECT clause for different database types
func (s *AIService) buildSelectClause(selectFields []interface{}, dbType string) string {
	if len(selectFields) == 0 {
		return "*"
	}

	var clauses []string
	for _, field := range selectFields {
		clause := s.buildSelectField(field, dbType)
		if clause != "" {
			clauses = append(clauses, clause)
		}
	}

	if len(clauses) == 0 {
		return "*"
	}

	return strings.Join(clauses, ", ")
}

// buildSelectField builds a single SELECT field
func (s *AIService) buildSelectField(field interface{}, dbType string) string {
	fieldMap, ok := field.(map[string]interface{})
	if !ok {
		// Simple field name
		return fmt.Sprintf("%v", field)
	}

	fieldName, _ := fieldMap["field"].(string)
	funcName, _ := fieldMap["func"].(string)
	alias, _ := fieldMap["alias"].(string)
	as, _ := fieldMap["as"].(string)

	// Use alias or as for the final name
	finalName := alias
	if finalName == "" {
		finalName = as
	}
	if finalName == "" {
		finalName = fieldName
	}

	// Check for aggregation functions like {"sum": "sales"}
	for funcName, fieldName := range fieldMap {
		if funcName == "sum" || funcName == "avg" || funcName == "count" || funcName == "max" || funcName == "min" {
			if fieldStr, ok := fieldName.(string); ok {
				expr := s.buildFunctionCall(fieldStr, funcName, dbType)
				// Add alias for aggregation
				expr += " AS " + funcName + "_" + fieldStr
				return expr
			}
		}
	}

	// Check for full SQL expressions like {"SUM(total_amount)": "total_sales"}
	for expr, value := range fieldMap {
		if (value == nil || value == "") && (strings.Contains(strings.ToUpper(expr), "SUM(") ||
			strings.Contains(strings.ToUpper(expr), "AVG(") ||
			strings.Contains(strings.ToUpper(expr), "COUNT(") ||
			strings.Contains(strings.ToUpper(expr), "MAX(") ||
			strings.Contains(strings.ToUpper(expr), "MIN(")) {
			// This is a full SQL expression, return it as-is
			return expr
		}
		// Also handle cases where the value is the alias
		if value != nil && (strings.Contains(strings.ToUpper(expr), "SUM(") ||
			strings.Contains(strings.ToUpper(expr), "AVG(") ||
			strings.Contains(strings.ToUpper(expr), "COUNT(") ||
			strings.Contains(strings.ToUpper(expr), "MAX(") ||
			strings.Contains(strings.ToUpper(expr), "MIN(")) {
			// This is a full SQL expression with alias, add AS clause
			alias := fmt.Sprintf("%v", value)
			return expr + " AS " + alias
		}
	}

	// Build the field expression
	var expr string
	if funcName != "" {
		expr = s.buildFunctionCall(fieldName, funcName, dbType)
	} else {
		expr = fieldName
	}

	// Add alias if we have a valid final name
	if finalName != "" {
		return fmt.Sprintf("%s AS %s", expr, finalName)
	}

	return expr
}

// buildFunctionCall builds a function call for different database types
func (s *AIService) buildFunctionCall(field, funcName, dbType string) string {
	switch strings.ToLower(funcName) {
	case "sum":
		return fmt.Sprintf("SUM(%s)", field)
	case "avg":
		return fmt.Sprintf("AVG(%s)", field)
	case "count":
		return fmt.Sprintf("COUNT(%s)", field)
	case "max":
		return fmt.Sprintf("MAX(%s)", field)
	case "min":
		return fmt.Sprintf("MIN(%s)", field)
	case "date":
		switch dbType {
		case "sqlite":
			return fmt.Sprintf("date(%s)", field)
		case "postgres":
			return fmt.Sprintf("date_trunc('day', %s)", field)
		case "mysql":
			return fmt.Sprintf("DATE(%s)", field)
		default:
			return fmt.Sprintf("date(%s)", field)
		}
	default:
		return field
	}
}

// buildGroupByClause builds the GROUP BY clause
func (s *AIService) buildGroupByClause(groupBy []interface{}, dbType string) string {
	var clauses []string
	for _, field := range groupBy {
		clause := s.buildGroupByField(field, dbType)
		if clause != "" {
			clauses = append(clauses, clause)
		}
	}
	return strings.Join(clauses, ", ")
}

// buildGroupByField builds a single GROUP BY field
func (s *AIService) buildGroupByField(field interface{}, dbType string) string {
	fieldMap, ok := field.(map[string]interface{})
	if !ok {
		// Simple field name
		return fmt.Sprintf("%v", field)
	}

	fieldName, _ := fieldMap["field"].(string)
	funcName, _ := fieldMap["func"].(string)

	if funcName != "" {
		return s.buildFunctionCall(fieldName, funcName, dbType)
	}

	return fieldName
}

// buildOrderByClause builds the ORDER BY clause
func (s *AIService) buildOrderByClause(orderBy []interface{}, dbType string) string {
	var clauses []string
	for _, field := range orderBy {
		clause := s.buildOrderByField(field, dbType)
		if clause != "" {
			clauses = append(clauses, clause)
		}
	}
	return strings.Join(clauses, ", ")
}

// buildOrderByField builds a single ORDER BY field
func (s *AIService) buildOrderByField(field interface{}, dbType string) string {
	fieldMap, ok := field.(map[string]interface{})
	if !ok {
		// Simple field name
		return fmt.Sprintf("%v", field)
	}

	fieldName, _ := fieldMap["field"].(string)
	funcName, _ := fieldMap["func"].(string)
	direction, _ := fieldMap["dir"].(string)

	// Build the field expression
	var expr string
	if funcName != "" {
		expr = s.buildFunctionCall(fieldName, funcName, dbType)
	} else {
		expr = fieldName
	}

	// Add direction
	if strings.ToUpper(direction) == "DESC" {
		expr += " DESC"
	} else {
		expr += " ASC"
	}

	return expr
}

// buildWhereClause builds the WHERE clause
func (s *AIService) buildWhereClause(filters []interface{}, dbType string) string {
	if len(filters) == 0 {
		return ""
	}

	var conditions []string
	for _, filter := range filters {
		filterMap, ok := filter.(map[string]interface{})
		if !ok {
			continue
		}

		field, _ := filterMap["field"].(string)
		op, _ := filterMap["op"].(string)
		value := filterMap["value"]

		if field == "" || op == "" {
			continue
		}

		// Handle different operators
		switch op {
		case "=":
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "{{") && strings.HasSuffix(valueStr, "}}") {
				// This is a placeholder, use it as-is
				conditions = append(conditions, fmt.Sprintf("%s = %s", field, valueStr))
			} else {
				// This is a literal value, quote it
				quoted := "'" + strings.ReplaceAll(fmt.Sprintf("%v", value), "'", "''") + "'"
				conditions = append(conditions, fmt.Sprintf("%s = %s", field, quoted))
			}
		case "!=":
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "{{") && strings.HasSuffix(valueStr, "}}") {
				conditions = append(conditions, fmt.Sprintf("%s != %s", field, valueStr))
			} else {
				quoted := "'" + strings.ReplaceAll(fmt.Sprintf("%v", value), "'", "''") + "'"
				conditions = append(conditions, fmt.Sprintf("%s != %s", field, quoted))
			}
		case "LIKE":
			if valueStr, ok := value.(string); ok && strings.HasPrefix(valueStr, "{{") && strings.HasSuffix(valueStr, "}}") {
				conditions = append(conditions, fmt.Sprintf("%s LIKE %s", field, valueStr))
			} else {
				quoted := "'" + strings.ReplaceAll(fmt.Sprintf("%v", value), "'", "''") + "'"
				conditions = append(conditions, fmt.Sprintf("%s LIKE %s", field, quoted))
			}
		case "IN":
			if valueList, ok := value.([]interface{}); ok {
				var quotedValues []string
				for _, v := range valueList {
					quoted := "'" + strings.ReplaceAll(fmt.Sprintf("%v", v), "'", "''") + "'"
					quotedValues = append(quotedValues, quoted)
				}
				conditions = append(conditions, fmt.Sprintf("%s IN (%s)", field, strings.Join(quotedValues, ", ")))
			}
		case "BETWEEN":
			if valueList, ok := value.([]interface{}); ok && len(valueList) == 2 {
				start := "'" + strings.ReplaceAll(fmt.Sprintf("%v", valueList[0]), "'", "''") + "'"
				end := "'" + strings.ReplaceAll(fmt.Sprintf("%v", valueList[1]), "'", "''") + "'"
				conditions = append(conditions, fmt.Sprintf("%s BETWEEN %s AND %s", field, start, end))
			}
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	return strings.Join(conditions, " AND ")
}

// buildDateRangeWhereClause builds a WHERE clause for date range filtering
func (s *AIService) buildDateRangeWhereClause(timestampField string, dbType string) string {
	switch dbType {
	case "sqlite":
		return fmt.Sprintf("date(%s) BETWEEN '{{start_date}}' AND '{{end_date}}'", timestampField)
	case "postgres":
		return fmt.Sprintf("date_trunc('day', %s) BETWEEN '{{start_date}}' AND '{{end_date}}'", timestampField)
	case "mysql":
		return fmt.Sprintf("DATE(%s) BETWEEN '{{start_date}}' AND '{{end_date}}'", timestampField)
	default:
		return fmt.Sprintf("date(%s) BETWEEN '{{start_date}}' AND '{{end_date}}'", timestampField)
	}
}

// AnalyzeRun analyzes a report run with AI
func (s *AIService) AnalyzeRun(runID uint, req store.AnalyzeRunRequest) (*store.ReportAnalysis, error) {
	start := time.Now()

	// Load run
	var run store.ReportRun
	if err := s.db.First(&run, runID).Error; err != nil {
		logger.LogError(logger.ServiceAI, "Failed to load report run", err, map[string]interface{}{
			"run_id": runID,
		})
		return nil, fmt.Errorf("run not found")
	}

	// Build prompt requesting structured verdict JSON and markdown analysis
	rubricVersion := req.RubricVersion
	if rubricVersion == "" {
		rubricVersion = "v1"
	}

	systemMsg := llm.Message{
		Role:    "system",
		Content: "You are a senior data analyst. Analyze the SQL execution results and produce: (1) a JSON verdict with keys: {score: number 0-100, severity: one of [info,warning,error], key_findings: [string], anomalies: [string], recommendations: [string]}, and (2) a concise Markdown analysis. Respond with ONLY JSON in the shape {\"verdict\": {...}, \"analysis_md\": string}.",
	}

	summary := fmt.Sprintf("Run Summary:\nStatus: %s\nRow Count: %d\nParams: %s\nSQL:\n%s\n\n", run.Status, run.RowCount, run.ParamsJSON, run.SQLText)
	if run.ErrorText != "" {
		summary += fmt.Sprintf("Error: %s\n", run.ErrorText)
	}

	userMsg := llm.Message{Role: "user", Content: summary}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	model := llm.GetModelName(s.Config, "chat")

	chatReq := llm.ChatRequest{
		Model:    model,
		Messages: []llm.Message{systemMsg, userMsg},
		Stream:   false,
		Options:  &api.Options{Temperature: 0.3, TopP: 0.9},
	}

	resp, err := s.llmClient.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	content := strings.TrimSpace(resp.Message.Content)
	jsonBytes := sanitizeModelJSONOutput(content)

	// Parse expected JSON shape
	var parsed struct {
		Verdict    map[string]interface{} `json:"verdict"`
		AnalysisMD string                 `json:"analysis_md"`
	}
	if uErr := json.Unmarshal(jsonBytes, &parsed); uErr != nil {
		logger.LogError(logger.ServiceAI, "Failed to parse analysis JSON", uErr, map[string]interface{}{
			"content_head": content[:min(200, len(content))],
		})
		return nil, fmt.Errorf("model did not return valid analysis JSON: %w", uErr)
	}

	verdictJSON, _ := json.Marshal(parsed.Verdict)

	analysis := &store.ReportAnalysis{
		RunID:         run.ID,
		ModelUsed:     model,
		RubricVersion: rubricVersion,
		VerdictJSON:   string(verdictJSON),
		AnalysisMD:    parsed.AnalysisMD,
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(analysis).Error; err != nil {
		logger.LogError(logger.ServiceAI, "Failed to save analysis", err, map[string]interface{}{
			"run_id": run.ID,
		})
		return nil, fmt.Errorf("failed to save analysis: %w", err)
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceAI, "Run analysis completed", map[string]interface{}{
		"run_id":   run.ID,
		"duration": duration.String(),
	})

	return analysis, nil
}

// GetAITools returns available AI tools
func (s *AIService) GetAITools() ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check LLM health first
	if err := s.llmClient.Health(ctx); err != nil {
		return nil, fmt.Errorf("LLM service unavailable: %w", err)
	}

	// List available models
	models, err := s.llmClient.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	chatModel := llm.GetModelName(s.Config, "chat")
	sqlModel := llm.GetModelName(s.Config, "sql")

	tools := []map[string]interface{}{
		{
			"name":        "chat_completion",
			"description": "Generate chat completions using available models",
			"models":      []string{chatModel},
			"type":        "chat",
		},
		{
			"name":        "sql_generation",
			"description": "Generate SQL queries using SQL model",
			"models":      []string{sqlModel},
			"type":        "sql",
		},
		{
			"name":        "text_generation",
			"description": "Generate text using available models",
			"models":      []string{chatModel},
			"type":        "generate",
		},
	}

	// Add model information
	modelInfo := make([]map[string]interface{}, 0, len(models.Models))
	for _, model := range models.Models {
		modelInfo = append(modelInfo, map[string]interface{}{
			"name":     model.Name,
			"size":     model.Size,
			"modified": model.ModifiedAt,
		})
	}

	provider := "Ollama"
	if s.Config.Models.ChatPrimary == "openai" {
		provider = "OpenAI"
	}

	return append(tools, map[string]interface{}{
		"name":        "available_models",
		"description": fmt.Sprintf("List of available %s models", provider),
		"models":      modelInfo,
		"type":        "info",
	}), nil
}

// ChatCompletion performs a chat completion using the configured model
func (s *AIService) ChatCompletion(messages []llm.Message) (*llm.ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	model := llm.GetModelName(s.Config, "chat")

	req := llm.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options: &api.Options{
			Temperature: 0.7,
			TopP:        0.9,
		},
	}

	return s.llmClient.ChatCompletion(ctx, req)
}

// AiRaw performs raw AI completion without any system prompts or backend interference
func (s *AIService) AiRaw(messages []llm.Message, modelOverride string) (*llm.ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Use model override if provided, otherwise use the configured chat model
	model := modelOverride
	if model == "" {
		model = llm.GetModelName(s.Config, "chat")
	}

	// Create the appropriate LLM client based on the model
	var client llm.LLMClient
	var err error

	// Determine which client to use based on the model name
	if strings.HasPrefix(model, "gpt-") {
		// OpenAI model
		client, err = llm.NewOpenAIClient(s.Config.Models.OpenAI)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
		}
	} else {
		// Ollama model (llama3:latest, sqlcoder:7b, etc.)
		client, err = llm.NewOllamaClient(s.Config.Models.Ollama)
		if err != nil {
			return nil, fmt.Errorf("failed to create Ollama client: %w", err)
		}
	}

	req := llm.ChatRequest{
		Model:    model,
		Messages: messages, // Pass messages exactly as provided - no system prompts added
		Stream:   false,
		Options: &api.Options{
			Temperature: 0.7,
			TopP:        0.9,
		},
	}

	return client.ChatCompletion(ctx, req)
}

// GenerateSQL generates SQL using SQLCoder model
func (s *AIService) GenerateSQL(prompt string, schema string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	model := llm.GetModelName(s.Config, "sql")

	// Create a comprehensive prompt for SQL generation using SQLCoder format
	fullPrompt := fmt.Sprintf(`-- Database: PostgreSQL
-- Schema:
%s
-- Task: %s

SELECT`, schema, prompt)

	req := llm.GenerateRequest{
		Model:  model,
		Prompt: fullPrompt,
		Stream: false,
		Options: &api.Options{
			Temperature: 0.1, // Lower temperature for more deterministic SQL
			TopP:        0.9,
		},
	}

	resp, err := s.sqlClient.GenerateText(ctx, req)
	if err != nil {
		return "", fmt.Errorf("SQL generation failed: %w", err)
	}

	return resp.Response, nil
}

// buildSQLCoderPromptFromIR converts IR into a natural language prompt for SQLCoder
func (s *AIService) buildSQLCoderPromptFromIR(ir map[string]interface{}, datasourceKind string) (string, error) {
	// Extract basic info from IR
	dataset, _ := ir["dataset"].(string)
	if dataset == "" {
		return "", fmt.Errorf("no dataset specified in IR")
	}

	var promptParts []string

	// Start with the main table
	promptParts = append(promptParts, fmt.Sprintf("Query the %s table", dataset))

	// Add SELECT fields description
	if selectFields, ok := ir["select"].([]interface{}); ok && len(selectFields) > 0 {
		var fieldDescriptions []string
		for _, field := range selectFields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				for alias, expr := range fieldMap {
					if exprStr, ok := expr.(string); ok {
						if alias != exprStr {
							fieldDescriptions = append(fieldDescriptions, fmt.Sprintf("%s (as %s)", exprStr, alias))
						} else {
							fieldDescriptions = append(fieldDescriptions, exprStr)
						}
					}
				}
			} else if fieldStr, ok := field.(string); ok {
				fieldDescriptions = append(fieldDescriptions, fieldStr)
			}
		}
		if len(fieldDescriptions) > 0 {
			promptParts = append(promptParts, fmt.Sprintf("selecting %s", strings.Join(fieldDescriptions, ", ")))
		}
	}

	// Add WHERE conditions description
	if whereConditions, ok := ir["where"].([]interface{}); ok && len(whereConditions) > 0 {
		var conditionDescriptions []string
		for _, condition := range whereConditions {
			if condMap, ok := condition.(map[string]interface{}); ok {
				field, _ := condMap["field"].(string)
				operator, _ := condMap["operator"].(string)
				value, _ := condMap["value"].(string)

				if field != "" && operator != "" && value != "" {
					// Handle placeholder values
					if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
						paramName := strings.Trim(value, "{{}}")
						conditionDescriptions = append(conditionDescriptions, fmt.Sprintf("%s %s %s", field, operator, paramName))
					} else {
						conditionDescriptions = append(conditionDescriptions, fmt.Sprintf("%s %s '%s'", field, operator, value))
					}
				}
			}
		}
		if len(conditionDescriptions) > 0 {
			promptParts = append(promptParts, fmt.Sprintf("where %s", strings.Join(conditionDescriptions, " AND ")))
		}
	}

	// Add GROUP BY description
	if groupByFields, ok := ir["group_by"].([]interface{}); ok && len(groupByFields) > 0 {
		var fields []string
		for _, field := range groupByFields {
			if fieldStr, ok := field.(string); ok {
				fields = append(fields, fieldStr)
			}
		}
		if len(fields) > 0 {
			promptParts = append(promptParts, fmt.Sprintf("grouped by %s", strings.Join(fields, ", ")))
		}
	}

	// Add ORDER BY description
	if orderByFields, ok := ir["order_by"].([]interface{}); ok && len(orderByFields) > 0 {
		var orderDescriptions []string
		for _, field := range orderByFields {
			if fieldMap, ok := field.(map[string]interface{}); ok {
				fieldName, _ := fieldMap["field"].(string)
				direction, _ := fieldMap["direction"].(string)

				if fieldName != "" {
					if direction != "" {
						orderDescriptions = append(orderDescriptions, fmt.Sprintf("%s %s", fieldName, strings.ToUpper(direction)))
					} else {
						orderDescriptions = append(orderDescriptions, fieldName)
					}
				}
			} else if fieldStr, ok := field.(string); ok {
				orderDescriptions = append(orderDescriptions, fieldStr)
			}
		}
		if len(orderDescriptions) > 0 {
			promptParts = append(promptParts, fmt.Sprintf("ordered by %s", strings.Join(orderDescriptions, ", ")))
		}
	}

	// Add LIMIT description
	if limit, ok := ir["limit"].(float64); ok && limit > 0 {
		promptParts = append(promptParts, fmt.Sprintf("limited to %d rows", int(limit)))
	}

	// Join all parts into a natural language description
	description := strings.Join(promptParts, ", ")

	// Add database-specific instructions
	switch strings.ToLower(datasourceKind) {
	case "sqlite", "sqlite3":
		description += ". Use SQLite syntax."
	case "postgres", "postgresql", "timescaledb":
		description += ". Use PostgreSQL syntax."
	case "mysql":
		description += ". Use MySQL syntax."
	}

	return description, nil
}

// getDatasourceSchema retrieves schema information for a datasource
func (s *AIService) getDatasourceSchema(datasourceID string) (string, error) {
	// Get datasource connector
	connector, err := s.registry.GetDatasource(datasourceID)
	if err != nil {
		return "", fmt.Errorf("datasource not found: %w", err)
	}

	// For now, return a basic schema description
	// In a full implementation, this would introspect the actual database schema
	schema := fmt.Sprintf(`-- Database: %s
-- Table structure will be provided by datasource learning
-- This is a placeholder schema that should be replaced with actual schema introspection`, connector.Kind)

	return schema, nil
}

// sanitizeModelJSONOutput removes common code fencing and yields raw JSON bytes
func sanitizeModelJSONOutput(content string) []byte {
	c := strings.TrimSpace(content)
	if strings.HasPrefix(c, "```") {
		// Strip leading fence line
		// Handle patterns like ```json\n...\n```
		// Remove first line
		newline := strings.IndexByte(c, '\n')
		if newline != -1 {
			c = c[newline+1:]
		}
		// Remove trailing fence
		if idx := strings.LastIndex(c, "```"); idx != -1 {
			c = c[:idx]
		}
	}
	return []byte(strings.TrimSpace(c))
}
