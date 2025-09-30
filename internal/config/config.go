package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server           ServerConfig            `mapstructure:"server"`
	ControlPlane     ControlPlaneConfig      `mapstructure:"control_plane"`
	AnalyticsSources []AnalyticsSourceConfig `mapstructure:"analytics_sources"`
	Models           ModelsConfig            `mapstructure:"models"`
	Safety           SafetyConfig            `mapstructure:"safety"`
	Telemetry        TelemetryConfig         `mapstructure:"telemetry"`
	Redis            RedisConfig             `mapstructure:"redis"`
	WebSocket        WebSocketConfig         `mapstructure:"websocket"`
	Chat             ChatConfig              `mapstructure:"chat"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host      string     `mapstructure:"host"`
	Port      int        `mapstructure:"port"`
	WSEnabled bool       `mapstructure:"ws_enabled"`
	Auth      AuthConfig `mapstructure:"auth"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	JWTSecret   string        `mapstructure:"jwt_secret"`
	TokenExpiry time.Duration `mapstructure:"token_expiry"`
}

// ControlPlaneConfig holds control plane database configuration
type ControlPlaneConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

// AnalyticsSourceConfig holds analytics database configuration
type AnalyticsSourceConfig struct {
	ID          string `mapstructure:"id"`
	Kind        string `mapstructure:"kind"`
	DSN         string `mapstructure:"dsn"`
	DisplayName string `mapstructure:"display_name"`
	Default     bool   `mapstructure:"default"`
}

// ModelsConfig holds AI model configuration
type ModelsConfig struct {
	ChatPrimary string           `mapstructure:"chat_primary"`
	ChatBackup  string           `mapstructure:"chat_backup"`
	SQLPrimary  string           `mapstructure:"sql_primary"`
	OpenAI      OpenAIConfig     `mapstructure:"openai"`
	Ollama      OllamaConfig     `mapstructure:"ollama"`
	Embeddings  EmbeddingsConfig `mapstructure:"embeddings"`
}

// OpenAIConfig holds OpenAI configuration
type OpenAIConfig struct {
	Model  string `mapstructure:"model"`
	APIKey string `mapstructure:"api_key"`
}

// OllamaConfig holds Ollama configuration
type OllamaConfig struct {
	Host          string `mapstructure:"host"`
	Llama3Model   string `mapstructure:"llama3_model"`
	SQLCoderModel string `mapstructure:"sqlcoder_model"`
}

// EmbeddingsConfig holds embeddings configuration
type EmbeddingsConfig struct {
	Provider string `mapstructure:"provider"`
	Model    string `mapstructure:"model"`
}

// SafetyConfig holds safety guardrails configuration
type SafetyConfig struct {
	DefaultRowLimit       int `mapstructure:"default_row_limit"`
	MaxRowLimit           int `mapstructure:"max_row_limit"`
	EnforceTimeFilterDays int `mapstructure:"enforce_time_filter_days"`
}

// TelemetryConfig holds logging configuration
type TelemetryConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	TimeFormat string `mapstructure:"time_format"`
	Color      bool   `mapstructure:"color"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	URL          string        `mapstructure:"url"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
}

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	BufferSize        int           `mapstructure:"buffer_size"`
	ReadBufferSize    int           `mapstructure:"read_buffer_size"`
	WriteBufferSize   int           `mapstructure:"write_buffer_size"`
	HandshakeTimeout  time.Duration `mapstructure:"handshake_timeout"`
	PingPeriod        time.Duration `mapstructure:"ping_period"`
	PongWait          time.Duration `mapstructure:"pong_wait"`
	MaxMessageSize    int64         `mapstructure:"max_message_size"`
	EnableCompression bool          `mapstructure:"enable_compression"`
}

