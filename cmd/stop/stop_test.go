package stop

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dknr/goober/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)
var _ *cobra.Command

// Helper to create a temporary client config file pointing to the given host:port (without scheme)
func createClientConfig(t *testing.T, hostPort string) string {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "goober-client.toml")
	configContent := fmt.Sprintf(`[server]
address = "%s"
`, hostPort)
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	return configPath
}

func TestStopCommand_ConfigNotLoaded(t *testing.T) {
	// Create stop command with nil config
	cmd := NewCommand(nil)
	require.NotNil(t, cmd)

	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	err := cmd.RunE(cmd, []string{"test-host"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "configuration not loaded")
	// No output expected
	require.Empty(t, outBuf.String())
	require.Empty(t, errBuf.String())
}

func TestStopCommand_StopNotImplemented(t *testing.T) {
	// Since SendCommand for "stop" is not implemented, it returns an error.
	// We can still test that the command propagates that error.

	// Create a dummy config pointing to any address (won't be used because SendCommand returns early)
	configPath := createClientConfig(t, "dummy:8080")

	cfg, err := config.LoadClientConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cmd := NewCommand(cfg)
	require.NotNil(t, cmd)

	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	err = cmd.RunE(cmd, []string{"test-host"})
	require.Error(t, err)
	// The error should contain the message from SendCommand
	require.Contains(t, err.Error(), "stop failed")
	require.Contains(t, err.Error(), "command not implemented: stop")
	// No output expected because error returned before printing
	require.Empty(t, outBuf.String())
	require.Empty(t, errBuf.String())
}
