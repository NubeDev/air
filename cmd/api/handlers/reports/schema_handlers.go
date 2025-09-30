package reports

import (
	"net/http"
	"strconv"

	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/services"
	"github.com/gin-gonic/gin"
)

// GetReportSchema generates JSON Schema for report parameters
func GetReportSchema(reportsService *services.ReportsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		reportIDStr := c.Param("id")
		reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
		if err != nil {
			logger.LogError(logger.ServiceREST, "Invalid report ID", err, map[string]interface{}{
				"report_id": reportIDStr,
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
			return
		}

		// Get the report to verify it exists
		_, err = reportsService.GetReportByID(uint(reportID))
		if err != nil {
			logger.LogError(logger.ServiceREST, "Failed to get report", err, map[string]interface{}{
				"report_id": reportID,
			})
			c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
			return
		}

		// For now, generate a simple schema based on common parameters
		// In production, this would parse the actual SQL from the report version
		schema := generateDefaultSchema()

		c.JSON(http.StatusOK, gin.H{
			"report_id": reportID,
			"schema":    schema,
		})
	}
}

// generateDefaultSchema creates a default JSON Schema for common parameters
func generateDefaultSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"customer_name": map[string]interface{}{
				"type":        "string",
				"title":       "Customer Name",
				"description": "Filter results by specific customer name",
			},
			"sales_rep_name": map[string]interface{}{
				"type":        "string",
				"title":       "Sales Representative",
				"description": "Filter results by sales representative",
			},
			"region": map[string]interface{}{
				"type":        "string",
				"title":       "Region",
				"description": "Filter results by geographic region",
			},
			"start_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"title":       "Start Date",
				"description": "Filter results from this date (YYYY-MM-DD)",
			},
			"end_date": map[string]interface{}{
				"type":        "string",
				"format":      "date",
				"title":       "End Date",
				"description": "Filter results to this date (YYYY-MM-DD)",
			},
			"min_sales_amount": map[string]interface{}{
				"type":        "number",
				"title":       "Minimum Sales Amount",
				"description": "Only show sales above this amount",
				"minimum":     0,
			},
			"max_sales_amount": map[string]interface{}{
				"type":        "number",
				"title":       "Maximum Sales Amount",
				"description": "Only show sales below this amount",
				"minimum":     0,
			},
			"product_category": map[string]interface{}{
				"type":        "string",
				"title":       "Product Category",
				"description": "Filter results by product category",
				"enum":        []string{"Electronics", "Clothing", "Home & Garden", "Books", "Sports", "Beauty"},
			},
			"payment_method": map[string]interface{}{
				"type":        "string",
				"title":       "Payment Method",
				"description": "Filter results by payment method used",
				"enum":        []string{"credit_card", "paypal", "bank_transfer", "cash", "check"},
			},
			"loyalty_tier": map[string]interface{}{
				"type":        "string",
				"title":       "Loyalty Tier",
				"description": "Filter results by customer loyalty tier",
				"enum":        []string{"bronze", "silver", "gold", "platinum"},
			},
		},
		"required": []string{"customer_name"},
	}
}

// generateParameterSchema creates JSON Schema from SQL placeholders
func generateParameterSchema(sql string) (map[string]interface{}, error) {
	// Extract parameters from SQL placeholders like {{param_name}}
	parameters := extractParametersFromSQL(sql)

	// Build JSON Schema
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}

	properties := make(map[string]interface{})
	required := []string{}

	for paramName, paramType := range parameters {
		property := map[string]interface{}{
			"title":       formatTitle(paramName),
			"description": generateDescription(paramName),
		}

		// Set type and format based on parameter name and type
		switch paramType {
		case "string":
			property["type"] = "string"
			if isDateParam(paramName) {
				property["format"] = "date"
			} else if isEmailParam(paramName) {
				property["format"] = "email"
			}
		case "number":
			property["type"] = "number"
		case "integer":
			property["type"] = "integer"
		case "boolean":
			property["type"] = "boolean"
		default:
			property["type"] = "string"
		}

		// Add enum values for common parameters
		if enumValues := getEnumValues(paramName); len(enumValues) > 0 {
			property["enum"] = enumValues
		}

		properties[paramName] = property
		required = append(required, paramName)
	}

	schema["properties"] = properties
	schema["required"] = required

	return schema, nil
}