// ChatConfig holds live chat configuration
type ChatConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	MessageRetention  time.Duration `mapstructure:"message_retention"`
	TypingTimeout     time.Duration `mapstructure:"typing_timeout"`
	PresenceTimeout   time.Duration `mapstructure:"presence_timeout"`
	MaxRoomSize       int           `mapstructure:"max_room_size"`
	AIStreaming       bool          `mapstructure:"ai_streaming"`
	AIResponseTimeout time.Duration `mapstructure:"ai_response_timeout"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Set default values
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.ws_enabled", true)
	viper.SetDefault("server.auth.enabled", true)
	viper.SetDefault("server.auth.token_expiry", "24h")
	viper.SetDefault("control_plane.driver", "sqlite")
	viper.SetDefault("control_plane.dsn", "file:air.db?_fk=1")
	viper.SetDefault("models.chat_primary", "openai")
	viper.SetDefault("models.chat_backup", "llama3")
	viper.SetDefault("models.sql_primary", "sqlcoder")
	viper.SetDefault("models.openai.model", "gpt-4o-mini")
	viper.SetDefault("models.ollama.host", "http://localhost:11434")
	viper.SetDefault("models.ollama.llama3_model", "llama3")
	viper.SetDefault("models.ollama.sqlcoder_model", "sqlcoder")
	viper.SetDefault("models.embeddings.provider", "openai")
	viper.SetDefault("models.embeddings.model", "text-embedding-3-small")
	viper.SetDefault("safety.default_row_limit", 5000)
	viper.SetDefault("safety.max_row_limit", 100000)
	viper.SetDefault("safety.enforce_time_filter_days", 370)
	viper.SetDefault("telemetry.level", "info")
	viper.SetDefault("telemetry.format", "console")
	viper.SetDefault("telemetry.time_format", "15:04:05")
	viper.SetDefault("telemetry.color", true)

	// Redis defaults
	viper.SetDefault("redis.enabled", true)
	viper.SetDefault("redis.url", "redis://localhost:6379/0")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)

	// WebSocket defaults
	viper.SetDefault("websocket.enabled", true)
	viper.SetDefault("websocket.buffer_size", 1024)
	viper.SetDefault("websocket.read_buffer_size", 4096)
	viper.SetDefault("websocket.write_buffer_size", 4096)
	viper.SetDefault("websocket.handshake_timeout", "10s")
	viper.SetDefault("websocket.ping_period", "54s")
	viper.SetDefault("websocket.pong_wait", "60s")
	viper.SetDefault("websocket.max_message_size", 512)
	viper.SetDefault("websocket.enable_compression", true)

	// Chat defaults
	viper.SetDefault("chat.enabled", true)
	viper.SetDefault("chat.message_retention", "24h")
	viper.SetDefault("chat.typing_timeout", "5s")
	viper.SetDefault("chat.presence_timeout", "5m")
	viper.SetDefault("chat.max_room_size", 100)
	viper.SetDefault("chat.ai_streaming", true)
	viper.SetDefault("chat.ai_response_timeout", "30s")

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Auth.Enabled {
		if c.Server.Auth.JWTSecret == "" {
			return fmt.Errorf("jwt_secret is required when auth is enabled")
		}

		if c.Server.Auth.JWTSecret == "your-secret-key-change-in-production" {
			return fmt.Errorf("jwt_secret must be changed from default value in production")
		}
	}

	if len(c.AnalyticsSources) == 0 {
		return fmt.Errorf("at least one analytics source is required")
	}

	// Validate each analytics source
	validKinds := map[string]bool{
		"postgres":    true,
		"timescaledb": true,
		"mysql":       true,
		"sqlite":      true,
		"files":       true,
	}

	ids := make(map[string]bool)
	defaultCount := 0

	for i, source := range c.AnalyticsSources {
		if source.ID == "" {
			return fmt.Errorf("analytics_sources[%d].id is required", i)
		}

		if ids[source.ID] {
			return fmt.Errorf("duplicate analytics source id: %s", source.ID)
		}
		ids[source.ID] = true

		if !validKinds[source.Kind] {
			return fmt.Errorf("analytics_sources[%d].kind must be one of: postgres, timescaledb, mysql, sqlite, files", i)
		}

		// DSN required for SQL engines; files uses base_path
		if (source.Kind == "postgres" || source.Kind == "timescaledb" || source.Kind == "mysql" || source.Kind == "sqlite") && source.DSN == "" {
			return fmt.Errorf("analytics_sources[%d].dsn is required for kind %s", i, source.Kind)
		}

		if source.DisplayName == "" {
			return fmt.Errorf("analytics_sources[%d].display_name is required", i)
		}

		if source.Default {
			defaultCount++
		}
	}

	if defaultCount == 0 {
		return fmt.Errorf("at least one analytics source must be marked as default")
	}

	if defaultCount > 1 {
		return fmt.Errorf("only one analytics source can be marked as default")
	}

	return nil
}

// GetServerAddr returns the server address
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDefaultDatasource returns the default analytics source
func (c *Config) GetDefaultDatasource() *AnalyticsSourceConfig {
	for _, source := range c.AnalyticsSources {
		if source.Default {
			return &source
		}
	}
	return nil
}

// GetDatasourceByID returns an analytics source by ID
func (c *Config) GetDatasourceByID(id string) *AnalyticsSourceConfig {
	for _, source := range c.AnalyticsSources {
		if source.ID == id {
			return &source
		}
	}
	return nil
}
