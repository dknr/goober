package stop

import (
	"fmt"

	"github.com/dknr/goober/internal/config"
	"github.com/dknr/goober/internal/http"
	"github.com/spf13/cobra"
)

// ClientConfig is the client configuration structure
type ClientConfig = config.ClientConfig

func NewCommand(cfg *ClientConfig) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a VM or host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg == nil {
				return fmt.Errorf("configuration not loaded")
			}

			name = args[0]

			// Connect to control daemon
			client, err := http.NewClient(cfg.Server.Address)
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer client.Close()

			// Send stop command
			result, err := client.SendCommand("stop", name)
			if err != nil {
				return fmt.Errorf("stop failed: %w", err)
			}

			fmt.Println(result)
			return nil
		},
	}

	return cmd
}