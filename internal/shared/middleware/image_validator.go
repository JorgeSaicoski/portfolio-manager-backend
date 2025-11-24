package middleware

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/gin-gonic/gin"
)

const (
	MaxImageFileSize = 10 * 1024 * 1024 // 10MB
)

var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
}

// ValidateImageUpload validates file size and MIME type for image uploads
func ValidateImageUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
			c.Abort()
			return
		}

		// Validate file size
		if file.Size > MaxImageFileSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File size exceeds maximum allowed size of 10MB",
			})
			c.Abort()
			return
		}

		// Validate MIME type
		contentType := file.Header.Get("Content-Type")
		if !allowedImageTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Invalid file type. Only JPEG, PNG, and WebP images are allowed",
				"received_type": contentType,
				"allowed_types": []string{"image/jpeg", "image/png", "image/webp"},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateImageOwnership validates that the user owns the image they're trying to modify/delete
func ValidateImageOwnership(imageRepo repo.ImageRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get image ID from URL
		imageIDStr := c.Param("id")
		imageID, err := strconv.ParseUint(imageIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
			c.Abort()
			return
		}

		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check ownership
		isOwner, err := imageRepo.CheckOwnership(uint(imageID), userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify image ownership"})
			c.Abort()
			return
		}

		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this image"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateEntityOwnership validates that the user owns the entity (project/portfolio/section) they're uploading an image to
func ValidateEntityOwnership(imageRepo repo.ImageRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get entity details from form
		entityType := c.PostForm("entity_type")
		entityIDStr := c.PostForm("entity_id")

		if entityType == "" || entityIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type and entity_id are required"})
			c.Abort()
			return
		}

		entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
			c.Abort()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Validate entity type
		validEntityTypes := map[string]bool{
			"project":   true,
			"portfolio": true,
			"section":   true,
		}

		if !validEntityTypes[entityType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Invalid entity_type",
				"allowed_types": []string{"project", "portfolio", "section"},
			})
			c.Abort()
			return
		}

		// Check entity ownership using the repository method
		isOwner, err := imageRepo.CheckEntityOwnership(uint(entityID), entityType, userID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify entity ownership"})
			c.Abort()
			return
		}

		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add images to this entity"})
			c.Abort()
			return
		}

		c.Next()
	}
}
