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
	upgrader      websocket.Upgrader
	hosts         map[string]HostInfo
	configured    map[string]bool
	mu            sync.RWMutex
}

type HostInfo struct {
	LastSeen time.Time `json:"last_seen"`
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
		hosts:    make(map[string]HostInfo),
		configured: make(map[string]bool),
	}
}

// SetConfiguredHosts sets the list of configured hostnames
func (s *Server) SetConfiguredHosts(hosts []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configured = make(map[string]bool)
	for _, h := range hosts {
		s.configured[h] = true
	}
}

// GetAllHosts returns a map of all configured hostnames to HostInfo (or null for offline)
func (s *Server) GetAllHosts() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]interface{})
	for hostname := range s.configured {
		if info, ok := s.hosts[hostname]; ok {
			result[hostname] = info
		} else {
			result[hostname] = nil
		}
	}
	return result
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
		LastSeen: time.Now(),
	}
	s.mu.Unlock()

	log.Printf("Host connected: %s", hostname)

	// Track the actual hostname seen on the connection (may differ from query param)
	activeHostname := hostname

	// Read messages from client
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection error for %s: %v", activeHostname, err)
			s.mu.Lock()
			delete(s.hosts, activeHostname)
			s.mu.Unlock()
			break
		}

		if messageType == websocket.TextMessage {
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to unmarshal message from %s: %v", activeHostname, err)
				continue
			}

			// Update active hostname based on the first valid message
			if activeHostname == "unknown" || activeHostname != msg.Hostname {
				activeHostname = msg.Hostname
			}

			// Update last seen timestamp (or create entry)
			s.mu.Lock()
			if host, ok := s.hosts[msg.Hostname]; ok {
				host.LastSeen = time.Now()
				s.hosts[msg.Hostname] = host
			} else {
				// First heartbeat from this host – create entry
				s.hosts[msg.Hostname] = HostInfo{
					LastSeen: time.Now(),
				}
			}

			s.mu.Unlock()

			log.Printf("Received %s from %s", msg.Type, msg.Hostname)

			// Handle different message types
			s.handleMessage(conn, &msg)
		}
	}

	log.Printf("Host disconnected: %s", activeHostname)
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

// GetConnectedHosts returns a slice of hostnames that have sent a heartbeat (i.e., are online).
func (s *Server) GetConnectedHosts() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.hosts))
	for name := range s.hosts {
		names = append(names, name)
	}
	return names
}

// GetHostsMap returns a map of hostname to HostInfo for all connected hosts.
func (s *Server) GetHostsMap() map[string]HostInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	hosts := make(map[string]HostInfo, len(s.hosts))
	for hostname, info := range s.hosts {
		hosts[hostname] = info
	}
	return hosts
}

// RemoveHost deletes a host entry (used to clear stale data after a wake request).
func (s *Server) RemoveHost(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.hosts, name)
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