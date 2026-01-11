package user

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// UpdateCurrentUserUseCase handles the business logic for updating the current user's profile
type UpdateCurrentUserUseCase struct {
	userRepo    contracts.UserRepository
	auditLogger contracts.AuditLogger
}

// NewUpdateCurrentUserUseCase creates a new instance of UpdateCurrentUserUseCase
func NewUpdateCurrentUserUseCase(
	userRepo contracts.UserRepository,
	auditLogger contracts.AuditLogger,
) *UpdateCurrentUserUseCase {
	return &UpdateCurrentUserUseCase{
		userRepo:    userRepo,
		auditLogger: auditLogger,
	}
}

// UpdateCurrentUserInput represents the input for updating user profile
type UpdateCurrentUserInput struct {
	UserID string
	Name   string
}

// Execute updates the current user's profile (name only)
func (uc *UpdateCurrentUserUseCase) Execute(ctx context.Context, input UpdateCurrentUserInput) (*dto.UserDTO, error) {
	// Validate input
	if input.UserID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Verify user exists before update
	_, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update user (email is empty string, so repository should not update it)
	err = uc.userRepo.Update(ctx, input.UserID, "", input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user
	updatedUser, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	// Audit logging
	if uc.auditLogger != nil {
		uc.auditLogger.LogUpdate(ctx, "user", input.UserID, map[string]interface{}{
			"name": input.Name,
		})
	}

	return updatedUser, nil
}
