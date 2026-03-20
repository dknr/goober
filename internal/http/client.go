package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	serverURL string
	client    *http.Client
}

func NewClient(serverURL string) (*Client, error) {
	return &Client{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *Client) Close() error {
	// No resources to close
	return nil
}

func (c *Client) buildURL(path string) string {
	return "http://" + c.serverURL + path
}

func (c *Client) SendWake(host string) (string, error) {
	reqBody := map[string]string{
		"host": host,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.buildURL("/api/wake")

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send wake request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("wake request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var respBody struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return respBody.Message, nil
}

func (c *Client) SendCommand(command string, param string) (string, error) {
	// TODO: Implement other commands (list, stop, status)
	if command == "wake" {
		return c.SendWake(param)
	}

	return "", fmt.Errorf("command not implemented: %s", command)
}

// GetHosts contacts the control daemon and returns a map of hostname to JSON raw message.
// Online hosts have last_seen, offline hosts have null.
func (c *Client) GetHosts() (map[string]json.RawMessage, error) {
	url := c.buildURL("/api/hosts")
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get hosts failed (status %d): %s", resp.StatusCode, string(body))
	}

	var hosts map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&hosts); err != nil {
		return nil, fmt.Errorf("failed to decode hosts: %w", err)
	}

	return hosts, nil
}
