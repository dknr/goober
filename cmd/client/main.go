package main

import (
	"fmt"
	"os"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/transport"
	"github.com/spf13/cobra"
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
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			configFile = fmt.Sprintf("%s/goober/goober.toml", home)
		}
	}

	// Load client config
	cfg, err := config.LoadClientConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "gbr",
		Short: "Goober - FreeBSD Orchestration Tool",
		// Long:  A longer description that will be shown on help text.
	}

	// Add subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(statusCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a VM or host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := transport.NewClient(cfg, cfg.Server.Address)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		result, err := client.SendCommand("start", args[0])
		if err != nil {
			return fmt.Errorf("start failed: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a VM or host",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := transport.NewClient(cfg, cfg.Server.Address)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		result, err := client.SendCommand("stop", args[0])
		if err != nil {
			return fmt.Errorf("stop failed: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List VMs and hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := transport.NewClient(cfg, cfg.Server.Address)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		result, err := client.SendCommand("list", "")
		if err != nil {
			return fmt.Errorf("list failed: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of VMs and hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := transport.NewClient(cfg, cfg.Server.Address)
		if err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
		defer client.Close()

		result, err := client.SendCommand("status", "")
		if err != nil {
			return fmt.Errorf("status failed: %w", err)
		}

		fmt.Println(result)
		return nil
	},
}