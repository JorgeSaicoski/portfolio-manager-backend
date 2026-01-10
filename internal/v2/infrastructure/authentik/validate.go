package authentik

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// userInfo represents the internal user info structure from Authentik API
type userInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// getUserInfo is a private helper method to retrieve user information from Authentik
// This is not part of the AuthProvider contract, just an internal helper
func (c *Client) getUserInfo(accessToken string) (*userInfo, error) {
	endpoint := fmt.Sprintf("%s/application/o/userinfo/", c.baseURL)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var info userInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}

	return &info, nil
}

// ValidateToken implements the AuthProvider contract - validates an access token
func (c *Client) ValidateToken(accessToken string) (string, error) {
	userInfo, err := c.getUserInfo(accessToken)
	if err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}
	return fmt.Sprintf("%d", userInfo.ID), nil
}
