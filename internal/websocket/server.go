package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	upgrader websocket.Upgrader
	hosts    map[string]HostInfo
	mu       sync.RWMutex
}

type HostInfo struct {
	Hostname    string    `json:"hostname"`
	LastSeen    time.Time `json:"last_seen"`
	ConnectedAt time.Time `json:"connected_at"`
}

func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for now
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		hosts: make(map[string]HostInfo),
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// Set max message size (1MB for heartbeats)
	conn.SetReadLimit(1024 * 1024)

	defer conn.Close()

	// Get hostname from query parameter or connection ID
	hostname := r.URL.Query().Get("hostname")
	if hostname == "" {
		hostname = "unknown"
	}

	s.mu.Lock()
	s.hosts[hostname] = HostInfo{
		Hostname:    hostname,
		LastSeen:    time.Now(),
		ConnectedAt: time.Now(),
	}
	s.mu.Unlock()

	log.Printf("Host connected: %s", hostname)

	// Read messages from client
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection error for %s: %v", hostname, err)
			s.mu.Lock()
			delete(s.hosts, hostname)
			s.mu.Unlock()
			break
		}

		if messageType == websocket.TextMessage {
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to unmarshal message from %s: %v", hostname, err)
				continue
			}

			// Update last seen timestamp
			s.mu.Lock()
			if host, ok := s.hosts[msg.Hostname]; ok {
				host.LastSeen = time.Now()
				s.hosts[msg.Hostname] = host
			}
			s.mu.Unlock()

			log.Printf("Received %s from %s", msg.Type, msg.Hostname)

			// Handle different message types
			s.handleMessage(conn, &msg)
		}
	}

	log.Printf("Host disconnected: %s", hostname)
}

func (s *Server) handleMessage(conn *websocket.Conn, msg *Message) {
	// Handle heartbeat
	if msg.Type == "heartbeat" {
		// Acknowledge heartbeat
		ack := Message{
			Type:      "heartbeat:ack",
			Hostname:  msg.Hostname,
			Timestamp: time.Now().Unix(),
		}

		if err := conn.WriteMessage(websocket.TextMessage, marshalJSON(ack)); err != nil {
			log.Printf("Failed to send ACK: %v", err)
		}
	}

	// Future: Handle other message types
	// - vm:start
	// - vm:stop
	// - vm:status
	// - status (query)
	// - log (stream)
}

func (s *Server) GetHosts() []HostInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return copy of hosts
	hosts := make([]HostInfo, 0, len(s.hosts))
	for _, h := range s.hosts {
		hosts = append(hosts, h)
	}
	return hosts
}

func (s *Server) GetHost(hostname string) (*HostInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	host, ok := s.hosts[hostname]
	return &host, ok
}

// marshalJSON marshals an object to JSON
func marshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return data
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/ws" {
		s.HandleWebSocket(w, r)
	} else {
		http.NotFound(w, r)
	}
}