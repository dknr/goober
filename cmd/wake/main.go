package wake

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dknr/goober/internal/config"
	"github.com/dknr/goober/internal/http"
	"github.com/spf13/cobra"
)

// ClientConfig is the client configuration structure
type ClientConfig = config.ClientConfig

func NewCommand(cfg *ClientConfig) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "wake <hostname>",
		Short: "Wake a host via Wake-on-LAN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg == nil {
				return fmt.Errorf("configuration not loaded")
			}

			hostname = args[0]

			// Connect to control daemon
			client, err := http.NewClient(cfg.Server.Listen)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer client.Close()

			// Send wake command
			result, err := client.SendWake(hostname)
			if err != nil {
				return fmt.Errorf("wake failed: %w", err)
			}

			// Inform user and start waiting for the host to appear online
			fmt.Printf("%s\n", result)
			fmt.Printf("Waiting up to 30 seconds for host '%s' to come online...\n", hostname)

			deadline := time.Now().Add(30 * time.Second)
			for time.Now().Before(deadline) {
				hosts, err := client.GetHosts()
				if err != nil {
					// non‑fatal, just retry after a short pause
					time.Sleep(2 * time.Second)
					continue
				}

				// Check if host is in response and has a valid last_seen timestamp
				rawStatus, ok := hosts[hostname]
				if !ok {
					// Host not in response yet
					time.Sleep(2 * time.Second)
					continue
				}

				var hostStatus struct {
					LastSeen time.Time `json:"last_seen"`
				}
				if err := json.Unmarshal(rawStatus, &hostStatus); err != nil {
					// Failed to parse, wait and retry
					time.Sleep(2 * time.Second)
					continue
				}

				// Check if LastSeen is not zero (host has actually connected)
				if !hostStatus.LastSeen.IsZero() {
					fmt.Printf("✅ Host %s is now online.\n", hostname)
					return nil
				}

				// Host exists but hasn't connected yet, wait and retry
				time.Sleep(2 * time.Second)
			}

			fmt.Printf("⚠️ Host %s did not appear online after 30 seconds.\n", hostname)
			return nil
		},
	}

	return cmd
}