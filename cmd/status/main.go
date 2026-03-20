package status

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/http"
	"github.com/spf13/cobra"
)

// ClientConfig is the client configuration structure
type ClientConfig = config.ClientConfig

func NewCommand(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of managed hosts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg == nil {
				return fmt.Errorf("configuration not loaded")
			}

			// Connect to control daemon
			client, err := http.NewClient(cfg.Server.Address)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer client.Close()

			// Get hosts
			rawHosts, err := client.GetHosts()
			if err != nil {
				return fmt.Errorf("failed to get hosts: %w", err)
			}

			// Build status table
			type HostStatus struct {
				LastSeen time.Time `json:"last_seen"`
			}
			hostnameLen := 0
			for hostname := range rawHosts {
				if len(hostname) > hostnameLen {
					hostnameLen = len(hostname)
				}
			}
			hostnameLen += 2 // Pad

			fmt.Printf("%-*s %-9s %s\n", hostnameLen, "NAME", "STATUS", "LAST SEEN")
			fmt.Printf("%s\n", strings.Repeat("-", hostnameLen+9+10))

			for hostname, status := range rawHosts {
				if len(status) == 0 {
					fmt.Printf("%-*s %-8s %s\n", hostnameLen, hostname, "offline", "never")
				} else {
					// Unmarshal status into HostStatus struct
					var hostStatus HostStatus
					if err := json.Unmarshal(status, &hostStatus); err != nil {
						fmt.Printf("%-*s %-8s %s\n", hostnameLen, hostname, "error", err.Error())
						continue
					}
					if hostStatus.LastSeen.IsZero() {
						fmt.Printf("%-*s %-8s %s\n", hostnameLen, hostname, "offline", "never")
					} else {
						ago := time.Since(hostStatus.LastSeen)
						var status string
						if ago < 2*time.Minute {
							status = "online"
						} else if ago < 5*time.Minute {
							status = "unknown"
						} else {
							status = "offline"
						}
						fmt.Printf("%-*s %-9s %s\n", hostnameLen, hostname, status, hostStatus.LastSeen.Format("2006-01-02 15:04:05"))
					}
				}
			}

			return nil
		},
	}

	return cmd
}