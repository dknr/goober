package wake

import (
	"fmt"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/http"
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

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}