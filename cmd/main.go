package main

import (
	"fmt"
	"os"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/http"
	"github.com/lore/goober/cmd/control-daemon"
	"github.com/lore/goober/cmd/host-daemon"
	"github.com/lore/goober/cmd/status"
	"github.com/lore/goober/cmd/version"
	"github.com/lore/goober/cmd/wake"
	"github.com/spf13/cobra"
)

func main() {
	// Load client config
	cfg, _ := config.LoadClientConfig("gbr-client.toml")

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "gbr",
		Short: "Goober - FreeBSD Orchestration Tool",
		Run:   showHelp,
	}

	// Register subcommands
	rootCmd.AddCommand(controldaemon.NewCommand())
	rootCmd.AddCommand(hostdaemon.NewCommand())
	rootCmd.AddCommand(status.NewCommand(cfg))
	rootCmd.AddCommand(version.NewCommand())
	rootCmd.AddCommand(wake.NewCommand())

	// Run
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp(cmd *cobra.Command, args []string) {
	fmt.Println("Goober - FreeBSD Orchestration Tool")
	fmt.Println("Use `gbr [command] [subcommand]` to see available commands")
	fmt.Println()
	cmd.Help()
}