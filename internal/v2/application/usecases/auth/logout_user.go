package auth

import (
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// LogoutUserUseCase handles user logout
type LogoutUserUseCase struct {
	authProvider contracts.AuthProvider
}

// NewLogoutUserUseCase creates a new logout use case
func NewLogoutUserUseCase(authProvider contracts.AuthProvider) *LogoutUserUseCase {
	return &LogoutUserUseCase{
		authProvider: authProvider,
	}
}

// Execute handles the user logout process
func (uc *LogoutUserUseCase) Execute(input dto.LogoutUserInput) (*dto.LogoutUserOutput, error) {
	// 1. Validate input
	if input.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}
	if input.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// 2. Call auth provider to logout
	output, err := uc.authProvider.Logout(input)
	if err != nil {
		return nil, fmt.Errorf("logout failed: %w", err)
	}

	return output, nil
}
