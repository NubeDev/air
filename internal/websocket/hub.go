package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/NubeDev/air/internal/llm"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/redis"
	"github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	Channel   string                 `json:"channel"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID           string
	UserID       string
	Conn         *websocket.Conn
	Send         chan []byte
	Hub          *Hub
	Channels     map[string]bool // Subscribed channels
	selectedFile string          // Currently selected file for analysis
	mu           sync.RWMutex
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Channel subscriptions
	Channels map[string]map[*Client]bool

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Inbound messages from the clients
	Broadcast chan []byte

	// Channel-specific messages
	ChannelMessage chan ChannelMessage

	// Redis client for pub/sub
	Redis *redis.Client

	// AI service for chat responses
	AIService interface{}

	// Configuration
	Config *Config

	// Mutex for thread safety
	Mu sync.RWMutex
}

// ChannelMessage represents a message sent to a specific channel
type ChannelMessage struct {
	Channel string
	Message []byte
}

// Config holds WebSocket hub configuration
type Config struct {
	ReadBufferSize    int
	WriteBufferSize   int
	HandshakeTimeout  time.Duration
	PingPeriod        time.Duration
	PongWait          time.Duration
	MaxMessageSize    int64
	EnableCompression bool
}

// NewHub creates a new WebSocket hub
func NewHub(redisClient *redis.Client, config *Config, aiService interface{}) *Hub {
	hub := &Hub{
		Clients:        make(map[*Client]bool),
		Channels:       make(map[string]map[*Client]bool),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan []byte),
		ChannelMessage: make(chan ChannelMessage),
		Redis:          redisClient,
		Config:         config,
	}

	// Set AI service if it implements the required interface
	if aiService != nil {
		if ai, ok := aiService.(interface {
			ChatCompletion(messages []llm.Message) (*llm.ChatResponse, error)
		}); ok {
			hub.AIService = ai
		}
	}

	return hub
}

// Run starts the hub
func (h *Hub) Run(ctx context.Context) {
	logger.LogInfo(logger.ServiceWS, "Starting WebSocket hub")

	// Start Redis subscriber if Redis is available
	if h.Redis != nil {
		go h.runRedisSubscriber(ctx)
	}

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastToAll(message)

		case channelMsg := <-h.ChannelMessage:
			h.broadcastToChannel(channelMsg.Channel, channelMsg.Message)

		case <-ctx.Done():
			logger.LogInfo(logger.ServiceWS, "WebSocket hub shutting down")
			return
		}
	}
}

// runRedisSubscriber subscribes to Redis channels and forwards messages
func (h *Hub) runRedisSubscriber(ctx context.Context) {
	if h.Redis == nil {
		return
	}

	// Subscribe to all channels
	channels := []string{
		"websocket:chat:*",
		"websocket:ai:*",
		"websocket:typing:*",
		"websocket:presence:*",
		"websocket:system:*",
	}

	pubsub := h.Redis.Subscribe(ctx, channels...)
	defer pubsub.Close()

	logger.LogInfo(logger.ServiceWS, "Redis subscriber started", map[string]interface{}{
		"channels": channels,
	})

	for {
		select {
		case <-ctx.Done():
			logger.LogInfo(logger.ServiceWS, "Redis subscriber shutting down")
			return
		default:
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				logger.LogError(logger.ServiceWS, "Redis subscription error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Forward Redis message to WebSocket clients
			h.ChannelMessage <- ChannelMessage{
				Channel: msg.Channel,
				Message: []byte(msg.Payload),
			}
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	h.Clients[client] = true
	client.Channels = make(map[string]bool)

	logger.LogInfo(logger.ServiceWS, "Client registered", map[string]interface{}{
		"client_id":     client.ID,
		"user_id":       client.UserID,
		"total_clients": len(h.Clients),
	})

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// unregisterClient unregisters a client
func (h *Hub) unregisterClient(client *Client) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	if _, ok := h.Clients[client]; ok {
		// Remove from all channels
		for channel := range client.Channels {
			if channelClients, exists := h.Channels[channel]; exists {
				delete(channelClients, client)
				if len(channelClients) == 0 {
					delete(h.Channels, channel)
				}
			}
		}

		delete(h.Clients, client)
		close(client.Send)

		logger.LogInfo(logger.ServiceWS, "Client unregistered", map[string]interface{}{
			"client_id":     client.ID,
			"user_id":       client.UserID,
			"total_clients": len(h.Clients),
		})
	}
}

// broadcastToAll broadcasts a message to all connected clients
func (h *Hub) broadcastToAll(message []byte) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	for client := range h.Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.Clients, client)
		}
	}
}

