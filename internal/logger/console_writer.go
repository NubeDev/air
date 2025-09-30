package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// CustomConsoleWriter implements a custom console writer for clean log format
type CustomConsoleWriter struct {
	Out        io.Writer
	TimeFormat string
	NoColor    bool
}

// Write implements io.Writer interface
func (w *CustomConsoleWriter) Write(p []byte) (n int, err error) {
	// Parse the JSON log entry
	var entry map[string]interface{}
	if err := json.Unmarshal(p, &entry); err != nil {
		// If parsing fails, just write the raw data
		return w.Out.Write(p)
	}

	// Extract fields
	timeStr := w.formatTime(entry["time"])
	level := w.formatLevel(entry["level"])
	message := w.getString(entry, "message")
	service := w.getString(entry, "service")

	// If no service specified, default to SVR
	if service == "" {
		service = "SVR"
	}

	// Build the log line
	var parts []string
	parts = append(parts, timeStr)
	parts = append(parts, fmt.Sprintf("[%s]", service))
	parts = append(parts, fmt.Sprintf("[%s]", level))
	parts = append(parts, message)

	// Add additional fields (excluding time, level, message, service)
	for key, value := range entry {
		if key != "time" && key != "level" && key != "message" && key != "service" {
			parts = append(parts, fmt.Sprintf("%s=%v", key, value))
		}
	}

	logLine := strings.Join(parts, " ") + "\n"

	// Apply colors if enabled
	if !w.NoColor {
		logLine = w.applyColors(logLine, service, level)
	}

	return w.Out.Write([]byte(logLine))
}

func (w *CustomConsoleWriter) formatTime(t interface{}) string {
	if t == nil {
		return time.Now().Format(w.TimeFormat)
	}

	if timeStr, ok := t.(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return parsedTime.Format(w.TimeFormat)
		}
	}

	return time.Now().Format(w.TimeFormat)
}

func (w *CustomConsoleWriter) formatLevel(level interface{}) string {
	if level == nil {
		return "INFO"
	}

	if levelStr, ok := level.(string); ok {
		switch strings.ToLower(levelStr) {
		case "info":
			return "INFO"
		case "warn", "warning":
			return "WARN"
		case "error":
			return "ERRO"
		case "debug":
			return "DEBG"
		default:
			return "INFO"
		}
	}

	return "INFO"
}

func (w *CustomConsoleWriter) getString(entry map[string]interface{}, key string) string {
	if value, exists := entry[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

func (w *CustomConsoleWriter) applyColors(line, service, level string) string {
	if w.NoColor {
		return line
	}

	// Always apply colors for now to test
	// Apply service color
	if color, exists := serviceColors[service]; exists {
		serviceTag := fmt.Sprintf("[%s]", service)
		line = w.colorize(line, serviceTag, color)
	}

	// Apply level color
	if color, exists := levelColors[level]; exists {
		levelTag := fmt.Sprintf("[%s]", level)
		line = w.colorize(line, levelTag, color)
	}

	return line
}

func (w *CustomConsoleWriter) isTerminal() bool {
	// Check if the output is a terminal
	if file, ok := w.Out.(*os.File); ok {
		return file == os.Stdout || file == os.Stderr
	}
	return false
}

func (w *CustomConsoleWriter) colorize(text, target string, colorCode int) string {
	// ANSI escape codes for colors
	start := fmt.Sprintf("\033[%dm", colorCode)
	end := "\033[0m"

	// Replace the target with colored version
	coloredTarget := start + target + end
	result := strings.Replace(text, target, coloredTarget, 1)

	// Debug: show what we're doing
	// fmt.Fprintf(os.Stderr, "DEBUG COLORIZE: target='%s', colorCode=%d, result='%s'\n", target, colorCode, result)

	return result
}
