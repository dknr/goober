package main

import (
	"fmt"
	"os"

	controldaemon "github.com/dknr/goober/cmd/control-daemon"
	genkey "github.com/dknr/goober/cmd/genkey"
	stop "github.com/dknr/goober/cmd/stop"
	"github.com/dknr/goober/cmd/status"
	"github.com/dknr/goober/cmd/wake"
	"github.com/dknr/goober/internal/config"
	hostdaemon "github.com/dknr/goober/cmd/host-daemon"
	"github.com/spf13/cobra"
	versionCmd "github.com/dknr/goober/cmd/version"
)

var (
	configFile string
)

func main() {
// Set default config file
		if configFile == "" {
			home, err := os.UserConfigDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to find config directory: %v\n", err)
				os.Exit(1)
			}
			configFile = fmt.Sprintf("%s/goober/goober-client.toml", home)
		}

	// Load client config for authentication
	clientConfig, err := config.LoadClientConfig(configFile)
	if err != nil {
		// Don't fail if config doesn't exist yet
		clientConfig = nil
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "gbr",
		Short: "Goober - FreeBSD Orchestration Tool",
	}

	// Add daemon subcommands
	rootCmd.AddCommand(controldaemon.NewCommand())
	rootCmd.AddCommand(hostdaemon.NewCommand())
	rootCmd.AddCommand(genkey.NewCommand())

	// Add client subcommands
	rootCmd.AddCommand(wake.NewCommand(clientConfig))
	rootCmd.AddCommand(stop.NewCommand(clientConfig))
	rootCmd.AddCommand(status.NewCommand(clientConfig))
	// Add utility subcommands
	rootCmd.AddCommand(versionCmd.NewCommand())

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}