// broadcastToChannel broadcasts a message to clients subscribed to a specific channel
func (h *Hub) broadcastToChannel(channel string, message []byte) {
	h.Mu.RLock()
	defer h.Mu.RUnlock()

	if channelClients, exists := h.Channels[channel]; exists {
		for client := range channelClients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.Clients, client)
			}
		}
	}
}

// SubscribeToChannel subscribes a client to a channel
func (h *Hub) SubscribeToChannel(client *Client, channel string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	client.mu.Lock()
	client.Channels[channel] = true
	client.mu.Unlock()

	if h.Channels[channel] == nil {
		h.Channels[channel] = make(map[*Client]bool)
	}
	h.Channels[channel][client] = true

	logger.LogDebug(logger.ServiceWS, "Client subscribed to channel", map[string]interface{}{
		"client_id": client.ID,
		"channel":   channel,
	})
}

// UnsubscribeFromChannel unsubscribes a client from a channel
func (h *Hub) UnsubscribeFromChannel(client *Client, channel string) {
	h.Mu.Lock()
	defer h.Mu.Unlock()

	client.mu.Lock()
	delete(client.Channels, channel)
	client.mu.Unlock()

	if channelClients, exists := h.Channels[channel]; exists {
		delete(channelClients, client)
		if len(channelClients) == 0 {
			delete(h.Channels, channel)
		}
	}

	logger.LogDebug(logger.ServiceWS, "Client unsubscribed from channel", map[string]interface{}{
		"client_id": client.ID,
		"channel":   channel,
	})
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID string, message Message) error {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.Mu.RLock()
	defer h.Mu.RUnlock()

	sent := false
	for client := range h.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- messageBytes:
				sent = true
			default:
				// Client channel is full, skip
			}
		}
	}

	if !sent {
		logger.LogWarn(logger.ServiceWS, "No active clients found for user", map[string]interface{}{
			"user_id": userID,
		})
	}

	return nil
}

