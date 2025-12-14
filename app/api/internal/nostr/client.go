package nostr

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// WebSocket-related errors
var (
	ErrConnectionClosed = errors.New("connection closed")
	ErrInvalidURL       = errors.New("invalid relay URL")
	ErrHandshakeFailed  = errors.New("WebSocket handshake failed")
	ErrInvalidFrame     = errors.New("invalid WebSocket frame")
	ErrMessageTooLarge  = errors.New("message too large")
)

// WebSocket opcodes
const (
	opContinuation = 0x0
	opText         = 0x1
	opBinary       = 0x2
	opClose        = 0x8
	opPing         = 0x9
	opPong         = 0xA
)

// Client is a Nostr relay WebSocket client.
type Client struct {
	url      string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	mu       sync.Mutex
	closed   atomic.Bool
	subCount atomic.Int64
}

// Filter represents a Nostr subscription filter.
type Filter struct {
	IDs     []string `json:"ids,omitempty"`
	Authors []string `json:"authors,omitempty"`
	Kinds   []int    `json:"kinds,omitempty"`
	Since   *int64   `json:"since,omitempty"`
	Until   *int64   `json:"until,omitempty"`
	Limit   int      `json:"limit,omitempty"`
}

// NewClient creates a new Nostr relay client.
func NewClient(relayURL string) *Client {
	return &Client{url: relayURL}
}

// Connect establishes a WebSocket connection to the relay.
func (c *Client) Connect(ctx context.Context) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// Determine host and port
	host := u.Hostname()
	port := u.Port()
	useTLS := u.Scheme == "wss"

	if port == "" {
		if useTLS {
			port = "443"
		} else {
			port = "80"
		}
	}

	addr := net.JoinHostPort(host, port)

	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	// Dial with context
	var conn net.Conn
	if useTLS {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Perform WebSocket handshake
	if err := c.handshake(conn, u); err != nil {
		conn.Close()
		return err
	}

	c.conn = conn
	c.reader = bufio.NewReader(conn)
	c.writer = bufio.NewWriter(conn)
	c.closed.Store(false)

	return nil
}

// handshake performs the WebSocket upgrade handshake.
func (c *Client) handshake(conn net.Conn, u *url.URL) error {
	// Generate random key
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return err
	}
	secKey := base64.StdEncoding.EncodeToString(key)

	// Build request path
	path := u.Path
	if path == "" {
		path = "/"
	}
	if u.RawQuery != "" {
		path += "?" + u.RawQuery
	}

	// Send upgrade request
	host := u.Host
	request := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Key: %s\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"\r\n",
		path, host, secKey)

	if _, err := conn.Write([]byte(request)); err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	// Read response
	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if !strings.Contains(statusLine, "101") {
		return fmt.Errorf("%w: %s", ErrHandshakeFailed, strings.TrimSpace(statusLine))
	}

	// Read headers until empty line
	expectedAccept := computeAcceptKey(secKey)
	gotAccept := false

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read headers: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		// Check Sec-WebSocket-Accept header
		if strings.HasPrefix(strings.ToLower(line), "sec-websocket-accept:") {
			value := strings.TrimSpace(line[21:])
			if value == expectedAccept {
				gotAccept = true
			}
		}
	}

	if !gotAccept {
		return fmt.Errorf("%w: invalid accept key", ErrHandshakeFailed)
	}

	return nil
}

// computeAcceptKey computes the Sec-WebSocket-Accept value.
func computeAcceptKey(key string) string {
	const guid = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + guid))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Close closes the WebSocket connection.
func (c *Client) Close() error {
	if c.closed.Swap(true) {
		return nil // Already closed
	}

	if c.conn != nil {
		// Send close frame
		c.writeFrame(opClose, []byte{})
		return c.conn.Close()
	}
	return nil
}

