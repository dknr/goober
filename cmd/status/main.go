package status

import (
	"fmt"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/http"
	"github.com/spf13/cobra"
)

// ClientConfig is the client configuration structure
type ClientConfig = config.ClientConfig

func NewCommand(cfg *ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of VMs and hosts",
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

			// Send status command
			result, err := client.SendCommand("status", "")
			if err != nil {
				return fmt.Errorf("status failed: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}