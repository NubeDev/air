package reports

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/services"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
)

// CreateScope creates a new scope
func CreateScope(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.CreateScopeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		scope, err := service.CreateScope(req)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create scope", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create scope",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, scope)
	}
}

// GetScope retrieves a scope by ID
func GetScope(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid scope ID",
				Details: err.Error(),
			})
			return
		}

		scope, err := service.GetScope(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error: "Scope not found",
			})
			return
		}

		c.JSON(http.StatusOK, scope)
	}
}

// CreateScopeVersion creates a new scope version
func CreateScopeVersion(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid scope ID",
				Details: err.Error(),
			})
			return
		}

		var req store.CreateScopeVersionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		version, err := service.CreateScopeVersion(uint(id), req)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create scope version", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create scope version",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, version)
	}
}

// CreateReport creates a new report
func CreateReport(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.CreateReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		report, err := service.CreateReport(req)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, report)
	}
}

// GetReport retrieves a report by key
func GetReport(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		report, err := service.GetReport(key)
		if err != nil {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error: "Report not found",
			})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// ListReports lists all reports
func ListReports(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		reports, err := service.ListReports()
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to list reports",
				Details: err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"reports": reports})
	}
}

// GetReportByID retrieves a report by numeric ID
func GetReportByID(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid report ID"})
			return
		}
		report, err := service.GetReportByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, store.ErrorResponse{Error: "Report not found"})
			return
		}
		c.JSON(http.StatusOK, report)
	}
}

// CreateReportVersion creates a new report version
func CreateReportVersion(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		var req store.CreateReportVersionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		version, err := service.CreateReportVersion(key, req)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create report version", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create report version",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, version)
	}
}

// CreateReportVersionByID creates a report version using report ID
func CreateReportVersionByID(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid report ID"})
			return
		}
		var req store.CreateReportVersionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid request", Details: err.Error()})
			return
		}
		report, err := service.GetReportByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, store.ErrorResponse{Error: "Report not found"})
			return
		}
		version, err := service.CreateReportVersion(report.Key, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{Error: "Failed to create report version", Details: err.Error()})
			return
		}
		c.JSON(http.StatusCreated, version)
	}
}

// RunReport executes a report
func RunReport(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		datasourceID := c.Query("datasource_id")

		var req store.RunReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		if datasourceID != "" {
			req.DatasourceID = &datasourceID
		}

		run, err := service.RunReport(key, req)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to run report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to run report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, run)
	}
}

// ExecuteReportByID runs a report by ID
func ExecuteReportByID(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid report ID"})
			return
		}
		datasourceID := c.Query("datasource_id")
		var req store.RunReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid request", Details: err.Error()})
			return
		}
		if datasourceID != "" {
			req.DatasourceID = &datasourceID
		}
		run, err := service.RunReportByID(uint(id), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{Error: "Failed to execute report", Details: err.Error()})
			return
		}
		c.JSON(http.StatusOK, run)
	}
}

// DeleteReportByID deletes a report by ID
func DeleteReportByID(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid report ID"})
			return
		}
		if err := service.DeleteReportByID(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{Error: "Failed to delete report", Details: err.Error()})
			return
		}
		c.JSON(http.StatusOK, store.SuccessResponse{Message: "Report deleted successfully"})
	}
}

// ExportReport exports a report
func ExportReport(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		format := c.DefaultQuery("format", "json")

		export, err := service.ExportReport(key, format)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to export report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to export report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, export)
	}
}

// GetReportData retrieves the latest execution data for a report
func GetReportData(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{Error: "Invalid report ID"})
			return
		}

		// Get the latest report run for this report
		run, err := service.GetLatestReportRun(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error:   "Report data not found",
				Details: err.Error(),
			})
			return
		}

		// Parse the results JSON
		var results []map[string]interface{}
		if run.Results != "" {
			if err := json.Unmarshal([]byte(run.Results), &results); err != nil {
				c.JSON(http.StatusInternalServerError, store.ErrorResponse{
					Error:   "Failed to parse report data",
					Details: err.Error(),
				})
				return
			}
		}

		// Return the data in a clean format
		response := map[string]interface{}{
			"report_id":    run.ReportID,
			"run_id":       run.ID,
			"status":       run.Status,
			"row_count":    run.RowCount,
			"data":         results,
			"executed_at":  run.StartedAt,
			"completed_at": run.FinishedAt,
			"sql":          run.SQLText,
		}

		c.JSON(http.StatusOK, response)
	}
}