// PublishToRedis publishes a message to Redis for distribution
func (h *Hub) PublishToRedis(ctx context.Context, channel string, message Message) error {
	if h.Redis == nil {
		return fmt.Errorf("redis client is not available")
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return h.Redis.Publish(ctx, channel, messageBytes)
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(c.Hub.Config.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Hub.Config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.Hub.Config.PongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.LogError(logger.ServiceWS, "WebSocket read error", err, map[string]interface{}{
					"client_id": c.ID,
				})
			}
			break
		}

		// Parse message
		var message Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			logger.LogError(logger.ServiceWS, "Failed to parse message", err, map[string]interface{}{
				"client_id": c.ID,
			})
			continue
		}

		// Handle different message types
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Hub.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage handles incoming messages from clients
func (c *Client) handleMessage(message Message) {
	switch message.Type {
	case "subscribe":
		if channel, ok := message.Payload["channel"].(string); ok {
			c.Hub.SubscribeToChannel(c, channel)
		}
	case "unsubscribe":
		if channel, ok := message.Payload["channel"].(string); ok {
			c.Hub.UnsubscribeFromChannel(c, channel)
		}
	case "ping":
		// Respond to ping with pong
		response := Message{
			Type:      "pong",
			Timestamp: time.Now(),
		}
		responseBytes, _ := json.Marshal(response)
		c.Send <- responseBytes
	case "file_analysis":
		// Handle file analysis request
		c.handleFileAnalysis(message)
	case "load_dataset":
		// Handle dataset loading
		c.handleLoadDataset(message)
	case "chat_message":
		// Handle chat message
		c.handleChatMessage(message)
	case "raw_ai_message":
		// Handle raw AI message (no system prompts)
		c.handleRawAIMessage(message)
	default:
		// Forward message to Redis for distribution
		message.UserID = c.UserID
		message.Timestamp = time.Now()

		// Determine Redis channel based on message type
		redisChannel := fmt.Sprintf("websocket:%s", message.Type)
		if message.Channel != "" {
			redisChannel = fmt.Sprintf("websocket:%s:%s", message.Type, message.Channel)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.Hub.PublishToRedis(ctx, redisChannel, message); err != nil {
			logger.LogError(logger.ServiceWS, "Failed to publish message to Redis", err, map[string]interface{}{
				"client_id": c.ID,
				"channel":   redisChannel,
			})
		}
	}
}

// handleFileAnalysis handles file analysis requests via WebSocket
func (c *Client) handleFileAnalysis(message Message) {
	// Extract file analysis parameters
	fileID, ok := message.Payload["file_id"].(string)
	if !ok {
		c.sendError("file_id is required")
		return
	}

	query, ok := message.Payload["query"].(string)
	if !ok {
		c.sendError("query is required")
		return
	}

	model, _ := message.Payload["model"].(string)
	if model == "" {
		model = "llama"
	}

	// Send analysis started message
	c.sendMessage(Message{
		Type: "file_analysis_started",
		Payload: map[string]interface{}{
			"file_id": fileID,
			"query":   query,
			"model":   model,
		},
		Timestamp: time.Now(),
	})

	// Perform file analysis with timeout
	go c.performFileAnalysisWithTimeout(fileID, query, model)
}

// performFileAnalysis performs the actual file analysis using real AI only
func (c *Client) performFileAnalysis(fileID, query, model string) {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	logger.LogInfo(logger.ServiceWS, "Starting AI file analysis", map[string]interface{}{
		"file_id": fileID,
		"query":   query,
		"model":   model,
	})

	// Use AI to analyze the file directly
	filePath := fmt.Sprintf("uploads/%s", fileID)

	analysis, insights, suggestions, err := c.analyzeFileWithAI(filePath, query, model)

	if err != nil {
		logger.LogError(logger.ServiceWS, "AI file analysis failed", err, map[string]interface{}{
			"file_path": filePath,
			"query":     query,
		})

		// Send error message
		if c.isConnected() {
			c.sendMessage(Message{
				Type: "file_analysis_error",
				Payload: map[string]interface{}{
					"file_id": fileID,
					"error":   fmt.Sprintf("Analysis failed: %v", err),
				},
				Timestamp: time.Now(),
			})
		}
		return
	}

	// Send analysis complete message
	if c.isConnected() {
		c.sendMessage(Message{
			Type: "file_analysis_complete",
			Payload: map[string]interface{}{
				"file_id":     fileID,
				"query":       query,
				"model":       model,
				"analysis":    analysis,
				"insights":    insights,
				"suggestions": suggestions,
			},
			Timestamp: time.Now(),
		})
	}

	logger.LogInfo(logger.ServiceWS, "AI file analysis completed successfully", map[string]interface{}{
		"file_id":           fileID,
		"insights_count":    len(insights),
		"suggestions_count": len(suggestions),
	})
}

// performFileAnalysisWithTimeout performs file analysis with timeout handling
func (c *Client) performFileAnalysisWithTimeout(fileID, query, model string) {
	done := make(chan bool, 1)

	go func() {
		c.performFileAnalysis(fileID, query, model)
		done <- true
	}()

	select {
	case <-done:
		// Analysis completed successfully
		return
	case <-time.After(60 * time.Second):
		// Timeout occurred
		logger.LogWarn(logger.ServiceWS, "File analysis timeout", map[string]interface{}{
			"file_id":         fileID,
			"timeout_seconds": 60,
		})
		if c.isConnected() {
			c.sendError("Analysis timed out after 60 seconds")
		}
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(message Message) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.LogError(logger.ServiceWS, "Failed to marshal message", err)
		return
	}

	logger.LogInfo(logger.ServiceWS, "Sending WebSocket message", map[string]interface{}{
		"type":         message.Type,
		"client_id":    c.ID,
		"message_size": len(messageBytes),
	})

	// Check if client is still connected before sending
	if !c.isConnected() {
		logger.LogWarn(logger.ServiceWS, "Client disconnected, skipping message", map[string]interface{}{
			"client_id": c.ID,
			"type":      message.Type,
		})
		return
	}

	// Try to send message with timeout
	select {
	case c.Send <- messageBytes:
		// Message sent successfully
	case <-time.After(1 * time.Second):
		// Channel is likely closed or blocked
		logger.LogWarn(logger.ServiceWS, "Failed to send message - channel closed or blocked", map[string]interface{}{
			"client_id": c.ID,
			"type":      message.Type,
		})
		return
	}
}

// isConnected checks if the client is still connected
func (c *Client) isConnected() bool {
	c.Hub.Mu.RLock()
	defer c.Hub.Mu.RUnlock()
	_, exists := c.Hub.Clients[c]
	return exists
}

// sendError sends an error message to the client
func (c *Client) sendError(errorMsg string) {
	c.sendMessage(Message{
		Type: "file_analysis_error",
		Payload: map[string]interface{}{
			"error": errorMsg,
		},
		Timestamp: time.Now(),
	})
}

// handleChatMessage handles chat messages via WebSocket
func (c *Client) handleChatMessage(message Message) {
	// Extract chat message parameters
	content, ok := message.Payload["content"].(string)
	if !ok {
		c.sendError("content is required")
		return
	}

	model, _ := message.Payload["model"].(string)
	if model == "" {
		model = "llama"
	}

	logger.LogInfo(logger.ServiceWS, "Processing chat message", map[string]interface{}{
		"content": content,
		"model":   model,
		"user_id": c.UserID,
	})

	// Send typing indicator
	c.sendMessage(Message{
		Type: "chat_typing",
		Payload: map[string]interface{}{
			"is_typing": true,
		},
		Timestamp: time.Now(),
	})

	// Process the chat message
	go c.processChatMessage(content, model)
}

// handleRawAIMessage handles raw AI messages via WebSocket
func (c *Client) handleRawAIMessage(message Message) {
	// Extract raw AI message parameters
	content, ok := message.Payload["content"].(string)
	if !ok {
		c.sendError("content is required")
		return
	}

	model, _ := message.Payload["model"].(string)
	if model == "" {
		model = "gpt-4o-mini" // Default to the actual OpenAI model name
	}

	// Map provider names to actual model names
	switch model {
	case "openai":
		model = "gpt-4o-mini"
	case "llama":
		model = "llama3:latest"
	case "sqlcoder":
		model = "sqlcoder:7b"
	}

	logger.LogInfo(logger.ServiceWS, "Processing raw AI message", map[string]interface{}{
		"content": content,
		"model":   model,
		"user_id": c.UserID,
	})

	// Send typing indicator
	c.sendMessage(Message{
		Type: "chat_typing",
		Payload: map[string]interface{}{
			"is_typing": true,
		},
		Timestamp: time.Now(),
	})

	// Process the raw AI message
	go c.processRawAIMessage(content, model)
}

// processRawAIMessage processes the actual raw AI message using real AI without system prompts
func (c *Client) processRawAIMessage(content, model string) {
	// Add panic recovery to prevent server crashes
	defer func() {
		if r := recover(); r != nil {
			logger.LogError(logger.ServiceWS, "Panic in processRawAIMessage", fmt.Errorf("panic: %v", r), map[string]interface{}{
				"content":   content,
				"model":     model,
				"client_id": c.ID,
			})
		}
	}()

	// Call raw AI service - no system prompts, just pass the user message directly
	response, err := c.callRawAIService(content, model)
	if err != nil {
		logger.LogError(logger.ServiceWS, "Raw AI service call failed", err, map[string]interface{}{
			"content": content,
			"model":   model,
		})
		response = "I'm sorry, I'm having trouble processing your request right now. Please try again."
	}

	// Stop typing indicator
	c.sendMessage(Message{
		Type: "chat_typing",
		Payload: map[string]interface{}{
			"is_typing": false,
		},
		Timestamp: time.Now(),
	})

	// Small delay to prevent message concatenation
	time.Sleep(50 * time.Millisecond)

	// Send AI response
	c.sendMessage(Message{
		Type: "raw_ai_response",
		Payload: map[string]interface{}{
			"content": response,
			"model":   model,
		},
		Timestamp: time.Now(),
	})

	logger.LogInfo(logger.ServiceWS, "Raw AI message processed", map[string]interface{}{
		"content":  content,
		"response": response,
		"model":    model,
	})
}

// callRawAIService calls the raw AI service without any system prompts
func (c *Client) callRawAIService(content, model string) (string, error) {
	if c.Hub.AIService == nil {
		return "AI service is not available. Please check the configuration.", nil
	}

	// Create messages with only the user content - no system prompts
	messages := []llm.Message{
		{
			Role:    "user",
			Content: content,
		},
	}

	// Type assert to get the AiRaw method
	aiService, ok := c.Hub.AIService.(interface {
		AiRaw(messages []llm.Message, modelOverride string) (*llm.ChatResponse, error)
	})
	if !ok {
		return "AI service does not support raw mode.", nil
	}

	// Call the raw AI service
	response, err := aiService.AiRaw(messages, model)
	if err != nil {
		return "", fmt.Errorf("raw AI service call failed: %w", err)
	}

	return response.Message.Content, nil
}

// processChatMessage processes the actual chat message using real AI
func (c *Client) processChatMessage(content, model string) {
	// Add panic recovery to prevent server crashes
	defer func() {
		if r := recover(); r != nil {
			logger.LogError(logger.ServiceWS, "Panic in processChatMessage", fmt.Errorf("panic: %v", r), map[string]interface{}{
				"content":   content,
				"model":     model,
				"client_id": c.ID,
			})
		}
	}()

	// Call real AI service
	response, err := c.callAIService(content, model)
	if err != nil {
		logger.LogError(logger.ServiceWS, "AI service call failed", err, map[string]interface{}{
			"content": content,
			"model":   model,
		})
		response = "I'm sorry, I'm having trouble processing your request right now. Please try again."
	}

	// Stop typing indicator
	c.sendMessage(Message{
		Type: "chat_typing",
		Payload: map[string]interface{}{
			"is_typing": false,
		},
		Timestamp: time.Now(),
	})

	// Small delay to prevent message concatenation
	time.Sleep(50 * time.Millisecond)

	// Send AI response
	c.sendMessage(Message{
		Type: "chat_response",
		Payload: map[string]interface{}{
			"content": response,
			"model":   model,
		},
		Timestamp: time.Now(),
	})

	logger.LogInfo(logger.ServiceWS, "Chat message processed", map[string]interface{}{
		"content":  content,
		"response": response,
		"model":    model,
	})
}

// callAIService calls the real AI service for chat responses
func (c *Client) callAIService(content, model string) (string, error) {
	if c.Hub.AIService == nil {
		return "AI service is not available. Please check the configuration.", nil
	}

	// Check if user has a loaded file and should analyze it
	var messages []llm.Message

	// If user asks generic questions and has a loaded file, analyze it
	if (strings.Contains(strings.ToLower(content), "what can you tell me") ||
		strings.Contains(strings.ToLower(content), "analyze") ||
		strings.Contains(strings.ToLower(content), "tell me about") ||
		strings.Contains(strings.ToLower(content), "what is") ||
		strings.Contains(strings.ToLower(content), "headers") ||
		strings.Contains(strings.ToLower(content), "columns") ||
		strings.Contains(strings.ToLower(content), "structure") ||
		strings.Contains(strings.ToLower(content), "count") ||
		strings.Contains(strings.ToLower(content), "how many") ||
		strings.Contains(strings.ToLower(content), "sum") ||
		strings.Contains(strings.ToLower(content), "total") ||
		strings.Contains(strings.ToLower(content), "years") ||
		strings.Contains(strings.ToLower(content), "data") ||
		strings.Contains(strings.ToLower(content), "show me") ||
		strings.Contains(strings.ToLower(content), "find") ||
		strings.Contains(strings.ToLower(content), "list")) && c.selectedFile != "" {

		// Get file data for analysis
		fileData, err := c.getFileDataForAnalysis(c.selectedFile)
		if err == nil && fileData != "" {
			messages = []llm.Message{
				{
					Role:    "system",
					Content: "You are AIR (AI Reporting Intelligence). You have access to a loaded dataset. Analyze the provided data and answer the user's question directly based on the actual data. Be specific and factual. Don't ask for more data - work with what you have.",
				},
				{
					Role:    "user",
					Content: fmt.Sprintf("User question: %s\n\nDataset:\n%s", content, fileData),
				},
			}
		} else {
			// Fallback to regular chat
			messages = []llm.Message{
				{
					Role:    "system",
					Content: "You are AIR (AI Reporting Intelligence), a specialized data analysis assistant. You help users analyze their specific datasets, create reports, and answer questions about their data. Always focus on the user's actual data and provide specific, actionable insights. Be concise and professional.",
				},
				{
					Role:    "user",
					Content: content,
				},
			}
		}
	} else {
		// Regular chat
		messages = []llm.Message{
			{
				Role:    "system",
				Content: "You are AIR (AI Reporting Intelligence), a specialized data analysis assistant. You help users analyze their specific datasets, create reports, and answer questions about their data. Always focus on the user's actual data and provide specific, actionable insights. Be concise and professional.",
			},
			{
				Role:    "user",
				Content: content,
			},
		}
	}

	// Type assert to get the ChatCompletion method
	aiService, ok := c.Hub.AIService.(interface {
		ChatCompletion(messages []llm.Message) (*llm.ChatResponse, error)
	})
	if !ok {
		return "AI service is not available.", nil
	}

	// Call the AI service
	response, err := aiService.ChatCompletion(messages)
	if err != nil {
		return "", fmt.Errorf("AI service call failed: %w", err)
	}

	return response.Message.Content, nil
}

// handleLoadDataset handles loading a dataset
func (c *Client) handleLoadDataset(message Message) {
	payload := message.Payload
	filename, ok := payload["filename"].(string)
	if !ok {
		c.sendMessage(Message{
			Type: "load_dataset_error",
			Payload: map[string]interface{}{
				"error": "Filename is required",
			},
			Timestamp: time.Now(),
		})
		return
	}

	// Set the selected file
	c.mu.Lock()
	c.selectedFile = filename
	c.mu.Unlock()

	// Send success response
	c.sendMessage(Message{
		Type: "load_dataset_success",
		Payload: map[string]interface{}{
			"filename": filename,
			"message":  fmt.Sprintf("✅ Loaded dataset: %s\n\nYou can now ask questions about this dataset or use /analyze to get insights.", filename),
		},
		Timestamp: time.Now(),
	})

	logger.LogInfo(logger.ServiceWS, "Dataset loaded", map[string]interface{}{
		"client_id": c.ID,
		"filename":  filename,
	})
}

// getFileDataForAnalysis reads file data for AI analysis
func (c *Client) getFileDataForAnalysis(fileID string) (string, error) {
	// Read the first 2000 characters of the file for analysis
	filePath := filepath.Join("uploads", fileID)

	logger.LogInfo(logger.ServiceWS, "Reading file for analysis", map[string]interface{}{
		"file_id":   fileID,
		"file_path": filePath,
	})

	file, err := os.Open(filePath)
	if err != nil {
		logger.LogError(logger.ServiceWS, "Failed to open file", err, map[string]interface{}{
			"file_id":   fileID,
			"file_path": filePath,
		})
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 2000 characters
	buffer := make([]byte, 2000)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Return file data with metadata
	fileData := fmt.Sprintf("File: %s\nSize: %d bytes\nData preview:\n%s",
		fileID, fileInfo.Size(), string(buffer[:n]))

	logger.LogInfo(logger.ServiceWS, "File data prepared for AI", map[string]interface{}{
		"file_id":      fileID,
		"file_size":    fileInfo.Size(),
		"data_preview": string(buffer[:n]),
	})

	return fileData, nil
}

// analyzeFileWithAI analyzes a file using real AI
func (c *Client) analyzeFileWithAI(filePath, query, model string) (string, []string, []string, error) {
	if c.Hub.AIService == nil {
		return "AI service is not available for file analysis.",
			[]string{"AI service unavailable"},
			[]string{"Please check AI service configuration"},
			nil
	}

	// Read file content
	file, err := os.Open(filePath)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read first 1000 characters for analysis
	buffer := make([]byte, 1000)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return "", nil, nil, fmt.Errorf("failed to read file: %w", err)
	}
	fileContent := string(buffer[:n])

	// Create AI prompt for file analysis
	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a data analysis expert. Analyze the provided file content and provide insights, suggestions, and a summary. Be specific and actionable.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Please analyze this file content:\n\n%s\n\nUser query: %s\n\nProvide:\n1. A summary analysis\n2. Key insights (as a JSON array)\n3. Actionable suggestions (as a JSON array)", fileContent, query),
		},
	}

	// Type assert to get the ChatCompletion method
	aiService, ok := c.Hub.AIService.(interface {
		ChatCompletion(messages []llm.Message) (*llm.ChatResponse, error)
	})
	if !ok {
		return "", nil, nil, fmt.Errorf("AI service is not available")
	}

	// Call AI service
	response, err := aiService.ChatCompletion(messages)
	if err != nil {
		return "", nil, nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Parse the response (simplified - in production, you'd want more robust parsing)
	analysis := response.Message.Content

	// Extract insights and suggestions from AI response
	// For now, return empty arrays - the AI response should contain the insights
	insights := []string{}
	suggestions := []string{}

	return analysis, insights, suggestions, nil
}
