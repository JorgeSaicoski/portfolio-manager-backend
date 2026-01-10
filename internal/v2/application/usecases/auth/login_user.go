package auth

import (
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
)

// LoginUserUseCase handles user login through an auth provider
type LoginUserUseCase struct {
	authProvider   contracts.AuthProvider
	userRepository contracts.UserRepository
}

// NewLoginUserUseCase creates a new login use case with explicit dependencies
func NewLoginUserUseCase(
	authProvider contracts.AuthProvider,
	userRepository contracts.UserRepository,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		authProvider:   authProvider,
		userRepository: userRepository,
	}
}

// Execute handles the user login process
func (uc *LoginUserUseCase) Execute(input dto.LoginUserInput) (*dto.LoginUserOutput, error) {
	// 1. Validate input
	if input.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if input.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// 2. Call auth provider to authenticate (returns everything in one call)
	output, err := uc.authProvider.Login(input)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// 3. Get or create user in database
	user, err := uc.userRepository.GetByExternalID(nil, output.UserID)
	if err != nil {
		// User doesn't exist, create them
		user, err = uc.userRepository.Create(nil, output.Email, output.Name, output.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists, update their info
		err = uc.userRepository.Update(nil, user.ID, output.Email, output.Name)
		if err != nil {
			// Log but continue - don't fail the login
			fmt.Printf("failed to update user: %v\n", err)
		}
	}

	// 4. Update the output with our internal user ID
	output.UserID = user.ID

	return output, nil
}
