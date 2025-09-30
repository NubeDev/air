package generated_reports

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListReports lists all generated reports
func ListReports(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reports []store.GeneratedReport
		if err := db.Where("status = ?", "active").Order("created_at DESC").Find(&reports).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to list reports", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to list reports",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"reports": reports,
		})
	}
}

// GetReport retrieves a report by ID
func GetReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid report ID",
				Details: err.Error(),
			})
			return
		}

		var report store.GeneratedReport
		if err := db.Preload("Session").First(&report, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, store.ErrorResponse{
					Error: "Report not found",
				})
				return
			}
			logger.LogError(logger.ServiceREST, "Failed to get report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// CreateReport creates a new generated report
func CreateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.CreateGeneratedReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		// Create report
		report := store.GeneratedReport{
			Name:           req.Name,
			Description:    req.Description,
			FilePath:       req.FilePath,
			ScopeJSON:      req.ScopeJSON,
			QueryPlanJSON:  req.QueryPlanJSON,
			ParametersJSON: req.ParametersJSON,
			SessionID:      req.SessionID,
			Status:         "active",
		}

		if err := db.Create(&report).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to create report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create report",
				Details: err.Error(),
			})
			return
		}

		logger.LogInfo(logger.ServiceREST, "Report created", map[string]interface{}{
			"report_id": report.ID,
			"name":      report.Name,
		})

		c.JSON(http.StatusCreated, report)
	}
}

// UpdateReport updates an existing report
func UpdateReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid report ID",
				Details: err.Error(),
			})
			return
		}

		var req struct {
			Name        string `json:"name,omitempty"`
			Description string `json:"description,omitempty"`
			Status      string `json:"status,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		// Update report
		updates := make(map[string]interface{})
		if req.Name != "" {
			updates["name"] = req.Name
		}
		if req.Description != "" {
			updates["description"] = req.Description
		}
		if req.Status != "" {
			updates["status"] = req.Status
		}
		updates["updated_at"] = time.Now()

		result := db.Model(&store.GeneratedReport{}).Where("id = ?", uint(id)).Updates(updates)
		if result.Error != nil {
			logger.LogError(logger.ServiceREST, "Failed to update report", result.Error)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to update report",
				Details: result.Error.Error(),
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error: "Report not found",
			})
			return
		}

		// Get updated report
		var report store.GeneratedReport
		if err := db.First(&report, uint(id)).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get updated report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get updated report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// DeleteReport deletes a report
func DeleteReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid report ID",
				Details: err.Error(),
			})
			return
		}

		// Soft delete by setting status to archived
		result := db.Model(&store.GeneratedReport{}).Where("id = ?", uint(id)).Update("status", "archived")
		if result.Error != nil {
			logger.LogError(logger.ServiceREST, "Failed to delete report", result.Error)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to delete report",
				Details: result.Error.Error(),
			})
			return
		}

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error: "Report not found",
			})
			return
		}

		logger.LogInfo(logger.ServiceREST, "Report deleted", map[string]interface{}{
			"report_id": id,
		})

		c.JSON(http.StatusOK, store.SuccessResponse{
			Message: "Report deleted successfully",
		})
	}
}

// ExecuteReport executes a generated report with parameters
func ExecuteReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid report ID",
				Details: err.Error(),
			})
			return
		}

		var req store.ExecuteReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		// Get report
		var report store.GeneratedReport
		if err := db.First(&report, uint(id)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, store.ErrorResponse{
					Error: "Report not found",
				})
				return
			}
			logger.LogError(logger.ServiceREST, "Failed to get report", err)
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get report",
				Details: err.Error(),
			})
			return
		}

		// Simulate execution (in real implementation, this would call Python backend)
		startTime := time.Now()

		// For now, return a simulated result
		// In the real implementation, this would:
		// 1. Parse the query plan from QueryPlanJSON
		// 2. Call Python backend with the query and parameters
		// 3. Return the actual results

		simulatedResults := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"site":      "site_a",
					"date":      "2024-01-01",
					"total_kwh": 1250.5,
					"avg_kwh":   52.1,
					"max_kwh":   75.2,
				},
				{
					"site":      "site_b",
					"date":      "2024-01-01",
					"total_kwh": 980.2,
					"avg_kwh":   40.8,
					"max_kwh":   65.4,
				},
			},
			"metadata": map[string]interface{}{
				"row_count":       2,
				"execution_time":  "0.15s",
				"parameters_used": req.Parameters,
			},
		}

		resultsJSON, _ := json.Marshal(simulatedResults)
		parametersJSON, _ := json.Marshal(req.Parameters)
		executionTime := time.Since(startTime)

		// Record execution
		execution := store.ReportExecution{
			ReportID:        report.ID,
			ParametersJSON:  string(parametersJSON),
			ResultsJSON:     string(resultsJSON),
			ExecutedAt:      time.Now(),
			ExecutionTimeMs: int(executionTime.Milliseconds()),
			Status:          "completed",
		}

		if err := db.Create(&execution).Error; err != nil {
			logger.LogError(logger.ServiceREST, "Failed to record execution", err)
			// Don't fail the request, just log the error
		}

		logger.LogInfo(logger.ServiceREST, "Report executed", map[string]interface{}{
			"report_id":      report.ID,
			"execution_time": executionTime.Milliseconds(),
		})

		c.JSON(http.StatusOK, simulatedResults)
	}
}
