package response

import (
	"time"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
)

// ImageResponse represents the response for an image
type ImageResponse struct {
	ID           uint      `json:"id"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	FileName     string    `json:"file_name"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	Alt          string    `json:"alt"`
	Type         string    `json:"type"`
	EntityID     uint      `json:"entity_id"`
	EntityType   string    `json:"entity_type"`
	IsMain       bool      `json:"is_main"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToImageResponse converts a models.Image to ImageResponse
func ToImageResponse(image *models.Image) ImageResponse {
	return ImageResponse{
		ID:           image.ID,
		URL:          image.URL,
		ThumbnailURL: image.ThumbnailURL,
		FileName:     image.FileName,
		FileSize:     image.FileSize,
		MimeType:     image.MimeType,
		Alt:          image.Alt,
		Type:         image.Type,
		EntityID:     image.EntityID,
		EntityType:   image.EntityType,
		IsMain:       image.IsMain,
		CreatedAt:    image.CreatedAt,
		UpdatedAt:    image.UpdatedAt,
	}
}

// ToImageResponses converts a slice of models.Image to ImageResponse slice
func ToImageResponses(images []models.Image) []ImageResponse {
	responses := make([]ImageResponse, len(images))
	for i, image := range images {
		responses[i] = ToImageResponse(&image) // pass pointer
	}
	return responses
}
