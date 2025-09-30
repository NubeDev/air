package reports

import (
	"fmt"
	"net/http"

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
		var id uint
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error: "Invalid scope ID",
			})
			return
		}

		scope, err := service.GetScope(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get scope",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, scope)
	}
}

// CreateScopeVersion creates a new version of a scope
func CreateScopeVersion(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		var id uint
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error: "Invalid scope ID",
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

		version, err := service.CreateScopeVersion(id, req)
		if err != nil {
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
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, report)
	}
}

// CreateReportVersion creates a new version of a report
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
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create report version",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, version)
	}
}

// RunReport executes a report with parameters
func RunReport(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		var req store.RunReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		run, err := service.RunReport(key, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to run report",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, run)
	}
}

// ExportReport exports a report in various formats
func ExportReport(service *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Param("key")
		format := c.DefaultQuery("format", "json")

		data, err := service.ExportReport(key, format)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to export report",
				Details: err.Error(),
			})
			return
		}

		c.Data(http.StatusOK, "application/json", data)
	}
}
