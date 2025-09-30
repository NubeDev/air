package websocket

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/NubeDev/air/internal/config"
	"github.com/NubeDev/air/internal/logger"
	"github.com/NubeDev/air/internal/redis"
	ws "github.com/NubeDev/air/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Handler handles WebSocket connections
type Handler struct {
	hub    *ws.Hub
	redis  *redis.Client
	config *config.WebSocketConfig
}

// NewHandler creates a new WebSocket handler
func NewHandler(redisClient *redis.Client, wsConfig *config.WebSocketConfig) *Handler {
	// Create WebSocket hub configuration
	hubConfig := &ws.Config{
		ReadBufferSize:    wsConfig.ReadBufferSize,
		WriteBufferSize:   wsConfig.WriteBufferSize,
		HandshakeTimeout:  wsConfig.HandshakeTimeout,
		PingPeriod:        wsConfig.PingPeriod,
		PongWait:          wsConfig.PongWait,
		MaxMessageSize:    wsConfig.MaxMessageSize,
		EnableCompression: wsConfig.EnableCompression,
	}

	hub := ws.NewHub(redisClient, hubConfig)

	return &Handler{
		hub:    hub,
		redis:  redisClient,
		config: wsConfig,
	}
}

// Upgrader handles WebSocket upgrades
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now - in production, implement proper CORS
		return true
	},
}

// HandleWebSocket handles WebSocket connections
func (h *Handler) HandleWebSocket(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebSocket service is disabled",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LogError(logger.ServiceWS, "Failed to upgrade connection", err)
		return
	}

	// Generate client ID
	clientID := generateClientID()

	// Extract user ID from query parameters or headers
	userID := c.Query("user_id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		userID = "anonymous"
	}

	// Create client
	client := &ws.Client{
		ID:       clientID,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
		Channels: make(map[string]bool),
	}

	// Register client with hub
	h.hub.Register <- client

	logger.LogInfo(logger.ServiceWS, "WebSocket client connected", map[string]interface{}{
		"client_id":   clientID,
		"user_id":     userID,
		"remote_addr": c.Request.RemoteAddr,
	})
}

// HandleChat handles chat-specific WebSocket connections
func (h *Handler) HandleChat(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebSocket service is disabled",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LogError(logger.ServiceWS, "Failed to upgrade chat connection", err)
		return
	}

	// Generate client ID
	clientID := generateClientID()

	// Extract user ID from query parameters or headers
	userID := c.Query("user_id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		userID = "anonymous"
	}

	// Create client
	client := &ws.Client{
		ID:       clientID,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
		Channels: make(map[string]bool),
	}

	// Register client with hub
	h.hub.Register <- client

	// Auto-subscribe to chat channels
	h.hub.SubscribeToChannel(client, "chat:general")
	h.hub.SubscribeToChannel(client, fmt.Sprintf("chat:user:%s", userID))

	logger.LogInfo(logger.ServiceWS, "Chat WebSocket client connected", map[string]interface{}{
		"client_id":   clientID,
		"user_id":     userID,
		"remote_addr": c.Request.RemoteAddr,
	})
}

// HandlePresence handles presence WebSocket connections
func (h *Handler) HandlePresence(c *gin.Context) {
	if !h.config.Enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebSocket service is disabled",
		})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.LogError(logger.ServiceWS, "Failed to upgrade presence connection", err)
		return
	}

	// Generate client ID
	clientID := generateClientID()

	// Extract user ID from query parameters or headers
	userID := c.Query("user_id")
	if userID == "" {
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		userID = "anonymous"
	}

	// Create client
	client := &ws.Client{
		ID:       clientID,
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
		Channels: make(map[string]bool),
	}

	// Register client with hub
	h.hub.Register <- client

	// Auto-subscribe to presence channels
	h.hub.SubscribeToChannel(client, "presence:online")
	h.hub.SubscribeToChannel(client, fmt.Sprintf("typing:user:%s", userID))

	// Set user as online in Redis
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		h.redis.SAdd(ctx, "online_users", userID)
		h.redis.Expire(ctx, "online_users", 5*time.Minute)
	}

	logger.LogInfo(logger.ServiceWS, "Presence WebSocket client connected", map[string]interface{}{
		"client_id":   clientID,
		"user_id":     userID,
		"remote_addr": c.Request.RemoteAddr,
	})
}

// GetOnlineUsers returns the list of online users
func (h *Handler) GetOnlineUsers(c *gin.Context) {
	if h.redis == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Redis service is not available",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := h.redis.SMembers(ctx, "online_users")
	if err != nil {
		logger.LogError(logger.ServiceWS, "Failed to get online users", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get online users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online_users": users,
		"count":        len(users),
	})
}

// SendMessage sends a message to a specific user or channel
func (h *Handler) SendMessage(c *gin.Context) {
	var req struct {
		UserID  string                 `json:"user_id,omitempty"`
		Channel string                 `json:"channel,omitempty"`
		Type    string                 `json:"type"`
		Payload map[string]interface{} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	message := ws.Message{
		Type:      req.Type,
		Channel:   req.Channel,
		Payload:   req.Payload,
		Timestamp: time.Now(),
	}

	// Send to specific user
	if req.UserID != "" {
		if err := h.hub.SendToUser(req.UserID, message); err != nil {
			logger.LogError(logger.ServiceWS, "Failed to send message to user", err, map[string]interface{}{
				"user_id": req.UserID,
			})
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to send message",
			})
			return
		}
	} else if req.Channel != "" {
		// Send to channel via Redis
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		redisChannel := fmt.Sprintf("websocket:%s", req.Channel)
		if err := h.hub.PublishToRedis(ctx, redisChannel, message); err != nil {
			logger.LogError(logger.ServiceWS, "Failed to send message to channel", err, map[string]interface{}{
				"channel": req.Channel,
			})
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to send message",
			})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Either user_id or channel must be specified",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message sent successfully",
	})
}

// GetHubStats returns WebSocket hub statistics
func (h *Handler) GetHubStats(c *gin.Context) {
	h.hub.Mu.RLock()
	defer h.hub.Mu.RUnlock()

	stats := gin.H{
		"total_clients":  len(h.hub.Clients),
		"total_channels": len(h.hub.Channels),
		"channels":       make(map[string]int),
	}

	// Count clients per channel
	for channel, clients := range h.hub.Channels {
		stats["channels"].(map[string]int)[channel] = len(clients)
	}

	c.JSON(http.StatusOK, stats)
}

// generateClientID generates a unique client ID
func generateClientID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// StartHub starts the WebSocket hub
func (h *Handler) StartHub(ctx context.Context) {
	go h.hub.Run(ctx)
}
