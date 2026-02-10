package queryx

import (
	"errors"
	"fmt"
)

// Common errors returned by QueryX
var (
	ErrTimeout         = errors.New("query timeout")
	ErrInvalidResponse = errors.New("invalid server response")
	ErrServerOffline   = errors.New("server offline")
	ErrUnsupportedGame = errors.New("unsupported game type")
	ErrInvalidHost     = errors.New("invalid host")
	ErrDNSResolution   = errors.New("dns resolution failed")
)

// QueryError wraps an error with context about the query that failed
type QueryError struct {
	GameType GameType
	Host     string
	Err      error
}

// Error implements the error interface
func (e *QueryError) Error() string {
	return fmt.Sprintf("query failed for %s at %s: %v", e.GameType, e.Host, e.Err)
}

// Unwrap implements the error unwrapping interface
func (e *QueryError) Unwrap() error {
	return e.Err
}

// NewQueryError creates a new QueryError
func NewQueryError(gameType GameType, host string, err error) *QueryError {
	return &QueryError{
		GameType: gameType,
		Host:     host,
		Err:      err,
	}
}
