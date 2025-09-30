package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Service codes (4 letters for consistency)
const (
	ServiceServer = "SERV"
	ServiceDB     = "DATA"
	ServiceREST   = "HTTP"
	ServiceAI     = "AI  "
	ServiceWS     = "WS  "
	ServiceAuth   = "AUTH"
	ServiceConfig = "CONF"
	ServiceFile   = "FILE"
)

// Log levels (4 letters for consistency)
const (
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERRO"
	LevelDebug = "DEBG"
)

// Service colors
var serviceColors = map[string]int{
	ServiceServer: 36, // Cyan
	ServiceDB:     36, // Cyan
	ServiceREST:   33, // Yellow
	ServiceAI:     35, // Magenta
	ServiceWS:     32, // Green
	ServiceAuth:   31, // Red
	ServiceConfig: 34, // Blue
	ServiceFile:   37, // White
}

// Level colors
var levelColors = map[string]int{
	LevelInfo:  36, // Cyan
	LevelWarn:  33, // Yellow
	LevelError: 31, // Red
	LevelDebug: 37, // White
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"` // json, console
	TimeFormat string `yaml:"time_format" mapstructure:"time_format"`
	Color      bool   `yaml:"color" mapstructure:"color"`
}

// DefaultLoggerConfig returns default logging configuration
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      "info",
		Format:     "console",
		TimeFormat: "15:04:05",
		Color:      true,
	}
}

// SetupLogger configures the global logger
func SetupLogger(config *LoggerConfig) {
	// Set log level
	level, err := zerolog.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure time format
	zerolog.TimeFieldFormat = time.RFC3339
	if config.TimeFormat != "" {
		zerolog.TimeFieldFormat = config.TimeFormat
	}

	// Configure output format
	if config.Format == "json" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Custom console format: time [SERVICE] [LEVEL] message
		output := &CustomConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: config.TimeFormat,
			NoColor:    !config.Color,
		}
		log.Logger = zerolog.New(output).With().Timestamp().Logger()
	}
}

// Common logging functions
func Log(service, level, message string, fields ...map[string]interface{}) {
	baseLogger := log.Logger

	// Add service and level as fields
	event := baseLogger.With().
		Str("service", service).
		Str("level", level)

	// Add any additional fields
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			event = event.Interface(key, value)
		}
	}

	// Get the logger and log based on level
	logger := event.Logger()
	switch level {
	case LevelInfo:
		logger.Info().Msg(message)
	case LevelWarn:
		logger.Warn().Msg(message)
	case LevelError:
		logger.Error().Msg(message)
	case LevelDebug:
		logger.Debug().Msg(message)
	default:
		logger.Info().Msg(message)
	}
}

// Convenience functions
func LogInfo(service, message string, fields ...map[string]interface{}) {
	Log(service, LevelInfo, message, fields...)
}

func LogWarn(service, message string, fields ...map[string]interface{}) {
	Log(service, LevelWarn, message, fields...)
}

func LogError(service, message string, err error, fields ...map[string]interface{}) {
	errorFields := append(fields, map[string]interface{}{"error": err.Error()})
	Log(service, LevelError, message, errorFields...)
}

func LogDebug(service, message string, fields ...map[string]interface{}) {
	Log(service, LevelDebug, message, fields...)
}

// Specialized logging functions
func LogRequest(method, path string, status int, duration time.Duration, clientIP string) {
	LogInfo(ServiceREST, "HTTP request", map[string]interface{}{
		"method":    method,
		"path":      path,
		"status":    status,
		"duration":  duration.String(),
		"client_ip": clientIP,
	})
}

func LogDBOperation(operation, table string, duration time.Duration, rowsAffected int64) {
	LogInfo(ServiceDB, "Database operation", map[string]interface{}{
		"operation":     operation,
		"table":         table,
		"duration":      duration.String(),
		"rows_affected": rowsAffected,
	})
}

func LogAIOperation(operation, model string, duration time.Duration, tokens int) {
	LogInfo(ServiceAI, "AI operation", map[string]interface{}{
		"operation": operation,
		"model":     model,
		"duration":  duration.String(),
		"tokens":    tokens,
	})
}

func LogFileOperation(operation, path string, size int64, duration time.Duration) {
	LogInfo(ServiceFile, "File operation", map[string]interface{}{
		"operation":  operation,
		"path":       path,
		"size_bytes": size,
		"duration":   duration.String(),
	})
}
