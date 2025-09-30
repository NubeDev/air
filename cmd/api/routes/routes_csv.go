package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/csv"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupCSVRoutes configures CSV import routes
func SetupCSVRoutes(rg *gin.RouterGroup, registry *datasource.Registry, db *gorm.DB, authMiddleware gin.HandlerFunc) {
	csvGroup := rg.Group("/csv")
	csvGroup.Use(authMiddleware)
	{
		csvGroup.POST("/import", csv.ImportCSV(registry, db))
	}
}
