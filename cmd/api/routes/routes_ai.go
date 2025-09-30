package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/ai"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupIRRoutes configures IR (Intermediate Representation) routes
func SetupIRRoutes(rg *gin.RouterGroup, service *services.AIService, authMiddleware gin.HandlerFunc) {
	ir := rg.Group("/ir")
	ir.Use(authMiddleware)
	{
		ir.POST("/build", ai.BuildIR(service))
	}
}

// SetupSQLRoutes configures SQL generation routes
func SetupSQLRoutes(rg *gin.RouterGroup, service *services.AIService, authMiddleware gin.HandlerFunc) {
	sql := rg.Group("/sql")
	sql.Use(authMiddleware)
	{
		sql.POST("", ai.GenerateSQL(service))
	}
}

// SetupAnalysisRoutes configures analysis routes
func SetupAnalysisRoutes(rg *gin.RouterGroup, service *services.AIService, authMiddleware gin.HandlerFunc) {
	analysis := rg.Group("/runs")
	analysis.Use(authMiddleware)
	{
		analysis.POST("/:run_id/analyze", ai.AnalyzeRun(service))
	}
}

// SetupAIToolsRoutes configures AI tools routes
func SetupAIToolsRoutes(rg *gin.RouterGroup, service *services.AIService, authMiddleware gin.HandlerFunc) {
	rg.GET("/ai/tools", ai.GetAITools(service))
}
