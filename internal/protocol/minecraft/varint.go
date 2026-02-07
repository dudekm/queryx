package minecraft

import (
	"bytes"
	"fmt"
	"io"
)

// writeVarInt writes a VarInt to the buffer
func writeVarInt(buf *bytes.Buffer, value int) error {
	// Convert to unsigned to handle negative values correctly
	uvalue := uint32(value)
	for {
		if (uvalue & 0xFFFFFF80) == 0 {
			buf.WriteByte(byte(uvalue))
			return nil
		}
		buf.WriteByte(byte(uvalue&0x7F | 0x80))
		uvalue >>= 7
	}
}

// readVarInt reads a VarInt from the reader
func readVarInt(r io.Reader) (int, error) {
	var result int
	var numRead int

	for {
		b := make([]byte, 1)
		if _, err := r.Read(b); err != nil {
			return 0, err
		}

		value := int(b[0] & 0x7F)
		result |= value << (7 * numRead)

		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("VarInt too big")
		}

		if (b[0] & 0x80) == 0 {
			break
		}
	}

	return result, nil
}

// writeString writes a VarInt-prefixed string
func writeString(buf *bytes.Buffer, s string) error {
	if err := writeVarInt(buf, len(s)); err != nil {
		return err
	}
	buf.WriteString(s)
	return nil
}