// extractParametersFromSQL finds {{param}} placeholders in SQL
func extractParametersFromSQL(sql string) map[string]string {
	parameters := make(map[string]string)

	// Simple regex-like extraction of {{param}} patterns
	// This is a basic implementation - could be enhanced with proper regex
	// For now, we'll use a simple approach

	// Common parameter patterns based on typical usage
	commonParams := map[string]string{
		"customer_name":    "string",
		"sales_rep_name":   "string",
		"region":           "string",
		"product_category": "string",
		"start_date":       "string",
		"end_date":         "string",
		"min_sales_amount": "number",
		"max_sales_amount": "number",
		"limit":            "integer",
		"offset":           "integer",
		"status":           "string",
		"payment_method":   "string",
		"loyalty_tier":     "string",
		"warehouse_id":     "string",
	}

	// Check if SQL contains any of these common parameters
	for param, paramType := range commonParams {
		if containsPlaceholder(sql, param) {
			parameters[param] = paramType
		}
	}

	return parameters
}

// containsPlaceholder checks if SQL contains {{param}} placeholder
func containsPlaceholder(sql, param string) bool {
	placeholder := "{{" + param + "}}"
	return len(sql) > 0 && len(placeholder) <= len(sql) &&
		findSubstring(sql, placeholder) >= 0
}

// findSubstring finds substring position (simple implementation)
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// formatTitle converts snake_case to Title Case
func formatTitle(param string) string {
	// Simple implementation - could be enhanced
	switch param {
	case "customer_name":
		return "Customer Name"
	case "sales_rep_name":
		return "Sales Representative"
	case "start_date":
		return "Start Date"
	case "end_date":
		return "End Date"
	case "min_sales_amount":
		return "Minimum Sales Amount"
	case "max_sales_amount":
		return "Maximum Sales Amount"
	case "product_category":
		return "Product Category"
	case "payment_method":
		return "Payment Method"
	case "loyalty_tier":
		return "Loyalty Tier"
	case "warehouse_id":
		return "Warehouse ID"
	default:
		return param
	}
}

// generateDescription creates helpful descriptions for parameters
func generateDescription(param string) string {
	switch param {
	case "customer_name":
		return "Filter results by specific customer name"
	case "sales_rep_name":
		return "Filter results by sales representative"
	case "region":
		return "Filter results by geographic region"
	case "start_date":
		return "Filter results from this date (YYYY-MM-DD)"
	case "end_date":
		return "Filter results to this date (YYYY-MM-DD)"
	case "min_sales_amount":
		return "Only show sales above this amount"
	case "max_sales_amount":
		return "Only show sales below this amount"
	case "product_category":
		return "Filter results by product category"
	case "payment_method":
		return "Filter results by payment method used"
	case "loyalty_tier":
		return "Filter results by customer loyalty tier"
	case "warehouse_id":
		return "Filter results by warehouse location"
	default:
		return "Parameter for filtering results"
	}
}

// isDateParam checks if parameter is date-related
func isDateParam(param string) bool {
	dateParams := []string{"start_date", "end_date", "date", "created_at", "updated_at"}
	for _, dp := range dateParams {
		if param == dp {
			return true
		}
	}
	return false
}

// isEmailParam checks if parameter is email-related
func isEmailParam(param string) bool {
	return param == "email" || param == "customer_email"
}

// getEnumValues returns possible values for dropdown parameters
func getEnumValues(param string) []string {
	switch param {
	case "region":
		return []string{"North America", "Europe", "Asia", "South America", "Africa", "Oceania"}
	case "product_category":
		return []string{"Electronics", "Clothing", "Home & Garden", "Books", "Sports", "Beauty"}
	case "payment_method":
		return []string{"credit_card", "paypal", "bank_transfer", "cash", "check"}
	case "loyalty_tier":
		return []string{"bronze", "silver", "gold", "platinum"}
	case "status":
		return []string{"pending", "processing", "shipped", "delivered", "cancelled", "returned"}
	default:
		return []string{}
	}
}
