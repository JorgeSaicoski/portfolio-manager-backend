package authentik

import (
	"fmt"
	"io"
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// logoutRequest represents the internal request structure for Authentik logout API
type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Logout implements the AuthProvider contract - revokes the refresh token
func (c *Client) Logout(input dto.LogoutUserInput) (*dto.LogoutUserOutput, error) {
	payload := logoutRequest{
		RefreshToken: input.RefreshToken,
	}

	endpoint := fmt.Sprintf("%s/application/o/revoke/", c.baseURL)
	resp, err := c.makeRequest("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to make logout request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("logout failed with status %d: %s", resp.StatusCode, string(body))
	}

	return &dto.LogoutUserOutput{
		Success: true,
		Message: "User logged out successfully",
	}, nil
}
