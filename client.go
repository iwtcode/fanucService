package fanucService

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ClientAPI interface {
	// Connection methods
	CreateConnection(ctx context.Context, req ConnectionRequest) (*MachineDTO, error)
	GetConnections(ctx context.Context) ([]MachineDTO, error)
	CheckConnection(ctx context.Context, machineID string) (*MachineDTO, error)
	DeleteConnection(ctx context.Context, machineID string) error

	// Polling methods
	StartPolling(ctx context.Context, machineID string, intervalMs int) error
	StopPolling(ctx context.Context, machineID string) error

	// Program methods
	GetControlProgram(ctx context.Context, machineID string) (string, error)
}

// Client реализует ClientAPI.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient создает новый экземпляр клиента.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

// --- Внутренние структуры для распаковки JSON-ответов API ---

type baseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type responseSingle struct {
	baseResponse
	Data MachineDTO `json:"data"`
}

type responseMulti struct {
	baseResponse
	Data []MachineDTO `json:"data"`
}

// --- Базовый метод запроса ---

func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	fullURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp baseResponse
		if jsonErr := json.Unmarshal(respBytes, &errResp); jsonErr == nil && errResp.Message != "" {
			return fmt.Errorf("api error (%d): %s", resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("api error (%d): %s", resp.StatusCode, string(respBytes))
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(respBytes, result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// --- Реализация методов API ---

func (c *Client) CreateConnection(ctx context.Context, req ConnectionRequest) (*MachineDTO, error) {
	var resp responseSingle
	if err := c.do(ctx, http.MethodPost, "/api/v1/connect", req, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) GetConnections(ctx context.Context) ([]MachineDTO, error) {
	var resp responseMulti
	if err := c.do(ctx, http.MethodGet, "/api/v1/connect", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c *Client) CheckConnection(ctx context.Context, machineID string) (*MachineDTO, error) {
	path := fmt.Sprintf("/api/v1/connect?id=%s", url.QueryEscape(machineID))
	var resp responseSingle
	if err := c.do(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) DeleteConnection(ctx context.Context, machineID string) error {
	path := fmt.Sprintf("/api/v1/connect?id=%s", url.QueryEscape(machineID))
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) StartPolling(ctx context.Context, machineID string, intervalMs int) error {
	req := StartPollingRequest{
		ID:       machineID,
		Interval: intervalMs,
	}
	return c.do(ctx, http.MethodPost, "/api/v1/polling/start", req, nil)
}

func (c *Client) StopPolling(ctx context.Context, machineID string) error {
	req := StopPollingRequest{
		ID: machineID,
	}
	return c.do(ctx, http.MethodPost, "/api/v1/polling/stop", req, nil)
}

func (c *Client) GetControlProgram(ctx context.Context, machineID string) (string, error) {
	path := fmt.Sprintf("/api/v1/program?id=%s", url.QueryEscape(machineID))
	fullURL := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp baseResponse
		if jsonErr := json.Unmarshal(bodyBytes, &errResp); jsonErr == nil && errResp.Message != "" {
			return "", fmt.Errorf("api error: %s", errResp.Message)
		}
		return "", fmt.Errorf("api returned status %d", resp.StatusCode)
	}

	return string(bodyBytes), nil
}
