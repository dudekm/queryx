package minecraft

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteVarInt(t *testing.T) {
	tests := []struct {
		name     string
		value    int
		expected []byte
	}{
		{"zero", 0, []byte{0x00}},
		{"small", 127, []byte{0x7F}},
		{"medium", 128, []byte{0x80, 0x01}},
		{"large", 300, []byte{0xAC, 0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := writeVarInt(buf, tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.Bytes())
		})
	}
}

func TestReadVarInt(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
	}{
		{"zero", []byte{0x00}, 0},
		{"small", []byte{0x7F}, 127},
		{"medium", []byte{0x80, 0x01}, 128},
		{"large", []byte{0xAC, 0x02}, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			result, err := readVarInt(reader)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteReadVarInt_RoundTrip(t *testing.T) {
	values := []int{0, 1, 127, 128, 255, 256, 2097151, 2147483647}

	for _, val := range values {
		buf := &bytes.Buffer{}
		err := writeVarInt(buf, val)
		assert.NoError(t, err)

		result, err := readVarInt(buf)
		assert.NoError(t, err)
		assert.Equal(t, val, result)
	}
}

func TestWriteString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect []byte
	}{
		{"empty", "", []byte{0x00}},
		{"hello", "hello", []byte{0x05, 'h', 'e', 'l', 'l', 'o'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := writeString(buf, tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, buf.Bytes())
		})
	}
}

func TestReadVarInt_TooLarge(t *testing.T) {
	// VarInt with 6 bytes (too long)
	data := []byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	reader := bytes.NewReader(data)
	_, err := readVarInt(reader)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VarInt too big")
}
