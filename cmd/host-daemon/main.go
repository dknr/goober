package hostdaemon

import (
	"fmt"
	"time"
	"strings"

	"github.com/dknr/goober/internal/config"
	"github.com/dknr/goober/internal/logging"
	"github.com/dknr/goober/internal/websocket"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "host-daemon",
		Short: "Run the host daemon on FreeBSD hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if configFile == "" {
				configFile = "goober-host.toml"
			}

			// Load configuration
			cfg, err := config.LoadHostDaemonConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize logging
			logger := logging.NewLogger("info")

			logger.Info("Starting gbr host-daemon...")
			logger.Infof("Hostname: %s", cfg.Name)
			logger.Infof("Heartbeat interval: %d seconds", cfg.Heartbeat.Interval)
			logger.Infof("Connecting to control daemon at ws://%s%s", cfg.Control.Address, cfg.Control.Path)

			// Create WebSocket client
			wsClient := websocket.NewClient(fmt.Sprintf("ws://%s%s", cfg.Control.Address, cfg.Control.Path))

// Main loop
	for {
				// Connect to control daemon
				if err := wsClient.Connect(); err != nil {
					logger.Errorf("Connection failed: %v", err)
					// Reconnect with exponential backoff (max 2 minutes)
					if err := wsClient.Reconnect(2 * time.Minute); err != nil {
						logger.Errorf("Reconnection failed: %v", err)
						time.Sleep(2 * time.Minute)
					}
					continue
				}

				// Connected successfully
				logger.Info("Connected to control daemon")

				// Heartbeat loop
				heartbeatInterval := time.Duration(cfg.Heartbeat.Interval) * time.Second
				ticker := time.NewTicker(heartbeatInterval)
				defer ticker.Stop()

			ConnectionLoop:
				for {
					select {
					case <-ticker.C:
						shortName := cfg.Name
						if idx := strings.IndexByte(cfg.Name, '.'); idx != -1 {
							shortName = cfg.Name[:idx]
						}
						if err := wsClient.SendHeartbeat(shortName); err != nil {
							logger.Errorf("Failed to send heartbeat: %v", err)
							break ConnectionLoop
						}

					case <-time.After(30 * time.Second):
						// Check connection every 30 seconds
						if wsClient.IsClosed() {
							logger.Info("Connection closed, reconnecting...")
							break ConnectionLoop
						}

					case msg := <-readMessages(wsClient, logger):
						if msg == nil {
							logger.Info("Connection lost, reconnecting...")
							break ConnectionLoop
						}
						logger.Debugf("Received: %s", msg.Type)
					}
				}

				// Reconnect
				logger.Info("Reconnecting...")
				if err := wsClient.Reconnect(2 * time.Minute); err != nil {
					logger.Errorf("Reconnection failed: %v", err)
					time.Sleep(2 * time.Minute)
				}
			}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")

	return cmd
}

// readMessages returns a channel for reading messages from WebSocket
// Returns nil message when connection is lost or error occurs
func readMessages(ws *websocket.Client, logger *logging.Logger) <-chan *websocket.Message {
	msgChan := make(chan *websocket.Message, 1)

	go func() {
		defer close(msgChan)
		for {
			msg, err := ws.ReceiveMessage()
			if err != nil {
				// Log the error for debugging
				logger.Errorf("WebSocket read error: %v", err)
				return
			}
			msgChan <- msg
		}
	}()

	return msgChan
}