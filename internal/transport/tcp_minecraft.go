package transport

import (
	"context"
	"fmt"
	"io"
	"net"
)

// SendTCPMinecraft sends data via TCP and reads a Minecraft protocol response
// Minecraft uses VarInt-prefixed packets, so we need to read the length first
func SendTCPMinecraft(ctx context.Context, addr string, data []byte) ([]byte, error) {
	// Resolve TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve TCP address %s: %w", addr, err)
	}

	// Create dialer with context
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", tcpAddr.String())
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP %s: %w", addr, err)
	}
	defer conn.Close()

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return nil, fmt.Errorf("failed to set deadline: %w", err)
		}
	}

	// Send data
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write TCP data: %w", err)
	}

	// Read VarInt packet length
	packetLen, err := readVarInt(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read packet length: %w", err)
	}

	// Read the exact packet data
	packetData := make([]byte, packetLen)
	if _, err := io.ReadFull(conn, packetData); err != nil {
		return nil, fmt.Errorf("failed to read packet data: %w", err)
	}

	// Prepend the length VarInt to the response (protocol expects it)
	result := encodeVarInt(packetLen)
	result = append(result, packetData...)

	return result, nil
}

// readVarInt reads a VarInt from the connection
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

// encodeVarInt encodes an integer as VarInt bytes
func encodeVarInt(value int) []byte {
	var result []byte
	uvalue := uint32(value)

	for {
		if (uvalue & 0xFFFFFF80) == 0 {
			result = append(result, byte(uvalue))
			return result
		}
		result = append(result, byte(uvalue&0x7F|0x80))
		uvalue >>= 7
	}
}
