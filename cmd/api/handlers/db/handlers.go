package db

import (
	"net/http"

	"github.com/NubeDev/air/internal/services"
	"github.com/NubeDev/air/internal/store"
	"github.com/gin-gonic/gin"
)

// GetDatasources returns all registered datasources
func GetDatasources(service *services.DatasourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		datasources, err := service.ListDatasources()
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to list datasources",
				Details: err.Error(),
			})
			return
		}

		response := store.DatasourcesResponse{Datasources: datasources}
		c.JSON(http.StatusOK, response)
	}
}

// CreateDatasource creates a new datasource
func CreateDatasource(service *services.DatasourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.CreateDatasourceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		if err := service.CreateDatasource(req); err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to create datasource",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, store.SuccessResponse{
			Message: "Datasource created successfully",
		})
	}
}

// GetDatasourceHealth checks the health of a specific datasource
func GetDatasourceHealth(service *services.DatasourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		response, err := service.GetDatasourceHealth(id)
		if err != nil {
			c.JSON(http.StatusNotFound, store.ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, response)
	}
}

// DeleteDatasource removes a datasource
func DeleteDatasource(service *services.DatasourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := service.DeleteDatasource(id); err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to delete datasource",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, store.SuccessResponse{
			Message: "Datasource deleted successfully",
		})
	}
}

// LearnDatasource learns schema from a datasource
func LearnDatasource(service *services.DatasourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req store.LearnDatasourceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, store.ErrorResponse{
				Error:   "Invalid request",
				Details: err.Error(),
			})
			return
		}

		if err := service.LearnDatasource(req); err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to learn datasource",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, store.SuccessResponse{
			Message: "Learning started successfully",
		})
	}
}

// GetSchema returns schema information for a datasource
func GetSchema(service *services.DatasourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		datasourceID := c.Param("datasource_id")

		schema, err := service.GetSchema(datasourceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, store.ErrorResponse{
				Error:   "Failed to get schema",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"datasource_id": datasourceID,
			"schema_notes":  schema,
		})
	}
}
