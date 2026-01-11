package contracts

import (
	"context"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/dto"
)

// UserRepository defines the contract for user data access
// The application depends on this to persist and retrieve users
// Infrastructure implements this (e.g., PostgreSQL)
type UserRepository interface {
	// Create creates a new user
	// Input: user data (email, name, external ID from auth provider)
	// Output: created user DTO with ID
	Create(ctx context.Context, email, name, externalID string) (*dto.UserDTO, error)

	// GetByID retrieves a user by ID
	// Input: user ID
	// Output: user DTO
	GetByID(ctx context.Context, userID string) (*dto.UserDTO, error)

	// GetByExternalID retrieves a user by external ID (from auth provider)
	// Input: external ID (e.g., from Authentik)
	// Output: user DTO
	GetByExternalID(ctx context.Context, externalID string) (*dto.UserDTO, error)

	// Update updates user information
	// Input: user ID, updated data
	// Output: error if update fails
	Update(ctx context.Context, userID, email, name string) error

	// Delete deletes a user
	// Input: user ID
	// Output: error if deletion fails
	Delete(ctx context.Context, userID string) error
}
