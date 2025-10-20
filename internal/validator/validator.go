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

	// Validate category_id is provided
	if project.CategoryID == 0 {
		return ValidationError{
			Field:   "CategoryID",
			Message: "Category ID is required",
		}
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

	// Validate portfolio_id is provided
	if category.PortfolioID == 0 {
		return ValidationError{
			Field:   "PortfolioID",
			Message: "Portfolio ID is required",
		}
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

	// Validate portfolio_id is provided
	if section.PortfolioID == 0 {
		return ValidationError{
			Field:   "PortfolioID",
			Message: "Portfolio ID is required",
		}
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

// ValidateSectionContent validates all section content fields
func ValidateSectionContent(content *models.SectionContent) error {
	// Validate section_id is provided
	if content.SectionID == 0 {
		return ValidationError{
			Field:   "SectionID",
			Message: "Section ID is required",
		}
	}

	// Validate type
	if content.Type != "text" && content.Type != "image" {
		return ValidationError{
			Field:   "Type",
			Message: "Type must be either 'text' or 'image'",
		}
	}

	// Validate content
	if err := ValidateStringLength(content.Content, "Content", 1, 5000); err != nil {
		return err
	}

	// Metadata is optional (pointer), so only validate if present
	if content.Metadata != nil {
		// Basic validation - check if not too long
		if len(*content.Metadata) > 10000 {
			return ValidationError{
				Field:   "Metadata",
				Message: "Metadata must be less than 10000 characters",
			}
		}
	}

	return nil
}
