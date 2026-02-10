package queryx

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryError_Error(t *testing.T) {
	err := NewQueryError(GameMinecraft, "example.com", ErrTimeout)
	expected := "query failed for minecraft at example.com: query timeout"
	assert.Equal(t, expected, err.Error())
}

func TestQueryError_Unwrap(t *testing.T) {
	originalErr := ErrTimeout
	err := NewQueryError(GameMinecraft, "example.com", originalErr)

	assert.Equal(t, originalErr, err.Unwrap())
	assert.True(t, errors.Is(err, ErrTimeout))
}

func TestQueryError_ErrorsIs(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "matches wrapped error",
			err:      NewQueryError(GameMinecraft, "example.com", ErrTimeout),
			target:   ErrTimeout,
			expected: true,
		},
		{
			name:     "does not match different error",
			err:      NewQueryError(GameMinecraft, "example.com", ErrTimeout),
			target:   ErrInvalidResponse,
			expected: false,
		},
		{
			name:     "matches nested wrapped error",
			err:      NewQueryError(GameMinecraft, "example.com", fmt.Errorf("context: %w", ErrServerOffline)),
			target:   ErrServerOffline,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.Is(tt.err, tt.target))
		})
	}
}

func TestQueryError_ErrorsAs(t *testing.T) {
	originalErr := NewQueryError(GameMinecraft, "example.com", ErrTimeout)
	wrappedErr := fmt.Errorf("outer error: %w", originalErr)

	var queryErr *QueryError
	assert.True(t, errors.As(wrappedErr, &queryErr))
	assert.Equal(t, GameMinecraft, queryErr.GameType)
	assert.Equal(t, "example.com", queryErr.Host)
}

func TestCommonErrors(t *testing.T) {
	assert.NotNil(t, ErrTimeout)
	assert.NotNil(t, ErrInvalidResponse)
	assert.NotNil(t, ErrServerOffline)
	assert.NotNil(t, ErrUnsupportedGame)
	assert.NotNil(t, ErrInvalidHost)
	assert.NotNil(t, ErrDNSResolution)
}
