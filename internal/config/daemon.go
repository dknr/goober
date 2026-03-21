package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type ControlDaemonConfig struct {
	Server ServerConfig `toml:"server"`
	Hosts  map[string]HostConfig `toml:"host"`
	// Future: database path, auth, etc.
}

type ServerConfig struct {
	Listen string `toml:"listen"`
	Address string `toml:"address"`
}

type HostConfig struct {
	Name  string `toml:"name"`
	MAC   string `toml:"mac"`
	// Future: IP, WoL port, etc.
}

type HostDaemonConfig struct {
	Server    ServerConfig    `toml:"server"`
	Name      string          `toml:"name"`
	Control   ControlConfig   `toml:"control"`
	Heartbeat HeartbeatConfig `toml:"heartbeat"`
}

type ControlConfig struct {
	Address string `toml:"address"`
	Path    string `toml:"path"`
}

type HeartbeatConfig struct {
	Interval int `toml:"interval"`
}

// LoadControlDaemonConfig loads the control daemon configuration
func LoadControlDaemonConfig(path string) (*ControlDaemonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ControlDaemonConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Server.Listen == "" {
		cfg.Server.Listen = "127.0.0.1:29530"
	}

	// Validate hosts
	for name, host := range cfg.Hosts {
		if host.MAC == "" {
			return nil, fmt.Errorf("host '%s' has no MAC address", name)
		}
	}

	return &cfg, nil
}

// LoadHostDaemonConfig loads the host daemon configuration
func LoadHostDaemonConfig(path string) (*HostDaemonConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg HostDaemonConfig
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Server.Listen == "" {
		cfg.Server.Listen = "127.0.0.1:29532"
	}

	if cfg.Name == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("failed to get hostname: %w", err)
		}
		cfg.Name = hostname
	}

	if cfg.Control.Address == "" {
		cfg.Control.Address = "127.0.0.1:29530"
	}

	if cfg.Control.Path == "" {
		cfg.Control.Path = "/ws"
	}

	if cfg.Heartbeat.Interval == 0 {
		cfg.Heartbeat.Interval = 60 // Default 60 seconds
	}

	return &cfg, nil
}

// LoadClientConfig loads the client configuration (supports a [server] table)
func LoadClientConfig(path string) (*ClientConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Don't fail if config doesn't exist yet - return defaults
		return &ClientConfig{
			Server: ServerConfig{
				Listen: "127.0.0.1:29530",
			},
		}, nil
	}

	// The config file may contain a top‑level "server" table. Decode into a wrapper.
	type wrapper struct {
		Server ServerConfig `toml:"server"`
	}

	var w wrapper
	if _, err := toml.Decode(string(data), &w); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg := w.Server

	// Support both "address" and "listen" field names
	if cfg.Address != "" && cfg.Listen == "" {
		cfg.Listen = cfg.Address
	}

	if cfg.Listen == "" {
		cfg.Listen = "127.0.0.1:29530"
	}

	return &ClientConfig{
		Server: cfg,
	}, nil
}

type ClientConfig struct {
	Server ServerConfig `toml:"server"`
}