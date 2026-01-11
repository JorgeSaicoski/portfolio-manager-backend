package portfolio

import (
	"fmt"
	"time"
)

// Portfolio represents the domain entity with business rules
// This is a pure domain object with NO dependencies on other layers
type Portfolio struct {
	id          uint
	title       string
	description string
	ownerID     string
	createdAt   time.Time
	updatedAt   time.Time
}

// NewPortfolio creates a new portfolio with validation
func NewPortfolio(id uint, title, description, ownerID string) (*Portfolio, error) {
	// Business rules validation
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}
	if len(title) > 255 {
		return nil, fmt.Errorf("title cannot exceed 255 characters")
	}
	if len(description) > 2000 {
		return nil, fmt.Errorf("description cannot exceed 2000 characters")
	}
	if ownerID == "" {
		return nil, fmt.Errorf("owner ID cannot be empty")
	}

	now := time.Now()
	return &Portfolio{
		id:          id,
		title:       title,
		description: description,
		ownerID:     ownerID,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Getters - Provide read-only access to private fields
func (p *Portfolio) ID() uint {
	return p.id
}

func (p *Portfolio) Title() string {
	return p.title
}

func (p *Portfolio) Description() string {
	return p.description
}

func (p *Portfolio) OwnerID() string {
	return p.ownerID
}

func (p *Portfolio) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Portfolio) UpdatedAt() time.Time {
	return p.updatedAt
}

// Business rule: Can update title only if different from current
func (p *Portfolio) CanUpdateTitle(newTitle string) error {
	if newTitle == "" {
		return fmt.Errorf("new title cannot be empty")
	}
	if len(newTitle) > 255 {
		return fmt.Errorf("title cannot exceed 255 characters")
	}
	if newTitle == p.title {
		return fmt.Errorf("new title is the same as current title")
	}
	return nil
}

// Business rule: Can update description
func (p *Portfolio) CanUpdateDescription(newDescription string) error {
	if len(newDescription) > 2000 {
		return fmt.Errorf("description cannot exceed 2000 characters")
	}
	return nil
}

// Update title - with business rule checking
func (p *Portfolio) UpdateTitle(newTitle string) error {
	if err := p.CanUpdateTitle(newTitle); err != nil {
		return err
	}
	p.title = newTitle
	p.updatedAt = time.Now()
	return nil
}

// Update description - with business rule checking
func (p *Portfolio) UpdateDescription(newDescription string) error {
	if err := p.CanUpdateDescription(newDescription); err != nil {
		return err
	}
	p.description = newDescription
	p.updatedAt = time.Now()
	return nil
}

// Belongs to user - business rule check
func (p *Portfolio) BelongsToUser(userID string) bool {
	return p.ownerID == userID
}
