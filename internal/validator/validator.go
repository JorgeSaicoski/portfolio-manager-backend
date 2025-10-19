package validator

import (
	"fmt"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidateStringLength validates that a string is within the specified length range
func ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(value)
	if length < min {
		if min == 1 {
			return ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%s is required", fieldName),
			}
		}
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be at least %d characters", fieldName, min),
		}
	}
	if max > 0 && length > max {
		return ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be less than %d characters", fieldName, max),
		}
	}
	return nil
}

// ValidateProject validates all project fields
func ValidateProject(project *models.Project) error {
	// Validate title
	if err := ValidateStringLength(project.Title, "Title", 1, 100); err != nil {
		return err
	}

	// Validate description
	if err := ValidateStringLength(project.Description, "Description", 1, 0); err != nil {
		return err
	}

	// Add more validations as needed
	// For example: link format, client name, etc.

	return nil
}

// ValidateCategory validates all category fields
func ValidateCategory(category *models.Category) error {
	// Validate title
	if err := ValidateStringLength(category.Title, "Title", 1, 100); err != nil {
		return err
	}

	// Description is optional (pointer), so only validate if present
	if category.Description != nil {
		if err := ValidateStringLength(*category.Description, "Description", 0, 500); err != nil {
			return err
		}
	}

	return nil
}

// ValidateSection validates all section fields
func ValidateSection(section *models.Section) error {
	// Validate title
	if err := ValidateStringLength(section.Title, "Title", 1, 100); err != nil {
		return err
	}

	// Validate type
	if err := ValidateStringLength(section.Type, "Type", 1, 50); err != nil {
		return err
	}

	// Description is optional (pointer), so only validate if present
	if section.Description != nil {
		if err := ValidateStringLength(*section.Description, "Description", 0, 500); err != nil {
			return err
		}
	}

	return nil
}

// ValidatePortfolio validates all portfolio fields
func ValidatePortfolio(portfolio *models.Portfolio) error {
	// Validate title
	if err := ValidateStringLength(portfolio.Title, "Title", 1, 100); err != nil {
		return err
	}

	// Description is optional (pointer), so only validate if present
	if portfolio.Description != nil {
		if err := ValidateStringLength(*portfolio.Description, "Description", 0, 500); err != nil {
			return err
		}
	}

	return nil
}
