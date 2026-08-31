package cfxre

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEndpoint_Normalization(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		wantBase string
	}{
		{"host:port adds http scheme", "1.2.3.4:30120", "http://1.2.3.4:30120"},
		{"keeps http scheme", "http://example.com:30120", "http://example.com:30120"},
		{"keeps https scheme", "https://example.com:30120", "https://example.com:30120"},
		{"trims trailing slash", "http://example.com:30120/", "http://example.com:30120"},
		{"trims surrounding spaces", "  1.2.3.4:30120  ", "http://1.2.3.4:30120"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantBase, NewEndpoint(tt.addr).Base())
		})
	}
}

func TestEndpoint_URLs(t *testing.T) {
	e := NewEndpoint("1.2.3.4:30120")
	assert.Equal(t, "http://1.2.3.4:30120/info.json", e.Info())
	assert.Equal(t, "http://1.2.3.4:30120/players.json", e.Players())
	assert.Equal(t, "http://1.2.3.4:30120/dynamic.json", e.Dynamic())
}

// Endpoint is a value object: equal inputs produce equal, comparable values.
func TestEndpoint_ValueEquality(t *testing.T) {
	a := NewEndpoint("example.com:30120")
	b := NewEndpoint("http://example.com:30120/")
	assert.Equal(t, a, b, "normalized endpoints should be equal by value")
	assert.True(t, a == b, "value objects should be comparable with ==")
}
