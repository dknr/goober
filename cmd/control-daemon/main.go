package controldaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/logging"
	"github.com/lore/goober/internal/websocket"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "control-daemon",
		Short: "Run the control daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if configFile == "" {
				configFile = "gbr-control.toml"
			}

			// Load configuration
			cfg, err := config.LoadControlDaemonConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize logging
			logger := logging.NewLogger("info")

			logger.Info("Starting gbr control-daemon...")
			logger.Infof("Listening on %s", cfg.Server.Listen)

			// Create HTTP server with routes
			mux := http.NewServeMux()

			// Register WebSocket endpoint
			wsServer := websocket.NewServer()
			mux.Handle("/ws", wsServer)

			// Set configured hostnames
			var configuredHosts []string
			for hostname := range cfg.Hosts {
				configuredHosts = append(configuredHosts, hostname)
			}
			wsServer.SetConfiguredHosts(configuredHosts)

			// Helper function to send JSON response
			sendResponse := func(w http.ResponseWriter, message string, statusCode int) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)
				json.NewEncoder(w).Encode(map[string]string{"message": message})
			}

			// Register /api/hosts endpoint (returns all configured hosts with last_seen for online ones)
			mux.HandleFunc("/api/hosts", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				hosts := wsServer.GetAllHosts()
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(hosts)
			})

			// Register /api/wake endpoint
			mux.HandleFunc("/api/wake", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					sendResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
					return
				}

				// Read request body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					sendResponse(w, "Failed to read request", http.StatusBadRequest)
					return
				}
				defer r.Body.Close()

				// Parse request
				var req struct {
					Host string `json:"host"`
				}
				if err := json.Unmarshal(body, &req); err != nil {
					sendResponse(w, "Invalid request body", http.StatusBadRequest)
					return
				}

				// Get host config
				host, ok := cfg.Hosts[req.Host]
				if !ok {
					sendResponse(w, fmt.Sprintf("Host '%s' not found", req.Host), http.StatusNotFound)
					return
				}

				// Check if wakeonlan is available
				if _, err := exec.LookPath("wakeonlan"); err != nil {
					logger.Errorf("wakeonlan not found: %v", err)
					sendResponse(w, "wakeonlan not found - please install it", http.StatusInternalServerError)
					return
				}

				// Call wakeonlan to send the packet
				cmd := exec.Command("wakeonlan", host.MAC)
				if err := cmd.Run(); err != nil {
					logger.Errorf("Failed to send WoL packet with wakeonlan for host %s: %v", req.Host, err)
					sendResponse(w, fmt.Sprintf("Failed to send WoL packet: %v", err), http.StatusInternalServerError)
					return
				}

				logger.Infof("Sent WoL packet to host: %s (MAC: %s)", req.Host, host.MAC)
			// Clear any stale online entry so client must wait for fresh heartbeat.
			wsServer.RemoveHost(req.Host)
				sendResponse(w, fmt.Sprintf("Sent wake-on-lan packet to %s (%s)", req.Host, host.MAC), http.StatusOK)
			})

			// Start server
			srv := &http.Server{
				Addr:    cfg.Server.Listen,
				Handler: mux,
			}

			if err := srv.ListenAndServe(); err != nil {
				return fmt.Errorf("server failed: %w", err)
			}

			// Keep running
			select {}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")

	return cmd
}