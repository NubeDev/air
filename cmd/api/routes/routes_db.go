package routes

import (
	"github.com/NubeDev/air/cmd/api/handlers/db"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupDatasourceRoutes configures datasource management routes
func SetupDatasourceRoutes(rg *gin.RouterGroup, service *services.DatasourceService, authMiddleware gin.HandlerFunc) {
	datasources := rg.Group("/datasources")
	datasources.Use(authMiddleware)
	{
		datasources.GET("", db.GetDatasources(service))
		datasources.POST("", db.CreateDatasource(service))
		datasources.GET("/:id/health", db.GetDatasourceHealth(service))
		datasources.DELETE("/:id", db.DeleteDatasource(service))
	}
}

// SetupLearnRoutes configures database learning routes
func SetupLearnRoutes(rg *gin.RouterGroup, service *services.DatasourceService, authMiddleware gin.HandlerFunc) {
	learn := rg.Group("/learn")
	learn.Use(authMiddleware)
	{
		learn.POST("", db.LearnDatasource(service))
	}
}

// SetupSchemaRoutes configures schema management routes
func SetupSchemaRoutes(rg *gin.RouterGroup, service *services.DatasourceService, authMiddleware gin.HandlerFunc) {
	schema := rg.Group("/schema")
	schema.Use(authMiddleware)
	{
		schema.GET("/:datasource_id", db.GetSchema(service))
	}
}
