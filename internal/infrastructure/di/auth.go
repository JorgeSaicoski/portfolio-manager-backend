package di

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	auth "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/auth"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/infrastructure/authentik"
)

// AuthServices holds all auth-related services and use cases
type AuthServices struct {
	InitializeAuthUseCase *auth.InitializeAuthUseCase
	LoginUserUseCase      *auth.LoginUserUseCase
	GetCurrentUserUseCase *auth.GetCurrentUserUseCase
	AuthService           contracts.AuthService
}

// NewAuthServices creates and initializes all auth services
func NewAuthServices() (*AuthServices, error) {
	// Create the infrastructure implementation
	authService := authentik.NewAuthentikService()

	// Create use cases with the service
	return &AuthServices{
		InitializeAuthUseCase: auth.NewInitializeAuthUseCase(authService),
		LoginUserUseCase:      auth.NewLoginUserUseCase(authService),
		GetCurrentUserUseCase: auth.NewGetCurrentUserUseCase(authService),
		AuthService:           authService,
	}, nil
}
