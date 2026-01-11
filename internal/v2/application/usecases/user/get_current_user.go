package user

import (
	"context"
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// GetCurrentUserUseCase handles the business logic for getting the current user's profile
type GetCurrentUserUseCase struct {
	userRepo contracts.UserRepository
}

// NewGetCurrentUserUseCase creates a new instance of GetCurrentUserUseCase
func NewGetCurrentUserUseCase(userRepo contracts.UserRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves the current user's profile by ID
func (uc *GetCurrentUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserDTO, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}
