package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/logger"
	"github.com/ollama/ollama/api"
)

// OllamaClient handles communication with Ollama models
type OllamaClient struct {
	client *api.Client
	config config.OllamaConfig
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(cfg config.OllamaConfig) (*OllamaClient, error) {
	baseURL, err := url.Parse(cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("invalid Ollama host URL: %w", err)
	}

	client := api.NewClient(baseURL, &http.Client{})

	return &OllamaClient{
		client: client,
		config: cfg,
	}, nil
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model    string       `json:"model"`
	Messages []Message    `json:"messages"`
	Stream   bool         `json:"stream"`
	Options  *api.Options `json:"options,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	Model     string  `json:"model"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
	CreatedAt string  `json:"created_at"`
}

// GenerateRequest represents a text generation request
type GenerateRequest struct {
	Model   string       `json:"model"`
	Prompt  string       `json:"prompt"`
	Stream  bool         `json:"stream"`
	Options *api.Options `json:"options,omitempty"`
}

// GenerateResponse represents a text generation response
type GenerateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// ChatCompletion performs a chat completion using the specified model
func (c *OllamaClient) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	logger.LogInfo(logger.ServiceAI, "Starting chat completion", map[string]interface{}{
		"model":    req.Model,
		"messages": len(req.Messages),
	})

	// Convert our request to Ollama API format
	ollamaReq := api.ChatRequest{
		Model:   req.Model,
		Stream:  &req.Stream,
		Options: make(map[string]any),
	}

	// Convert options
	if req.Options != nil {
		ollamaReq.Options["temperature"] = req.Options.Temperature
		ollamaReq.Options["top_p"] = req.Options.TopP
	}

	// Convert messages
	for _, msg := range req.Messages {
		ollamaReq.Messages = append(ollamaReq.Messages, api.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Make the request
	var response ChatResponse
	err := c.client.Chat(ctx, &ollamaReq, func(resp api.ChatResponse) error {
		response = ChatResponse{
			Model:     resp.Model,
			Message:   Message{Role: resp.Message.Role, Content: resp.Message.Content},
			Done:      resp.Done,
			CreatedAt: resp.CreatedAt.Format(time.RFC3339),
		}
		return nil
	})

	if err != nil {
		logger.LogError(logger.ServiceAI, "Chat completion failed", err)
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	logger.LogInfo(logger.ServiceAI, "Chat completion completed", map[string]interface{}{
		"model": response.Model,
		"done":  response.Done,
	})

	return &response, nil
}

// GenerateText performs text generation using the specified model
func (c *OllamaClient) GenerateText(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	logger.LogInfo(logger.ServiceAI, "Starting text generation", map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt[:min(100, len(req.Prompt))] + "...",
	})

	// Convert our request to Ollama API format
	ollamaReq := api.GenerateRequest{
		Model:   req.Model,
		Prompt:  req.Prompt,
		Stream:  &req.Stream,
		Options: make(map[string]any),
	}

	// Convert options
	if req.Options != nil {
		ollamaReq.Options["temperature"] = req.Options.Temperature
		ollamaReq.Options["top_p"] = req.Options.TopP
	}

	// Make the request
	var response GenerateResponse
	err := c.client.Generate(ctx, &ollamaReq, func(resp api.GenerateResponse) error {
		response = GenerateResponse{
			Model:     resp.Model,
			Response:  resp.Response,
			Done:      resp.Done,
			CreatedAt: resp.CreatedAt.Format(time.RFC3339),
		}
		return nil
	})

	if err != nil {
		logger.LogError(logger.ServiceAI, "Text generation failed", err)
		return nil, fmt.Errorf("text generation failed: %w", err)
	}

	logger.LogInfo(logger.ServiceAI, "Text generation completed", map[string]interface{}{
		"model": response.Model,
		"done":  response.Done,
	})

	return &response, nil
}

// Health checks if the Ollama service is healthy
func (c *OllamaClient) Health(ctx context.Context) error {
	// Try to list models as a health check
	_, err := c.client.List(ctx)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Ollama health check failed", err)
		return fmt.Errorf("ollama health check failed: %w", err)
	}

	logger.LogInfo(logger.ServiceAI, "Ollama service is healthy")
	return nil
}

// ListModels lists available models
func (c *OllamaClient) ListModels(ctx context.Context) (*api.ListResponse, error) {
	models, err := c.client.List(ctx)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to list models", err)
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	logger.LogInfo(logger.ServiceAI, "Listed models", map[string]interface{}{
		"count": len(models.Models),
	})

	return models, nil
}

// GetModelInfo gets information about a specific model
func (c *OllamaClient) GetModelInfo(ctx context.Context, modelName string) (*api.ShowResponse, error) {
	req := api.ShowRequest{
		Name: modelName,
	}

	info, err := c.client.Show(ctx, &req)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to get model info", err)
		return nil, fmt.Errorf("failed to get model info for %s: %w", modelName, err)
	}

	logger.LogInfo(logger.ServiceAI, "Retrieved model info", map[string]interface{}{
		"model": modelName,
	})

	return info, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
