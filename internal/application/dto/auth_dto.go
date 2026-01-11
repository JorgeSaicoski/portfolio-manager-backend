package dto

// ============================================================================
// Use Case DTOs (Application Layer)
// ============================================================================

// LoginUserInput is the input for the LoginUserUseCase
type LoginUserInput struct {
	Username string
	Password string
}

// LoginUserOutput is the output from the LoginUserUseCase
type LoginUserOutput struct {
	UserID        string
	Email         string
	Name          string
	Username      string
	EmailVerified bool
	AccessToken   string
	RefreshToken  string
	ExpiresIn     int
}

// LogoutUserInput is the input for the LogoutUserUseCase
type LogoutUserInput struct {
	RefreshToken string
	UserID       string
}

// LogoutUserOutput is the output from the LogoutUserUseCase
type LogoutUserOutput struct {
	Success bool
	Message string
}

// ============================================================================
// API DTOs (Interfaces Layer)
// ============================================================================

type AuthProviderConfig struct {
	BaseURL      string
	APIKey       string
	ClientID     string
	ClientSecret string
}
