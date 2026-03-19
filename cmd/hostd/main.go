package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lore/goober/internal/config"
	"github.com/lore/goober/internal/logging"
	"github.com/lore/goober/internal/transport"
)

func main() {
	configPath := flag.String("config", "gbr-hostd.toml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadHostDaemonConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging
	logger := logging.NewLogger(cfg.LogLevel)

	logger.Info("Starting gbr-hostd...")
	logger.Infof("Listening on %s", cfg.Server.Address)

	// Create and start transport server
	server := transport.NewServer(cfg, logger)

	if err := server.Start(); err != nil {
		logger.Errorf("Server failed: %v", err)
		os.Exit(1)
	}

	// Keep running
	select {}
}