package services

import (
	"context"
	"fmt"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/datasource"
	"github.com/NubeDev/air/internal/llm"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/store"
	"github.com/ollama/ollama/api"
	"gorm.io/gorm"
)

// AIService handles AI-related business logic
type AIService struct {
	registry     *datasource.Registry
	db           *gorm.DB
	ollamaClient *llm.OllamaClient
	config       *config.Config
}

// NewAIService creates a new AI service
func NewAIService(registry *datasource.Registry, db *gorm.DB, cfg *config.Config) (*AIService, error) {
	// Initialize Ollama client
	ollamaClient, err := llm.NewOllamaClient(cfg.Models.Ollama)
	if err != nil {
		return nil, fmt.Errorf("failed to create Ollama client: %w", err)
	}

	return &AIService{
		registry:     registry,
		db:           db,
		ollamaClient: ollamaClient,
		config:       cfg,
	}, nil
}

// BuildIR builds Intermediate Representation from scope
func (s *AIService) BuildIR(req store.BuildIRRequest) (map[string]interface{}, error) {
	start := time.Now()

	logger.LogInfo(logger.ServiceAI, "Building Intermediate Representation", map[string]interface{}{
		"scope_version_id": req.ScopeVersionID,
	})

	// TODO: Implement IR building logic
	result := map[string]interface{}{
		"status":  "not_implemented",
		"message": "IR building not yet implemented",
	}

	duration := time.Since(start)
	logger.LogInfo(logger.ServiceAI, "IR building completed (not implemented)", map[string]interface{}{
		"scope_version_id": req.ScopeVersionID,
		"duration":         duration.String(),
	})

	return result, fmt.Errorf("not implemented")
}

// GenerateSQLFromIR generates SQL from IR for a specific datasource
func (s *AIService) GenerateSQLFromIR(req store.GenerateSQLRequest) (string, map[string]interface{}, error) {
	// TODO: Implement SQL generation logic
	return "", nil, fmt.Errorf("not implemented")
}

// AnalyzeRun analyzes a report run with AI
func (s *AIService) AnalyzeRun(runID uint, req store.AnalyzeRunRequest) (*store.ReportAnalysis, error) {
	// TODO: Implement AI analysis logic
	return nil, fmt.Errorf("not implemented")
}

// GetAITools returns available AI tools
func (s *AIService) GetAITools() ([]map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check Ollama health first
	if err := s.ollamaClient.Health(ctx); err != nil {
		return nil, fmt.Errorf("ollama service unavailable: %w", err)
	}

	// List available models
	models, err := s.ollamaClient.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	tools := []map[string]interface{}{
		{
			"name":        "chat_completion",
			"description": "Generate chat completions using available models",
			"models":      []string{s.config.Models.Ollama.Llama3Model},
			"type":        "chat",
		},
		{
			"name":        "sql_generation",
			"description": "Generate SQL queries using SQLCoder model",
			"models":      []string{s.config.Models.Ollama.SQLCoderModel},
			"type":        "sql",
		},
		{
			"name":        "text_generation",
			"description": "Generate text using available models",
			"models":      []string{s.config.Models.Ollama.Llama3Model},
			"type":        "generate",
		},
	}

	// Add model information
	modelInfo := make([]map[string]interface{}, 0, len(models.Models))
	for _, model := range models.Models {
		modelInfo = append(modelInfo, map[string]interface{}{
			"name":     model.Name,
			"size":     model.Size,
			"modified": model.ModifiedAt,
		})
	}

	return append(tools, map[string]interface{}{
		"name":        "available_models",
		"description": "List of available Ollama models",
		"models":      modelInfo,
		"type":        "info",
	}), nil
}

// ChatCompletion performs a chat completion using the configured model
func (s *AIService) ChatCompletion(messages []llm.Message) (*llm.ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	model := s.config.Models.Ollama.Llama3Model
	if model == "" {
		model = "llama3"
	}

	req := llm.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options: &api.Options{
			Temperature: 0.7,
			TopP:        0.9,
		},
	}

	return s.ollamaClient.ChatCompletion(ctx, req)
}

// GenerateSQL generates SQL using SQLCoder model
func (s *AIService) GenerateSQL(prompt string, schema string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	model := s.config.Models.Ollama.SQLCoderModel
	if model == "" {
		model = "sqlcoder"
	}

	// Create a comprehensive prompt for SQL generation using SQLCoder format
	fullPrompt := fmt.Sprintf(`-- Database: PostgreSQL
-- Schema:
%s
-- Task: %s

SELECT`, schema, prompt)

	req := llm.GenerateRequest{
		Model:  model,
		Prompt: fullPrompt,
		Stream: false,
		Options: &api.Options{
			Temperature: 0.1, // Lower temperature for more deterministic SQL
			TopP:        0.9,
		},
	}

	resp, err := s.ollamaClient.GenerateText(ctx, req)
	if err != nil {
		return "", fmt.Errorf("SQL generation failed: %w", err)
	}

	return resp.Response, nil
}
