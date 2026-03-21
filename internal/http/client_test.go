package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("example.com:8080")
	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "example.com:8080", client.serverURL)
	require.NotNil(t, client.client)
	// Check that timeout is set to 10 seconds
	require.Equal(t, 10*time.Second, client.client.Timeout)
}

func TestSendWake_Success(t *testing.T) {
	// Create a test server that responds with a success message
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method, "POST method expected")
		require.Equal(t, "/api/wake", r.URL.Path, "Path should be /api/wake")
		require.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type header")

		// Decode request body
		var req struct {
			Host string `json:"host"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		require.Equal(t, "test-host", req.Host, "Host in request body")

		// Write success response
		resp := map[string]string{"message": "Wake sent"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	// Create client pointing to test server
	client, err := NewClient(ts.URL[len("http://"):]) // Remove http:// prefix
	require.NoError(t, err)

	// Call SendWake
	message, err := client.SendWake("test-host")
	require.NoError(t, err)
	require.Equal(t, "Wake sent", message)
}

func TestSendWake_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL[len("http://"):])
	require.NoError(t, err)

	_, err = client.SendWake("test-host")
	require.Error(t, err)
	require.Contains(t, err.Error(), "wake request failed (status 500)")
	require.Contains(t, err.Error(), "Internal Server Error")
}

func TestSendWake_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write invalid JSON
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL[len("http://"):])
	require.NoError(t, err)

	_, err = client.SendWake("test-host")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode response")
}

func TestSendCommand_Wake(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/wake", r.URL.Path)
		resp := map[string]string{"message": "Wake sent via SendCommand"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL[len("http://"):])
	require.NoError(t, err)

	message, err := client.SendCommand("wake", "test-host")
	require.NoError(t, err)
	require.Equal(t, "Wake sent via SendCommand", message)
}

func TestSendCommand_Unimplemented(t *testing.T) {
	client, err := NewClient("example.com:8080")
	require.NoError(t, err)

	_, err = client.SendCommand("stop", "test-host")
	require.Error(t, err)
	require.Contains(t, err.Error(), "command not implemented: stop")
}

func TestGetHosts_Success(t *testing.T) {
	// Prepare a response map that mimics the expected format
	// Online host with last_seen, offline host with null
	onlineHost := map[string]interface{}{
		"last_seen": "2026-03-21T20:14:12Z",
	}
	onlineBytes, err := json.Marshal(onlineHost)
	require.NoError(t, err)

	response := map[string]json.RawMessage{
		"online-host": onlineBytes,
		"offline-host": nil, // nil marshales to null in JSON
	}
	responseBytes, err := json.Marshal(response)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/api/hosts", r.URL.Path)
		w.Write(responseBytes)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL[len("http://"):])
	require.NoError(t, err)

	hosts, err := client.GetHosts()
	require.NoError(t, err)
	require.NotNil(t, hosts)
	require.Len(t, hosts, 2)

	// Check online host
	onlineData, ok := hosts["online-host"]
	require.True(t, ok, "online-host should be present")
	require.NotNil(t, onlineData, "online-host data should not be nil")
	var onlineHostStatus struct {
		LastSeen string `json:"last_seen"`
	}
	err = json.Unmarshal(onlineData, &onlineHostStatus)
	require.NoError(t, err)
	require.Equal(t, "2026-03-21T20:14:12Z", onlineHostStatus.LastSeen)

	// Check offline host
	offlineData, ok := hosts["offline-host"]
	require.True(t, ok, "offline-host should be present")
	require.Equal(t, json.RawMessage(`null`), offlineData, "offline-host data should be JSON null")
}

func TestGetHosts_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL[len("http://"):])
	require.NoError(t, err)

	_, err = client.GetHosts()
	require.Error(t, err)
	require.Contains(t, err.Error(), "get hosts failed (status 400)")
	require.Contains(t, err.Error(), "Bad Request")
}

func TestGetHosts_DecodeError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write invalid JSON
		w.Write([]byte("not json"))
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL[len("http://"):])
	require.NoError(t, err)

	_, err = client.GetHosts()
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode hosts")
}

func TestClose(t *testing.T) {
	client, err := NewClient("example.com:8080")
	require.NoError(t, err)

	err = client.Close()
	require.NoError(t, err) // Should always return nil
}