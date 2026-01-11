package authentik

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// Client implements the AuthProvider contract for Authentik
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewAuthProvider creates a new Authentik auth provider with the provided configuration
// Returns the AuthProvider interface
func NewAuthProvider(config *dto.AuthProviderConfig) contracts.AuthProvider {
	return &Client{
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		httpClient: &http.Client{},
	}
}

// makeRequest is a helper to make HTTP requests to Authentik
func (c *Client) makeRequest(method, endpoint string, payload interface{}) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}
