package controldaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"sync"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/logging"
	"github.com/lore/goober/internal/websocket"
)

type Server struct {
	config    *config.ControlDaemonConfig
	logger    *logging.Logger
	hosts     map[string]config.HostConfig
	connected map[string]bool // tracks hosts that have sent a heartbeat
	mu        sync.RWMutex
	http      *http.Server
	mux       *http.ServeMux
	wsServer  *websocket.Server
}

func NewServer(cfg *config.ControlDaemonConfig, logger *logging.Logger) *Server {
	return &Server{
		config:    cfg,
		logger:    logger,
		hosts:     cfg.Hosts,
		connected: make(map[string]bool),
		mux:       http.NewServeMux(),
		wsServer:  websocket.NewServer(),
	}
}

func (s *Server) Start() error {
	// Register WebSocket endpoint
	s.mux.Handle("/ws", s.wsServer)

	// Set configured hostnames
	var configuredHosts []string
	for hostname := range s.config.Hosts {
		configuredHosts = append(configuredHosts, hostname)
	}
	s.wsServer.SetConfiguredHosts(configuredHosts)

	// Register /api/hosts endpoint (returns all configured hosts with last_seen for online ones)
	s.mux.HandleFunc("/api/hosts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		hosts := s.wsServer.GetAllHosts()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hosts)
	})

	// Register /api/wake endpoint
	s.mux.HandleFunc("/api/wake", s.handleWake)

	// Start server
	s.http = &http.Server{
		Addr:    s.config.Server.Listen,
		Handler: s.mux,
	}

	s.logger.Infof("HTTP server listening on %s", s.config.Server.Listen)
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
	host, ok := s.config.Hosts[req.Host]
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
	// Clear any stale online entry so client must wait for fresh heartbeat.
	s.wsServer.RemoveHost(req.Host)
	s.respond(w, fmt.Sprintf("Sent wake-on-lan packet to %s (%s)", req.Host, host.MAC), http.StatusOK)
}

func (s *Server) respond(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}