package controldaemon

import (
	"fmt"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/logging"
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

			// Create and start HTTP server
			server := NewServer(cfg, logger)
			if err := server.Start(); err != nil {
				return fmt.Errorf("server failed: %w", err)
			}

			// Keep running
			select {}
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "Path to configuration file")

	return cmd
}