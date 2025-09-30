package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/reports"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupScopeRoutes configures scope management routes
func SetupScopeRoutes(rg *gin.RouterGroup, service *services.ReportsService, authMiddleware gin.HandlerFunc) {
	scopes := rg.Group("/scopes")
	scopes.Use(authMiddleware)
	{
		scopes.POST("", reports.CreateScope(service))
		scopes.GET("/:id", reports.GetScope(service))
		scopes.POST("/:id/version", reports.CreateScopeVersion(service))
	}
}

// SetupReportRoutes configures report management routes
func SetupReportRoutes(rg *gin.RouterGroup, service *services.ReportsService, authMiddleware gin.HandlerFunc) {
	reportsGroup := rg.Group("/reports")
	reportsGroup.Use(authMiddleware)
	{
		reportsGroup.POST("", reports.CreateReport(service))
		reportsGroup.GET("/:key", reports.GetReport(service))
		reportsGroup.POST("/:key/versions", reports.CreateReportVersion(service))
		reportsGroup.POST("/:key/run", reports.RunReport(service))
		reportsGroup.GET("/:key/export", reports.ExportReport(service))
	}
}
