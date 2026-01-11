package response

import "time"

// SectionContentResponse represents a section content in HTTP responses
type SectionContentResponse struct {
	ID        uint      `json:"id"`
	SectionID uint      `json:"section_id"`
	Type      string    `json:"type"`
	Content   *string   `json:"content,omitempty"`
	Order     uint      `json:"order"`
	ImageID   *uint     `json:"image_id,omitempty"`
	OwnerID   string    `json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListSectionContentsResponse represents the response for listing section contents
type ListSectionContentsResponse struct {
	Contents []SectionContentResponse `json:"contents"`
}
