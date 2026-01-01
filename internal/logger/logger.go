package logger

import (
	"fmt"
	"math/rand"
	"time"

	json "github.com/goccy/go-json"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	DEBUG   LogLevel = "DEBUG"
	INFO    LogLevel = "INFO"
	WARNING LogLevel = "WARNING"
	ERROR   LogLevel = "ERROR"
	FATAL   LogLevel = "FATAL"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string   `json:"timestamp"`
	Level     LogLevel `json:"level"`
	Service   string   `json:"service"`
	Message   string   `json:"message"`
	RequestID string   `json:"request_id,omitempty"`
	UserID    string   `json:"user_id,omitempty"`
	Duration  int      `json:"duration_ms,omitempty"`
}

// Service generates logs for testing purposes
type Service struct {
	serviceName string
	services    []string
	messages    map[LogLevel][]string
}

// NewService creates a new logging service
func NewService(serviceName string) *Service {
	return &Service{
		serviceName: serviceName,
		services:    []string{"api-gateway", "auth-service", "user-service", "payment-service", "notification-service"},
		messages: map[LogLevel][]string{
			DEBUG:   {"Processing request", "Cache hit", "Database query executed", "Parsing input"},
			INFO:    {"Request completed", "User logged in", "Session created", "File uploaded"},
			WARNING: {"Slow query detected", "High memory usage", "Rate limit approaching", "Deprecated API usage"},
			ERROR:   {"Database connection failed", "Authentication failed", "Invalid input", "Service timeout"},
			FATAL:   {"Out of memory", "Disk full", "Critical service unavailable", "Configuration error"},
		},
	}
}

// generateRequestID creates a random request ID
func generateRequestID() string {
	const chars = "abcdef0123456789"
	id := make([]byte, 8)
	for i := range id {
		id[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("req-%s", string(id))
}

// generateUserID creates a random user ID
func generateUserID() string {
	return fmt.Sprintf("user-%d", rand.Intn(10000))
}

// GenerateLog creates a random log entry
func (s *Service) GenerateLog() LogEntry {
	levels := []LogLevel{DEBUG, INFO, INFO, INFO, WARNING, ERROR, FATAL}
	weights := []int{15, 50, 50, 50, 20, 10, 2} // Weighted distribution

	// Weighted random selection
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	r := rand.Intn(totalWeight)
	cumulative := 0
	selectedLevel := INFO
	for i, w := range weights {
		cumulative += w
		if r < cumulative {
			selectedLevel = levels[i]
			break
		}
	}

	messages := s.messages[selectedLevel]
	message := messages[rand.Intn(len(messages))]
	service := s.services[rand.Intn(len(s.services))]

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     selectedLevel,
		Service:   service,
		Message:   message,
		RequestID: generateRequestID(),
	}

	// Add optional fields based on context
	if rand.Float32() > 0.3 {
		entry.UserID = generateUserID()
	}
	if selectedLevel == INFO || selectedLevel == WARNING {
		entry.Duration = rand.Intn(5000) + 1
	}

	return entry
}

// GenerateLogs continuously generates logs at the specified interval
func (s *Service) GenerateLogs(interval time.Duration, output chan<- LogEntry, done <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			output <- s.GenerateLog()
		case <-done:
			return
		}
	}
}

// FormatJSON converts a log entry to JSON string
func (e LogEntry) FormatJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// FormatText converts a log entry to human-readable text
func (e LogEntry) FormatText() string {
	return fmt.Sprintf("[%s] %s | %s | %s | %s",
		e.Timestamp, e.Level, e.Service, e.RequestID, e.Message)
}
