package section_content

import "time"

// SectionContent represents a section content entity in the domain layer
type SectionContent struct {
	ID        uint
	SectionID uint
	Type      string
	Content   *string
	Order     uint
	ImageID   *uint
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
