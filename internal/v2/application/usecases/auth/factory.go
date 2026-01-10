package auth

import (
	"os"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	authentikInfra "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/authentik"
)

// UseCases holds all auth-related use cases
// This is a container that groups related use cases together
type UseCases struct {
	Login  *LoginUserUseCase
	Logout *LogoutUserUseCase
}

// NewAuthUseCases creates all auth use cases with the same auth provider
// It reads configuration from environment and creates the provider once
func NewAuthUseCases(userRepository contracts.UserRepository) *UseCases {
	// Read config from environment once
	config := &dto.AuthProviderConfig{
		BaseURL:      os.Getenv("AUTHENTIK_URL"),
		APIKey:       os.Getenv("AUTHENTIK_API_KEY"),
		ClientID:     os.Getenv("AUTHENTIK_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTHENTIK_CLIENT_SECRET"),
	}

	// Fallback for BaseURL
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("AUTHENTIK_ISSUER")
	}

	// Create the auth provider once
	authProvider := authentikInfra.NewAuthProvider(config)

	// Create all use cases with the same provider
	return &UseCases{
		Login:  NewLoginUserUseCase(authProvider, userRepository),
		Logout: NewLogoutUserUseCase(authProvider),
	}
}
