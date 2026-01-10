package dto

import "time"

// UserDTO represents a user in the application layer
type UserDTO struct {
	ID         string
	Email      string
	Name       string
	ExternalID string // ID from auth provider (e.g., Authentik user ID)
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
