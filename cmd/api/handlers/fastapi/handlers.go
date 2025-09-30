package fastapi

import (
	"net/http"
	"time"

	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

type FastAPIHandler struct {
	client *services.FastAPIClient
}

func NewFastAPIHandler(baseURL string) *FastAPIHandler {
	return &FastAPIHandler{
		client: services.NewFastAPIClient(baseURL),
	}
}

// Health checks the FastAPI service health
func (h *FastAPIHandler) Health(c *gin.Context) {
	err := h.client.Health()
	if err != nil {
		logger.LogError(logger.ServiceAI, "FastAPI health check failed", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "FastAPI service unavailable",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "FastAPI data processing",
	})
}

// TestEnergyData tests the FastAPI service with energy data
func (h *FastAPIHandler) TestEnergyData(c *gin.Context) {
	logger.LogInfo(logger.ServiceAI, "Starting FastAPI energy data test")

	// 1. Test schema inference
	logger.LogInfo(logger.ServiceAI, "Step 1: Testing schema inference")
	schemaReq := services.InferSchemaRequest{
		DatasourceID: "energy",
		URI:          "../testdata/ts-energy.csv",
		InferRows:    50,
	}

	schemaToken, err := h.client.InferSchema(schemaReq)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Schema inference failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Schema inference failed",
			"details": err.Error(),
		})
		return
	}

	// Wait for schema inference to complete
	schemaResult, err := h.client.WaitForJobCompletion(schemaToken.Token, 30*time.Second)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Schema inference job failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Schema inference job failed",
			"details": err.Error(),
		})
		return
	}

	// 2. Test data preview
	logger.LogInfo(logger.ServiceAI, "Step 2: Testing data preview")
	previewReq := services.PreviewRequest{
		DatasourceID: "energy",
		Path:         stringPtr("../testdata/ts-energy.csv"),
		Limit:        5,
	}

	previewToken, err := h.client.PreviewData(previewReq)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Data preview failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Data preview failed",
			"details": err.Error(),
		})
		return
	}

	// Wait for preview to complete
	previewResult, err := h.client.WaitForJobCompletion(previewToken.Token, 30*time.Second)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Data preview job failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Data preview job failed",
			"details": err.Error(),
		})
		return
	}

	// 3. Test data analysis
	logger.LogInfo(logger.ServiceAI, "Step 3: Testing data analysis")
	analyzeReq := services.AnalyzeRequest{
		DatasourceID: stringPtr("energy"),
		Plan: &services.QueryPlan{
			Dataset: "../testdata/ts-energy.csv",
			Select:  []string{"timestamp", "kwh", "voltage", "current", "temperature"},
		},
		JobKind: "eda",
		Options: map[string]interface{}{
			"correlations": true,
			"outliers":     true,
		},
	}

	analyzeToken, err := h.client.AnalyzeData(analyzeReq)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Data analysis failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Data analysis failed",
			"details": err.Error(),
		})
		return
	}

	// Wait for analysis to complete
	analyzeResult, err := h.client.WaitForJobCompletion(analyzeToken.Token, 30*time.Second)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Data analysis job failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Data analysis job failed",
			"details": err.Error(),
		})
		return
	}

	// Return comprehensive results
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "FastAPI energy data test completed successfully",
		"results": gin.H{
			"schema_inference": gin.H{
				"token":  schemaResult.Token,
				"status": schemaResult.Status,
				"data":   schemaResult.Data,
			},
			"data_preview": gin.H{
				"token":  previewResult.Token,
				"status": previewResult.Status,
				"data":   previewResult.Data,
			},
			"data_analysis": gin.H{
				"token":  analyzeResult.Token,
				"status": analyzeResult.Status,
				"data":   analyzeResult.Data,
			},
		},
	})
}

// TestDiscoverFiles tests file discovery
func (h *FastAPIHandler) TestDiscoverFiles(c *gin.Context) {
	logger.LogInfo(logger.ServiceAI, "Testing file discovery")

	discoverReq := services.DiscoverRequest{
		DatasourceID: "test",
		URI:          ".",
		Recurse:      false,
		MaxFiles:     intPtr(10),
	}

	token, err := h.client.DiscoverFiles(discoverReq)
	if err != nil {
		logger.LogError(logger.ServiceAI, "File discovery failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "File discovery failed",
			"details": err.Error(),
		})
		return
	}

	// Wait for discovery to complete
	result, err := h.client.WaitForJobCompletion(token.Token, 30*time.Second)
	if err != nil {
		logger.LogError(logger.ServiceAI, "File discovery job failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "File discovery job failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "File discovery completed",
		"token":   result.Token,
		"data":    result.Data,
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
