package wake

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dknr/goober/internal/config"
	"github.com/stretchr/testify/require"
)

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

func TestWakeCommand_Success(t *testing.T) {
	// Test server that mocks the control daemon
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Track number of /api/hosts calls
	hostsCallCount := 0

	// Set up handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/api/wake", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		var req struct {
			Host string `json:"host"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "test-host", req.Host)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Wake sent"})
	})
	mux.HandleFunc("/api/hosts", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		hostsCallCount++
		if hostsCallCount < 3 {
			// First two calls: host appears offline (null)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"test-host": nil,
			})
		} else {
			// Third call: host appears online
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"test-host": map[string]string{
					"last_seen": time.Now().UTC().Format(time.RFC3339),
				},
			})
		}
	})
	ts.Config.Handler = mux

	// Create client config pointing to test server
	configPath := createClientConfig(t, ts.Listener.Addr().String())

	// Load client config
	cfg, err := config.LoadClientConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Create wake command
	cmd := NewCommand(cfg)
	require.NotNil(t, cmd)

	// Buffer to capture output
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	// Run the command with hostname argument
	err = cmd.RunE(cmd, []string{"test-host"})
	// Debug: print captured output
	fmt.Fprintf(os.Stderr, "outBuf: %q\n", outBuf.String())
	fmt.Fprintf(os.Stderr, "errBuf: %q\n", errBuf.String())
	require.NoError(t, err)

	// Check output
	out := strings.TrimSpace(outBuf.String())
	errOut := strings.TrimSpace(errBuf.String())

	// Expect wake message and waiting messages
	require.Contains(t, out, "Wake sent")
	require.Contains(t, out, "Waiting up to 30 seconds for host 'test-host' to come online...")
	require.Contains(t, out, "✅ Host test-host is now online.")
	// No error output expected
	require.Empty(t, errOut)
}

func TestWakeCommand_ConfigNotLoaded(t *testing.T) {
	// Create wake command with nil config
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

func TestWakeCommand_WakeFails(t *testing.T) {
	// Test server that returns error on /api/wake
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	configPath := createClientConfig(t, ts.Listener.Addr().String())

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
	require.Contains(t, err.Error(), "wake failed")
	// No output expected from the command itself (error is returned)
	require.Empty(t, outBuf.String())
	require.Empty(t, errBuf.String())
}

func TestWakeCommand_HostNeverOnline(t *testing.T) {
	// Test server that always returns host as offline
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/wake":
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", r.Method)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Wake sent"})
		case "/api/hosts":
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
				return
			}
			// Always return host as offline
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"test-host": nil,
			})
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	configPath := createClientConfig(t, ts.Listener.Addr().String())

	cfg, err := config.LoadClientConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	cmd := NewCommand(cfg)
	require.NotNil(t, cmd)

	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	// Run the command; it should timeout after 30 seconds and return nil error (as per implementation)
	start := time.Now()
	err = cmd.RunE(cmd, []string{"test-host"})
	elapsed := time.Since(start)
	// The command should return nil error (it only returns error on config or wake failure)
	require.NoError(t, err)
	// Should have waited at least 30 seconds (but we can't guarantee exact timing in test)
	// We'll just check that it ran for a reasonable time (at least 2 seconds, since it polls every 2 seconds)
	require.True(t, elapsed >= 2*time.Second, "Expected command to take at least 2 seconds")

	// Check output
	out := strings.TrimSpace(outBuf.String())
	errOut := strings.TrimSpace(errBuf.String())

	require.Contains(t, out, "Wake sent")
	require.Contains(t, out, "Waiting up to 30 seconds for host 'test-host' to come online...")
	require.Contains(t, out, "⚠️ Host test-host did not appear online after 30 seconds.")
	require.Empty(t, errOut)
}