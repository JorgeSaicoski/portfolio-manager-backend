package contracts

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// AuthProvider defines the contract for an authentication provider
// The application depends on this to authenticate users
// Infrastructure implements this (e.g., Authentik, Auth0, etc.)
type AuthProvider interface {
	// Login authenticates a user with credentials
	// Input: LoginUserInput (username, password)
	// Output: LoginUserOutput (user info + tokens)
	Login(input dto.LoginUserInput) (*dto.LoginUserOutput, error)

	// Logout revokes a refresh token
	// Input: LogoutUserInput (refresh token, user ID)
	// Output: LogoutUserOutput (success status)
	Logout(input dto.LogoutUserInput) (*dto.LogoutUserOutput, error)

	// ValidateToken validates an access token
	// Input: access token string
	// Output: user ID if valid
	ValidateToken(accessToken string) (userID string, err error)
}
