package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	// Server connection
	conn      *websocket.Conn
	serverURL string
	closed    bool // Track if connection is closed
}

type Message struct {
	Type      string      `json:"type"`
	Hostname  string      `json:"hostname,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
	VM        string      `json:"vm,omitempty"`
	Status    string      `json:"status,omitempty"`
	Error     string      `json:"error,omitempty"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

// Connect establishes WebSocket connection to control daemon
func (c *Client) Connect() error {
	log.Printf("Connecting to %s", c.serverURL)

	conn, _, err := websocket.DefaultDialer.Dial(c.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	log.Println("Connected to control daemon")
	return nil
}

// SendHeartbeat sends a heartbeat message to control daemon
func (c *Client) SendHeartbeat(hostname string) error {
	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	msg := Message{
		Type:      "heartbeat",
		Hostname:  hostname,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Heartbeat sent for %s", hostname)
	return nil
}

// ReceiveMessage reads a message from control daemon
func (c *Client) ReceiveMessage() (*Message, error) {
	if c.closed {
		return nil, fmt.Errorf("connection closed")
	}
	if c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	_, message, err := c.conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Printf("Received message: %s", msg.Type)
	return &msg, nil
}

// IsConnected checks if WebSocket connection is active
func (c *Client) IsConnected() bool {
	return c.conn != nil && !c.closed
}

// IsClosed returns true if the connection has been closed
func (c *Client) IsClosed() bool {
	return c.closed
}

// Close closes the WebSocket connection and clears the pointer
func (c *Client) Close() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		c.closed = true // Mark as closed
		return err
	}
	return nil
}

// Reconnect attempts to reconnect with exponential backoff
func (c *Client) Reconnect(maxDelay time.Duration) error {
	// Close existing connection if any
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// Reset closed state for new connection attempt
	c.closed = false

	delay := 5 * time.Second
	for delay < maxDelay {
		log.Printf("Reconnecting in %v...", delay)

		if err := c.Connect(); err == nil {
			c.closed = false // Ensure closed flag is false after success
			return nil
		}

		time.Sleep(delay)
		delay *= 2
	}

	return fmt.Errorf("reconnection failed after %v", maxDelay)
}