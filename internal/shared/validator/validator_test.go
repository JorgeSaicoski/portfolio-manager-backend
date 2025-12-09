package validator

import (
	"testing"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/stretchr/testify/assert"
)

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		min       int
		max       int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Valid string within range",
			value:     "hello",
			fieldName: "TestField",
			min:       1,
			max:       10,
			wantErr:   false,
		},
		{
			name:      "Empty string with min 1 (required field)",
			value:     "",
			fieldName: "TestField",
			min:       1,
			max:       10,
			wantErr:   true,
			errMsg:    "TestField is required",
		},
		{
			name:      "String too short",
			value:     "hi",
			fieldName: "TestField",
			min:       5,
			max:       10,
			wantErr:   true,
			errMsg:    "TestField must be at least 5 characters",
		},
		{
			name:      "String too long",
			value:     "this is a very long string",
			fieldName: "TestField",
			min:       1,
			max:       10,
			wantErr:   true,
			errMsg:    "TestField must be less than 10 characters",
		},
		{
			name:      "No max limit (max = 0)",
			value:     "this is a very long string that should be valid",
			fieldName: "TestField",
			min:       1,
			max:       0,
			wantErr:   false,
		},
		{
			name:      "Exact minimum length",
			value:     "12345",
			fieldName: "TestField",
			min:       5,
			max:       10,
			wantErr:   false,
		},
		{
			name:      "Exact maximum length",
			value:     "1234567890",
			fieldName: "TestField",
			min:       5,
			max:       10,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStringLength(tt.value, tt.fieldName, tt.min, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateProject(t *testing.T) {
	tests := []struct {
		name    string
		project *models.Project
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid project",
			project: &models.Project{
				Title:       "Test Project",
				Description: "A test project description",
				CategoryID:  1,
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			project: &models.Project{
				Title:       "",
				Description: "A test project description",
				CategoryID:  1,
			},
			wantErr: true,
			errMsg:  "Title is required",
		},
		{
			name: "Title too long",
			project: &models.Project{
				Title:       "This is a very long title that exceeds the maximum allowed length of 100 characters for a project title field",
				Description: "A test project description",
				CategoryID:  1,
			},
			wantErr: true,
			errMsg:  "Title must be less than 100 characters",
		},
		{
			name: "Missing description",
			project: &models.Project{
				Title:       "Test Project",
				Description: "",
				CategoryID:  1,
			},
			wantErr: true,
			errMsg:  "Description is required",
		},
		{
			name: "Missing category ID",
			project: &models.Project{
				Title:       "Test Project",
				Description: "A test project description",
				CategoryID:  0,
			},
			wantErr: true,
			errMsg:  "Category ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProject(tt.project)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCategory(t *testing.T) {
	validDescription := "A valid description"
	longDescription := "This is a very long description that exceeds the maximum allowed length of 500 characters. " +
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
		"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. " +
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. " +
		"Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum. " +
		"Additional text to make this exceed 500 characters for testing purposes."

	tests := []struct {
		name     string
		category *models.Category
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid category",
			category: &models.Category{
				Title:       "Test Category",
				PortfolioID: 1,
			},
			wantErr: false,
		},
		{
			name: "Valid category with description",
			category: &models.Category{
				Title:       "Test Category",
				Description: &validDescription,
				PortfolioID: 1,
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			category: &models.Category{
				Title:       "",
				PortfolioID: 1,
			},
			wantErr: true,
			errMsg:  "Title is required",
		},
		{
			name: "Title too long",
			category: &models.Category{
				Title:       "This is a very long title that exceeds the maximum allowed length of 100 characters for a category field",
				PortfolioID: 1,
			},
			wantErr: true,
			errMsg:  "Title must be less than 100 characters",
		},
		{
			name: "Missing portfolio ID",
			category: &models.Category{
				Title:       "Test Category",
				PortfolioID: 0,
			},
			wantErr: true,
			errMsg:  "Portfolio ID is required",
		},
		{
			name: "Description too long",
			category: &models.Category{
				Title:       "Test Category",
				Description: &longDescription,
				PortfolioID: 1,
			},
			wantErr: true,
			errMsg:  "Description must be less than 500 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategory(tt.category)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSection(t *testing.T) {
	validDescription := "A valid section description"

	tests := []struct {
		name    string
		section *models.Section
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid section",
			section: &models.Section{
				Title:       "Test Section",
				Type:        "text",
				PortfolioID: 1,
			},
			wantErr: false,
		},
		{
			name: "Valid section with description",
			section: &models.Section{
				Title:       "Test Section",
				Type:        "text",
				Description: &validDescription,
				PortfolioID: 1,
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			section: &models.Section{
				Title:       "",
				Type:        "text",
				PortfolioID: 1,
			},
			wantErr: true,
			errMsg:  "Title is required",
		},
		{
			name: "Missing type",
			section: &models.Section{
				Title:       "Test Section",
				Type:        "",
				PortfolioID: 1,
			},
			wantErr: true,
			errMsg:  "Type is required",
		},
		{
			name: "Missing portfolio ID",
			section: &models.Section{
				Title:       "Test Section",
				Type:        "text",
				PortfolioID: 0,
			},
			wantErr: true,
			errMsg:  "Portfolio ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSection(tt.section)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePortfolio(t *testing.T) {
	validDescription := "A valid portfolio description"

	tests := []struct {
		name      string
		portfolio *models.Portfolio
		wantErr   bool
		errMsg    string
	}{
		{
			name: "Valid portfolio",
			portfolio: &models.Portfolio{
				Title: "Test Portfolio",
			},
			wantErr: false,
		},
		{
			name: "Valid portfolio with description",
			portfolio: &models.Portfolio{
				Title:       "Test Portfolio",
				Description: &validDescription,
			},
			wantErr: false,
		},
		{
			name: "Missing title",
			portfolio: &models.Portfolio{
				Title: "",
			},
			wantErr: true,
			errMsg:  "Title is required",
		},
		{
			name: "Title too long",
			portfolio: &models.Portfolio{
				Title: "This is a very long title that exceeds the maximum allowed length of 100 characters for a portfolio title",
			},
			wantErr: true,
			errMsg:  "Title must be less than 100 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePortfolio(tt.portfolio)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSectionContent(t *testing.T) {
	tests := []struct {
		name    string
		content *models.SectionContent
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid section content",
			content: &models.SectionContent{
				SectionID: 1,
				Type:      "text",
				Content:   "Test content",
			},
			wantErr: false,
		},
		{
			name: "Missing section ID",
			content: &models.SectionContent{
				SectionID: 0,
				Type:      "text",
				Content:   "Test content",
			},
			wantErr: true,
			errMsg:  "Section ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSectionContent(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Field:   "TestField",
		Message: "Test error message",
	}

	assert.Equal(t, "Test error message", err.Error())
}
