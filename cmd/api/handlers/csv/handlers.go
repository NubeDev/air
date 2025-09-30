package csv

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// ImportCSVRequest represents the request to import CSV data
type ImportCSVRequest struct {
	FilePath     string `json:"file_path" binding:"required"`
	TableName    string `json:"table_name" binding:"required"`
	DatasourceID string `json:"datasource_id" binding:"required"`
	HasHeader    bool   `json:"has_header"`
	Delimiter    string `json:"delimiter"`
	QuoteChar    string `json:"quote_char"`
	CreateTable  bool   `json:"create_table"`
	ReplaceData  bool   `json:"replace_data"`
}

// ImportCSVResponse represents the response from CSV import
type ImportCSVResponse struct {
	Status       string   `json:"status"`
	Message      string   `json:"message"`
	TableName    string   `json:"table_name"`
	RowsImported int      `json:"rows_imported"`
	Columns      []string `json:"columns"`
	ImportTime   string   `json:"import_time"`
}

// ImportCSV imports CSV data into a database table
func ImportCSV(registry *datasource.Registry, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ImportCSVRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		// Set defaults
		if req.Delimiter == "" {
			req.Delimiter = ","
		}
		if req.QuoteChar == "" {
			req.QuoteChar = "\""
		}

		// Get datasource connection
		connector, err := registry.GetDatasource(req.DatasourceID)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get datasource connector", err)
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error:   "Datasource not found",
				Details: err.Error(),
			})
			return
		}

		// Get DSN from database
		var datasource store.Datasource
		if err := db.Where("id = ?", req.DatasourceID).First(&datasource).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get datasource DSN", err)
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error:   "Datasource DSN not found",
				Details: err.Error(),
			})
			return
		}

		// Import CSV data
		result, err := importCSVToDatabase(connector, datasource.DSN, req)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to import CSV", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to import CSV",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// importCSVToDatabase performs the actual CSV import
func importCSVToDatabase(connector *datasource.DatasourceConnector, dsn string, req ImportCSVRequest) (*ImportCSVResponse, error) {
	// Open CSV file
	file, err := os.Open(req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	reader.Comma = rune(req.Delimiter[0])
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Clean column names (remove spaces, special chars, make lowercase)
	cleanColumns := make([]string, len(header))
	for i, col := range header {
		cleanColumns[i] = cleanColumnName(col)
	}

	// Map connector kind to driver name
	driverName := connector.Kind
	if connector.Kind == "sqlite" {
		driverName = "sqlite3"
	}

	// Get database connection
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create table if requested
	if req.CreateTable {
		if err := createTableFromCSV(db, req.TableName, cleanColumns, reader); err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	// Replace data if requested
	if req.ReplaceData {
		_, err = db.Exec(fmt.Sprintf("DELETE FROM %s", req.TableName))
		if err != nil {
			return nil, fmt.Errorf("failed to clear table: %w", err)
		}
	}

	// Prepare insert statement
	placeholders := make([]string, len(cleanColumns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		req.TableName,
		strings.Join(cleanColumns, ", "),
		strings.Join(placeholders, ", "))

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Import data rows
	rowsImported := 0
	startTime := time.Now()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV record: %w", err)
		}

		// Convert record to interface{} slice for prepared statement
		values := make([]interface{}, len(record))
		for i, val := range record {
			values[i] = val
		}

		// Insert row
		_, err = stmt.Exec(values...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert row %d: %w", rowsImported+1, err)
		}

		rowsImported++
	}

	importTime := time.Since(startTime)

	return &ImportCSVResponse{
		Status:       "success",
		Message:      fmt.Sprintf("Successfully imported %d rows", rowsImported),
		TableName:    req.TableName,
		RowsImported: rowsImported,
		Columns:      cleanColumns,
		ImportTime:   importTime.String(),
	}, nil
}

// createTableFromCSV creates a table based on CSV structure
func createTableFromCSV(db *sql.DB, tableName string, columns []string, reader *csv.Reader) error {
	// Read a few sample rows to infer data types
	sampleRows := make([][]string, 0, 10)
	for i := 0; i < 10; i++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read sample row: %w", err)
		}
		sampleRows = append(sampleRows, record)
	}

	// Infer column types
	columnTypes := make([]string, len(columns))
	for i := range columns {
		columnTypes[i] = inferColumnType(sampleRows, i)
	}

	// Build CREATE TABLE statement
	columnDefs := make([]string, len(columns))
	for i := range columns {
		columnDefs[i] = fmt.Sprintf("%s %s", columns[i], columnTypes[i])
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)",
		tableName,
		strings.Join(columnDefs, ", "))

	_, err := db.Exec(createSQL)
	return err
}

// cleanColumnName cleans a column name for database use
func cleanColumnName(name string) string {
	// Remove quotes and spaces
	name = strings.Trim(name, `"' `)
	// Replace spaces and special chars with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	// Convert to lowercase
	name = strings.ToLower(name)
	// Ensure it starts with a letter
	if len(name) > 0 && !isLetter(name[0]) {
		name = "col_" + name
	}
	return name
}

// isLetter checks if a character is a letter
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// inferColumnType infers the SQL data type for a column
func inferColumnType(sampleRows [][]string, columnIndex int) string {
	if len(sampleRows) == 0 {
		return "TEXT"
	}

	hasNumbers := false
	hasDecimals := false
	hasDates := false

	for _, row := range sampleRows {
		if columnIndex >= len(row) {
			continue
		}
		value := strings.TrimSpace(row[columnIndex])
		if value == "" {
			continue
		}

		// Check for numbers
		if _, err := strconv.Atoi(value); err == nil {
			hasNumbers = true
			continue
		}

		// Check for decimals
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			hasNumbers = true
			hasDecimals = true
			continue
		}

		// Check for dates (basic patterns)
		if isDateLike(value) {
			hasDates = true
			continue
		}
	}

	// Return appropriate type
	if hasDates {
		return "TEXT" // SQLite doesn't have native date type
	} else if hasDecimals {
		return "REAL"
	} else if hasNumbers {
		return "INTEGER"
	} else {
		return "TEXT"
	}
}

// isDateLike checks if a string looks like a date
func isDateLike(value string) bool {
	// Basic date patterns
	datePatterns := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, pattern := range datePatterns {
		if _, err := time.Parse(pattern, value); err == nil {
			return true
		}
	}
	return false
}
