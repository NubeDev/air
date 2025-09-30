package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/logger"
)

// OpenAIClient handles communication with OpenAI models
type OpenAIClient struct {
	client  *http.Client
	config  config.OpenAIConfig
	baseURL string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(cfg config.OpenAIConfig) (*OpenAIClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	return &OpenAIClient{
		client:  &http.Client{Timeout: 60 * time.Second},
		config:  cfg,
		baseURL: "https://api.openai.com/v1",
	}, nil
}

// OpenAIRequest represents an OpenAI API request
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// OpenAIMessage represents an OpenAI message
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents an OpenAI API response
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatCompletion performs a chat completion using the specified model
func (c *OpenAIClient) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	logger.LogInfo(logger.ServiceAI, "Starting OpenAI chat completion", map[string]interface{}{
		"model":    req.Model,
		"messages": len(req.Messages),
	})

	// Convert our request to OpenAI API format
	openaiReq := OpenAIRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: 0.7,
		TopP:        0.9,
		MaxTokens:   4000,
		Stream:      req.Stream,
	}

	// Override with options if provided
	if req.Options != nil {
		openaiReq.Temperature = float64(req.Options.Temperature)
		openaiReq.TopP = float64(req.Options.TopP)
	}

	// Marshal request
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Make the request
	resp, err := c.client.Do(httpReq)
	if err != nil {
		logger.LogError(logger.ServiceAI, "OpenAI request failed", err)
		return nil, fmt.Errorf("OpenAI request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errorResp struct {
			Error struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	// Parse response
	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to our response format
	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	choice := openaiResp.Choices[0]
	response := &ChatResponse{
		Model: openaiResp.Model,
		Message: Message{
			Role:    choice.Message.Role,
			Content: choice.Message.Content,
		},
		Done:      choice.FinishReason == "stop",
		CreatedAt: time.Unix(openaiResp.Created, 0).Format(time.RFC3339),
	}

	logger.LogInfo(logger.ServiceAI, "OpenAI chat completion completed", map[string]interface{}{
		"model":  response.Model,
		"done":   response.Done,
		"tokens": openaiResp.Usage.TotalTokens,
	})

	return response, nil
}

// GenerateText performs text generation using the specified model
func (c *OpenAIClient) GenerateText(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	logger.LogInfo(logger.ServiceAI, "Starting OpenAI text generation", map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt[:min(100, len(req.Prompt))] + "...",
	})

	// For text generation, we'll use chat completion with a single user message
	chatReq := ChatRequest{
		Model: req.Model,
		Messages: []Message{
			{Role: "user", Content: req.Prompt},
		},
		Stream:  req.Stream,
		Options: req.Options,
	}

	chatResp, err := c.ChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	response := &GenerateResponse{
		Model:     chatResp.Model,
		Response:  chatResp.Message.Content,
		Done:      chatResp.Done,
		CreatedAt: chatResp.CreatedAt,
	}

	logger.LogInfo(logger.ServiceAI, "OpenAI text generation completed", map[string]interface{}{
		"model": response.Model,
		"done":  response.Done,
	})

	return response, nil
}

// Health checks if the OpenAI service is healthy
func (c *OpenAIClient) Health(ctx context.Context) error {
	// Try to list models as a health check
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		logger.LogError(logger.ServiceAI, "OpenAI health check failed", err)
		return fmt.Errorf("OpenAI health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	logger.LogInfo(logger.ServiceAI, "OpenAI service is healthy")
	return nil
}

// ListModels lists available models
func (c *OpenAIClient) ListModels(ctx context.Context) (*ModelsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list models request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		logger.LogError(logger.ServiceAI, "Failed to list models", err)
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d", resp.StatusCode)
	}

	var openaiResp struct {
		Data []struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	// Convert to our format
	models := make([]Model, 0, len(openaiResp.Data))
	for _, model := range openaiResp.Data {
		models = append(models, Model{
			Name:       model.ID,
			Size:       0, // OpenAI doesn't provide size info
			ModifiedAt: time.Unix(model.Created, 0).Format(time.RFC3339),
		})
	}

	response := &ModelsResponse{
		Models: models,
	}

	logger.LogInfo(logger.ServiceAI, "Listed OpenAI models", map[string]interface{}{
		"count": len(models),
	})

	return response, nil
}

// GetModelInfo gets information about a specific model
func (c *OpenAIClient) GetModelInfo(ctx context.Context, modelName string) (*ModelInfo, error) {
	// OpenAI doesn't have a specific model info endpoint, so we'll return basic info
	return &ModelInfo{
		Name:       modelName,
		Size:       0,
		ModifiedAt: time.Now().Format(time.RFC3339),
	}, nil
}
