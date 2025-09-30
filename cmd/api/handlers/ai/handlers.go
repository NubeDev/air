package ai

import (
	"fmt"
	"net/http"

	"github.com/NubeDev/air/internal/services"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
)

// BuildIR builds Intermediate Representation from scope
func BuildIR(service *services.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.BuildIRRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		ir, err := service.BuildIR(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to build IR",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ir":               ir,
			"scope_version_id": req.ScopeVersionID,
		})
	}
}

// GenerateSQL generates SQL from IR for a specific datasource
func GenerateSQL(service *services.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.GenerateSQLRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		sql, safetyReport, err := service.GenerateSQL(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to generate SQL",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sql":           sql,
			"safety_report": safetyReport,
			"datasource_id": req.DatasourceID,
		})
	}
}

// AnalyzeRun analyzes a report run with AI
func AnalyzeRun(service *services.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		runIDStr := c.Param("run_id")
		var runID uint
		if _, err := fmt.Sscanf(runIDStr, "%d", &runID); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error: "Invalid run ID",
			})
			return
		}

		var req store.AnalyzeRunRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		analysis, err := service.AnalyzeRun(runID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to analyze run",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, analysis)
	}
}

// GetAITools returns available AI tools
func GetAITools(service *services.AIService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tools, err := service.GetAITools()
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get AI tools",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tools": tools,
		})
	}
}
