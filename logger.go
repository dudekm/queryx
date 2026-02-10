package queryx

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Logger is the interface for logging in QueryX
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// F creates a new Field for structured logging
func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// NoOpLogger is a logger that does nothing (default)
type NoOpLogger struct{}

func (n *NoOpLogger) Debug(msg string, fields ...Field) {}
func (n *NoOpLogger) Info(msg string, fields ...Field)  {}
func (n *NoOpLogger) Warn(msg string, fields ...Field)  {}
func (n *NoOpLogger) Error(msg string, fields ...Field) {}

// ConsoleLogger is a simple console logger
type ConsoleLogger struct {
	writer io.Writer
}

// NewConsoleLogger creates a new console logger that writes to stdout
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{writer: os.Stdout}
}

func (c *ConsoleLogger) log(level string, msg string, fields ...Field) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(c.writer, "[%s] %s: %s", level, timestamp, msg)
	for _, f := range fields {
		fmt.Fprintf(c.writer, " %s=%v", f.Key, f.Value)
	}
	fmt.Fprintln(c.writer)
}

func (c *ConsoleLogger) Debug(msg string, fields ...Field) { c.log("DEBUG", msg, fields...) }
func (c *ConsoleLogger) Info(msg string, fields ...Field)  { c.log("INFO", msg, fields...) }
func (c *ConsoleLogger) Warn(msg string, fields ...Field)  { c.log("WARN", msg, fields...) }
func (c *ConsoleLogger) Error(msg string, fields ...Field) { c.log("ERROR", msg, fields...) }
