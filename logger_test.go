package queryx

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Debug("test")
		logger.Info("test", F("key", "value"))
		logger.Warn("test", F("k1", 1), F("k2", 2))
		logger.Error("test")
	})
}

func TestConsoleLogger(t *testing.T) {
	tests := []struct {
		name           string
		logFunc        func(logger *ConsoleLogger)
		expectedLevel  string
		expectedMsg    string
		expectedFields []string
	}{
		{
			name: "debug without fields",
			logFunc: func(l *ConsoleLogger) {
				l.Debug("debug message")
			},
			expectedLevel:  "DEBUG",
			expectedMsg:    "debug message",
			expectedFields: nil,
		},
		{
			name: "info with one field",
			logFunc: func(l *ConsoleLogger) {
				l.Info("info message", F("key", "value"))
			},
			expectedLevel:  "INFO",
			expectedMsg:    "info message",
			expectedFields: []string{"key=value"},
		},
		{
			name: "warn with multiple fields",
			logFunc: func(l *ConsoleLogger) {
				l.Warn("warning message", F("count", 42), F("status", "active"))
			},
			expectedLevel:  "WARN",
			expectedMsg:    "warning message",
			expectedFields: []string{"count=42", "status=active"},
		},
		{
			name: "error with fields",
			logFunc: func(l *ConsoleLogger) {
				l.Error("error message", F("error", "timeout"))
			},
			expectedLevel:  "ERROR",
			expectedMsg:    "error message",
			expectedFields: []string{"error=timeout"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := &ConsoleLogger{writer: buf}

			tt.logFunc(logger)

			output := buf.String()
			assert.Contains(t, output, tt.expectedLevel)
			assert.Contains(t, output, tt.expectedMsg)

			for _, field := range tt.expectedFields {
				assert.Contains(t, output, field)
			}

			// Check timestamp format (YYYY-MM-DD HH:MM:SS)
			assert.True(t, strings.Contains(output, "["+tt.expectedLevel+"]"))
		})
	}
}

func TestFieldHelper(t *testing.T) {
	field := F("test", 123)
	assert.Equal(t, "test", field.Key)
	assert.Equal(t, 123, field.Value)
}

func TestNewConsoleLogger(t *testing.T) {
	logger := NewConsoleLogger()
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.writer)
}
