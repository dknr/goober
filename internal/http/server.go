package http

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"net/http"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/logging"
)

type Server struct {
	config    *config.ControlDaemonConfig
	logger    *logging.Logger
	hosts     map[string]config.HostConfig
	connected map[string]bool // tracks hosts that have sent a heartbeat
	http      *http.Server
}
func NewServer(cfg *config.ControlDaemonConfig, logger *logging.Logger) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		hosts:     cfg.Hosts,
		connected: make(map[string]bool),
	}
}

func (s *Server) Start() error {
	// Create HTTP server
	mux := http.NewServeMux()

	// Register endpoints
	mux.HandleFunc("/api/wake", s.handleWake)
	mux.HandleFunc("/api/hosts", s.handleHosts)

	// TODO: Mount WebSocket server here for host heartbeats
	// wsServer := websocket.NewServer()
	// mux.Handle("/ws", wsServer)
	// mux.HandleFunc("/api/hosts", wsServer.GetHosts)

	s.http = &http.Server{
		Addr:    s.config.Server.Address,
		Handler: mux,
	}

	// Start server
	s.logger.Infof("HTTP server listening on %s", s.config.Server.Address)
	return s.http.ListenAndServe()
}

func (s *Server) handleWake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.respond(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.respond(w, "Failed to read request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse request
	var req struct {
		Host string `json:"host"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		s.respond(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get host config
	host, ok := s.hosts[req.Host]
	if !ok {
		s.respond(w, fmt.Sprintf("Host '%s' not found", req.Host), http.StatusNotFound)
		return
	}

	// Check if wakeonlan is available
	if _, err := exec.LookPath("wakeonlan"); err != nil {
		s.logger.Errorf("wakeonlan not found: %v", err)
		s.respond(w, "wakeonlan not found - please install it", http.StatusInternalServerError)
		return
	}

	// Call wakeonlan to send the packet
	cmd := exec.Command("wakeonlan", host.MAC)
	if err := cmd.Run(); err != nil {
		s.logger.Errorf("Failed to send WoL packet with wakeonlan for host %s: %v", req.Host, err)
		s.respond(w, fmt.Sprintf("Failed to send WoL packet: %v", err), http.StatusInternalServerError)
		return
	}

	s.logger.Infof("Sent WoL packet to host: %s (MAC: %s)", req.Host, host.MAC)

	// Return success
	s.respond(w, fmt.Sprintf("Sent wake-on-lan packet to %s (%s)", req.Host, host.MAC), http.StatusOK)
}

func (s *Server) respond(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// handleHosts returns the list of currently connected hosts.
func (s *Server) handleHosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.respond(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build slice of connected host names
	names := make([]string, 0, len(s.connected))
	for name := range s.connected {
		names = append(names, name)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]string{"hosts": names})
}