// Subscribe sends a REQ message and calls the callback for each event until EOSE.
func (c *Client) Subscribe(ctx context.Context, filter Filter, callback func(*SyncEvent) error) error {
	subID := fmt.Sprintf("sub-%d", c.subCount.Add(1))

	// Build REQ message: ["REQ", subID, filter]
	req := []interface{}{"REQ", subID, filter}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal REQ: %w", err)
	}

	// Send REQ
	if err := c.writeFrame(opText, reqJSON); err != nil {
		return fmt.Errorf("failed to send REQ: %w", err)
	}

	// Read messages until EOSE
	for {
		select {
		case <-ctx.Done():
			// Send CLOSE message
			closeMsg, _ := json.Marshal([]interface{}{"CLOSE", subID})
			c.writeFrame(opText, closeMsg)
			return ctx.Err()
		default:
		}

		// Set read deadline
		if c.conn != nil {
			c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		}

		// Read frame
		opcode, payload, err := c.readFrame()
		if err != nil {
			if errors.Is(err, ErrConnectionClosed) || c.closed.Load() {
				return err
			}
			// Timeout or other error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Read timeout, keep trying
			}
			return fmt.Errorf("failed to read frame: %w", err)
		}

		// Handle frame based on opcode
		switch opcode {
		case opText:
			// Parse message
			msgType, data, err := parseRelayMessage(payload)
			if err != nil {
				continue // Skip malformed messages
			}

			switch msgType {
			case "EVENT":
				// Parse event
				event, err := ParseEventFromRelay(data)
				if err != nil {
					continue // Skip malformed events
				}

				// Call callback
				if err := callback(event); err != nil {
					return err
				}

			case "EOSE":
				// End of stored events - send CLOSE and return
				closeMsg, _ := json.Marshal([]interface{}{"CLOSE", subID})
				c.writeFrame(opText, closeMsg)
				return nil

			case "NOTICE":
				// Log notices but continue
				continue

			case "CLOSED":
				// Subscription was closed by relay
				return nil
			}

		case opClose:
			c.closed.Store(true)
			return ErrConnectionClosed

		case opPing:
			// Respond with pong
			c.writeFrame(opPong, payload)

		case opPong:
			// Ignore pongs
		}
	}
}

// parseRelayMessage parses a Nostr relay message and returns the message type and data.
func parseRelayMessage(payload []byte) (string, []byte, error) {
	var raw []json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return "", nil, err
	}
	if len(raw) < 2 {
		return "", nil, errors.New("message too short")
	}

	var msgType string
	if err := json.Unmarshal(raw[0], &msgType); err != nil {
		return "", nil, err
	}

	// For EVENT messages, the event is the third element
	if msgType == "EVENT" && len(raw) >= 3 {
		return msgType, raw[2], nil
	}

	// For other messages, return the second element
	return msgType, raw[1], nil
}

// writeFrame writes a WebSocket frame.
func (c *Client) writeFrame(opcode byte, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed.Load() {
		return ErrConnectionClosed
	}

	// Generate mask key (clients must mask frames)
	maskKey := make([]byte, 4)
	if _, err := rand.Read(maskKey); err != nil {
		return err
	}

	// Build frame header
	frame := make([]byte, 0, 14+len(payload))

	// First byte: FIN + opcode
	frame = append(frame, 0x80|opcode)

	// Second byte: MASK + payload length
	payloadLen := len(payload)
	if payloadLen <= 125 {
		frame = append(frame, byte(0x80|payloadLen))
	} else if payloadLen <= 65535 {
		frame = append(frame, 0x80|126)
		frame = append(frame, byte(payloadLen>>8), byte(payloadLen))
	} else {
		frame = append(frame, 0x80|127)
		lenBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(lenBytes, uint64(payloadLen))
		frame = append(frame, lenBytes...)
	}

	// Add mask key
	frame = append(frame, maskKey...)

	// Mask and add payload
	maskedPayload := make([]byte, payloadLen)
	for i := 0; i < payloadLen; i++ {
		maskedPayload[i] = payload[i] ^ maskKey[i%4]
	}
	frame = append(frame, maskedPayload...)

	// Write frame
	_, err := c.conn.Write(frame)
	return err
}

// readFrame reads a WebSocket frame.
func (c *Client) readFrame() (byte, []byte, error) {
	if c.closed.Load() {
		return 0, nil, ErrConnectionClosed
	}

	// Read first two bytes
	header := make([]byte, 2)
	if _, err := io.ReadFull(c.reader, header); err != nil {
		return 0, nil, err
	}

	// fin := header[0]&0x80 != 0
	opcode := header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	// Extended payload length
	if payloadLen == 126 {
		extLen := make([]byte, 2)
		if _, err := io.ReadFull(c.reader, extLen); err != nil {
			return 0, nil, err
		}
		payloadLen = int(binary.BigEndian.Uint16(extLen))
	} else if payloadLen == 127 {
		extLen := make([]byte, 8)
		if _, err := io.ReadFull(c.reader, extLen); err != nil {
			return 0, nil, err
		}
		payloadLen = int(binary.BigEndian.Uint64(extLen))
	}

	// Sanity check on payload size (max 16MB)
	if payloadLen > 16*1024*1024 {
		return 0, nil, ErrMessageTooLarge
	}

	// Read mask key if present (server frames are typically not masked)
	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		if _, err := io.ReadFull(c.reader, maskKey); err != nil {
			return 0, nil, err
		}
	}

	// Read payload
	payload := make([]byte, payloadLen)
	if _, err := io.ReadFull(c.reader, payload); err != nil {
		return 0, nil, err
	}

	// Unmask if needed
	if masked {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	return opcode, payload, nil
}

// IsConnected returns true if the client is connected.
func (c *Client) IsConnected() bool {
	return !c.closed.Load() && c.conn != nil
}

// URL returns the relay URL.
func (c *Client) URL() string {
	return c.url
}
