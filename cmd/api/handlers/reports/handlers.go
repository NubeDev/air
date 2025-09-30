package reports

import (
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
