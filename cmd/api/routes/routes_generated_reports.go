package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/generated_reports"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupGeneratedReportRoutes configures generated report routes
func SetupGeneratedReportRoutes(rg *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	generated := rg.Group("/generated")
	generated.Use(authMiddleware)
	{
		generated.GET("/reports", generated_reports.ListReports(db))
		generated.POST("/reports", generated_reports.CreateReport(db))
		generated.GET("/reports/:id", generated_reports.GetReport(db))
		generated.PUT("/reports/:id", generated_reports.UpdateReport(db))
		generated.DELETE("/reports/:id", generated_reports.DeleteReport(db))
		generated.POST("/reports/:id/execute", generated_reports.ExecuteReport(db))
	}
}
