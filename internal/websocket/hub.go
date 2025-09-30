package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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
	ID       string
	UserID   string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
	Channels map[string]bool // Subscribed channels
	mu       sync.RWMutex
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
func NewHub(redisClient *redis.Client, config *Config) *Hub {
	return &Hub{
		Clients:        make(map[*Client]bool),
		Channels:       make(map[string]map[*Client]bool),
		Register:       make(chan *Client),
		Unregister:     make(chan *Client),
		Broadcast:      make(chan []byte),
		ChannelMessage: make(chan ChannelMessage),
		Redis:          redisClient,
		Config:         config,
	}
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
		return fmt.Errorf("Redis client is not available")
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
