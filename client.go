package fanucService

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ClientAPI defines the interface for interacting with the Fanuc Service
type ClientAPI interface {
	CreateConnection(ctx context.Context, req ConnectionRequest) (*ConnectionResponse, error)
	GetConnections(ctx context.Context) (*ConnectionResponse, error)
	CheckConnection(ctx context.Context, machineID string) (*ConnectionResponse, error)
	DeleteConnection(ctx context.Context, machineID string) (*ConnectionResponse, error)
}

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*ConnectionResponse, error) {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ConnectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return &result, fmt.Errorf("api error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) CreateConnection(ctx context.Context, req ConnectionRequest) (*ConnectionResponse, error) {
	return c.do(ctx, http.MethodPost, "/api/v1/connect", req)
}

func (c *Client) GetConnections(ctx context.Context) (*ConnectionResponse, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/connect", nil)
}

func (c *Client) CheckConnection(ctx context.Context, machineID string) (*ConnectionResponse, error) {
	return c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/connect?id=%s", machineID), nil)
}

func (c *Client) DeleteConnection(ctx context.Context, machineID string) (*ConnectionResponse, error) {
	return c.do(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/connect?id=%s", machineID), nil)
}
