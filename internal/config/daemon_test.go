package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadControlDaemonConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "control.toml")

	// Test valid config
	validConfig := `
[server]
listen = "127.0.0.1:29530"

[host.home-server]
mac = "00:11:22:33:44:55"

[host.office-server]
mac = "aa:bb:cc:dd:ee:ff"
`
	err := os.WriteFile(configPath, []byte(validConfig), 0644)
	require.NoError(t, err)

	cfg, err := LoadControlDaemonConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "127.0.0.1:29530", cfg.Server.Listen)
	require.Len(t, cfg.Hosts, 2)
	require.Equal(t, "00:11:22:33:44:55", cfg.Hosts["home-server"].MAC)
	require.Equal(t, "aa:bb:cc:dd:ee:ff", cfg.Hosts["office-server"].MAC)
}

func TestLoadControlDaemonConfig_DefaultListen(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "control.toml")

	// Config without listen
	config := `
[host.home-server]
mac = "00:11:22:33:44:55"
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	cfg, err := LoadControlDaemonConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:29530", cfg.Server.Listen) // default
}

func TestLoadControlDaemonConfig_MissingMAC(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "control.toml")

	// Config with missing MAC
	config := `
[server]
listen = "127.0.0.1:29530"

[host.home-server]
# mac is missing
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	_, err = LoadControlDaemonConfig(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host 'home-server' has no MAC address")
}

func TestLoadControlDaemonConfig_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "control.toml")

	// Invalid TOML
	err := os.WriteFile(configPath, []byte("invalid toml"), 0644)
	require.NoError(t, err)

	_, err = LoadControlDaemonConfig(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse config")
}

func TestLoadHostDaemonConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "host.toml")

	// Valid host daemon config
	config := `
[server]
listen = "127.0.0.1:29532"

[control]
address = "127.0.0.1:29530"
path = "/ws"

[heartbeat]
interval = 30
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	cfg, err := LoadHostDaemonConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "127.0.0.1:29532", cfg.Server.Listen)
	require.Equal(t, "127.0.0.1:29530", cfg.Control.Address)
	require.Equal(t, "/ws", cfg.Control.Path)
	require.Equal(t, 30, cfg.Heartbeat.Interval)
	// Name should be set from hostname (we'll mock this later)
}

func TestLoadHostDaemonConfig_Defaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "host.toml")

	// Minimal config
	config := `
[server]
# listen is missing
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	cfg, err := LoadHostDaemonConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:29532", cfg.Server.Listen) // default
	require.Equal(t, "127.0.0.1:29530", cfg.Control.Address) // default
	require.Equal(t, "/ws", cfg.Control.Path) // default
	require.Equal(t, 60, cfg.Heartbeat.Interval) // default
	// Name should be set to hostname
}

func TestLoadHostDaemonConfig_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "host.toml")

	err := os.WriteFile(configPath, []byte("invalid toml"), 0644)
	require.NoError(t, err)

	_, err = LoadHostDaemonConfig(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse config")
}

func TestLoadClientConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.toml")

	// Valid client config
	config := `
[server]
address = "127.0.0.1:29530"
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	cfg, err := LoadClientConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "127.0.0.1:29530", cfg.Server.Listen)
	require.Equal(t, "127.0.0.1:29530", cfg.Server.Address) // address is copied to listen
}

func TestLoadClientConfig_ListenField(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.toml")

	// Using listen field
	config := `
[server]
listen = "127.0.0.1:29530"
`
	err := os.WriteFile(configPath, []byte(config), 0644)
	require.NoError(t, err)

	cfg, err := LoadClientConfig(configPath)
	require.NoError(t, err)
	require.Equal(t, "127.0.0.1:29530", cfg.Server.Listen)
}

func TestLoadClientConfig_DefaultWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.toml")

	// Empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	cfg, err := LoadClientConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "127.0.0.1:29530", cfg.Server.Listen) // default
}

func TestLoadClientConfig_InvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "client.toml")

	err := os.WriteFile(configPath, []byte("invalid toml"), 0644)
	require.NoError(t, err)

	_, err = LoadClientConfig(configPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse config")
}