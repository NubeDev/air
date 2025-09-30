package llm

import (
	"context"
	"fmt"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/logger"
)

// LLMClient interface for different LLM providers
type LLMClient interface {
	ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	GenerateText(ctx context.Context, req GenerateRequest) (*GenerateResponse, error)
	Health(ctx context.Context) error
	ListModels(ctx context.Context) (*ModelsResponse, error)
	GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error)
}

// NewLLMClient creates the appropriate LLM client based on config
func NewLLMClient(cfg *config.Config) (LLMClient, error) {
	// Check if OpenAI is configured and should be used
	if cfg.Models.ChatPrimary == "openai" && cfg.Models.OpenAI.APIKey != "" {
		logger.LogInfo(logger.ServiceAI, "Using OpenAI as primary chat model", map[string]interface{}{
			"model": cfg.Models.OpenAI.Model,
		})
		return NewOpenAIClient(cfg.Models.OpenAI)
	}

	// Fall back to Ollama
	logger.LogInfo(logger.ServiceAI, "Using Ollama as chat model", map[string]interface{}{
		"model": cfg.Models.Ollama.Llama3Model,
	})
	return NewOllamaClient(cfg.Models.Ollama)
}

// NewSQLClient creates the appropriate client for SQL generation
func NewSQLClient(cfg *config.Config) (LLMClient, error) {
	// For SQL generation, we might want to use a different model
	// For now, use the same logic as chat
	return NewLLMClient(cfg)
}

// GetModelName returns the appropriate model name based on config and type
func GetModelName(cfg *config.Config, modelType string) string {
	switch modelType {
	case "chat":
		if cfg.Models.ChatPrimary == "openai" && cfg.Models.OpenAI.APIKey != "" {
			return cfg.Models.OpenAI.Model
		}
		return cfg.Models.Ollama.Llama3Model
	case "sql":
		if cfg.Models.SQLPrimary == "openai" && cfg.Models.OpenAI.APIKey != "" {
			return cfg.Models.OpenAI.Model
		}
		return cfg.Models.Ollama.SQLCoderModel
	default:
		// Default to chat model
		if cfg.Models.ChatPrimary == "openai" && cfg.Models.OpenAI.APIKey != "" {
			return cfg.Models.OpenAI.Model
		}
		return cfg.Models.Ollama.Llama3Model
	}
}

// CheckModelHealth checks if the specified model is available and healthy
func CheckModelHealth(cfg *config.Config, modelType string) error {
	client, err := NewLLMClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		return fmt.Errorf("model health check failed: %w", err)
	}

	// Get model name
	modelName := GetModelName(cfg, modelType)

	// Try to get model info to verify it exists
	_, err = client.GetModelInfo(ctx, modelName)
	if err != nil {
		return fmt.Errorf("model %s not available: %w", modelName, err)
	}

	logger.LogInfo(logger.ServiceAI, "Model health check passed", map[string]interface{}{
		"model": modelName,
		"type":  modelType,
	})

	return nil
}
