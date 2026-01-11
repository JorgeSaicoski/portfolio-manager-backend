package authentik

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// loginRequest represents the internal request structure for Authentik login API
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginResponse represents the internal response structure from Authentik login API
type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// Login implements the AuthProvider contract - authenticates user and returns user info + tokens
func (c *Client) Login(input dto.LoginUserInput) (*dto.LoginUserOutput, error) {
	// 1. Call Authentik login API
	payload := loginRequest{
		Username: input.Username,
		Password: input.Password,
	}

	endpoint := fmt.Sprintf("%s/application/o/token/", c.baseURL)
	resp, err := c.makeRequest("POST", endpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to make login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode login response: %w", err)
	}

	// 2. Get user info using the access token
	userInfo, err := c.getUserInfo(loginResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// 3. Build and return the output DTO
	return &dto.LoginUserOutput{
		UserID:        fmt.Sprintf("%d", userInfo.ID),
		Email:         userInfo.Email,
		Name:          userInfo.Name,
		Username:      userInfo.Username,
		EmailVerified: true,
		AccessToken:   loginResp.AccessToken,
		RefreshToken:  loginResp.RefreshToken,
		ExpiresIn:     loginResp.ExpiresIn,
	}, nil
}
