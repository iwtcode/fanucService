package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/iwtcode/fanucService"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type apiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func TestClient_CreateConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/connect", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody fanucService.ConnectionRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.10:8193", reqBody.Endpoint)
		assert.Equal(t, "0i", reqBody.Series)

		resp := apiResponse{
			Status: "ok",
			Data: fanucService.MachineDTO{
				ID:       "uuid-123",
				Endpoint: "192.168.1.10:8193",
				Status:   "connected",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")

	req := fanucService.ConnectionRequest{
		Endpoint: "192.168.1.10:8193",
		Timeout:  1000,
		Model:    "TestModel",
		Series:   "0i",
	}
	result, err := client.CreateConnection(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "uuid-123", result.ID)
	assert.Equal(t, "connected", result.Status)
}

func TestClient_GetConnections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/connect", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		resp := apiResponse{
			Status: "ok",
			Data: []fanucService.MachineDTO{
				{ID: "1", Endpoint: "10.0.0.1:8193"},
				{ID: "2", Endpoint: "10.0.0.2:8193"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")
	machines, err := client.GetConnections(context.Background())

	require.NoError(t, err)
	assert.Len(t, machines, 2)
	assert.Equal(t, "1", machines[0].ID)
	assert.Equal(t, "2", machines[1].ID)
}

func TestClient_CheckConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/connect", r.URL.Path)
		assert.Equal(t, "id=uuid-123", r.URL.RawQuery)
		assert.Equal(t, http.MethodGet, r.Method)

		resp := apiResponse{
			Status: "ok",
			Data: fanucService.MachineDTO{
				ID:     "uuid-123",
				Status: "reconnecting",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")
	machine, err := client.CheckConnection(context.Background(), "uuid-123")

	require.NoError(t, err)
	assert.Equal(t, "reconnecting", machine.Status)
}

func TestClient_DeleteConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/connect", r.URL.Path)
		assert.Equal(t, "id=uuid-123", r.URL.RawQuery)
		assert.Equal(t, http.MethodDelete, r.Method)

		resp := apiResponse{
			Status:  "ok",
			Message: "deleted",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")
	err := client.DeleteConnection(context.Background(), "uuid-123")

	require.NoError(t, err)
}

func TestClient_StartPolling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/polling/start", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var reqBody fanucService.StartPollingRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "uuid-123", reqBody.ID)
		assert.Equal(t, 500, reqBody.Interval)

		resp := apiResponse{Status: "ok", Message: "polling started"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")
	err := client.StartPolling(context.Background(), "uuid-123", 500)

	require.NoError(t, err)
}

func TestClient_StopPolling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/polling/stop", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var reqBody fanucService.StopPollingRequest
		json.NewDecoder(r.Body).Decode(&reqBody)
		assert.Equal(t, "uuid-123", reqBody.ID)

		resp := apiResponse{Status: "ok", Message: "polling stopped"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")
	err := client.StopPolling(context.Background(), "uuid-123")

	require.NoError(t, err)
}

func TestClient_GetControlProgram(t *testing.T) {
	expectedProgram := "%\nO1000\nN10 G90 G00 X0 Y0\nM30\n%"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/program", r.URL.Path)
		assert.Equal(t, "id=uuid-123", r.URL.RawQuery)
		assert.Equal(t, http.MethodGet, r.Method)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedProgram))
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "test-api-key")
	program, err := client.GetControlProgram(context.Background(), "uuid-123")

	require.NoError(t, err)
	assert.Equal(t, expectedProgram, program)
}

func TestClient_AuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		resp := apiResponse{
			Status:  "error",
			Message: "unauthorized",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "wrong-key")
	_, err := client.GetConnections(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "api error (401): unauthorized")
}

func TestClient_GetControlProgram_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		resp := apiResponse{
			Status:  "error",
			Message: "machine not reachable",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "key")
	prog, err := client.GetControlProgram(context.Background(), "uuid-123")

	require.Error(t, err)
	assert.Empty(t, prog)
	assert.Contains(t, err.Error(), "api error: machine not reachable")
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "key")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetConnections(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestClient_MalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid-json"))
	}))
	defer server.Close()

	client := fanucService.NewClient(server.URL, "key")
	_, err := client.GetConnections(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}
