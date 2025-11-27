package middleware

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp"
)

const (
	MaxImageFileSize = 10 * 1024 * 1024 // 10MB
)

// ValidateImageUpload validates file size and image format for image uploads
func ValidateImageUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_UPLOAD_NO_FILE",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageUpload",
				"error":     err.Error(),
			}).Warn("No file provided")
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
			c.Abort()
			return
		}

		// Validate file size
		if file.Size > MaxImageFileSize {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_UPLOAD_FILE_TOO_LARGE",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageUpload",
				"fileSize":  file.Size,
			}).Warn("File size exceeds maximum allowed size of 10MB")
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File size exceeds maximum allowed size of 10MB",
			})
			c.Abort()
			return
		}

		// Validate image format by actually decoding the file
		src, err := file.Open()
		if err != nil {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_UPLOAD_FILE_OPEN_ERROR",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageUpload",
				"error":     err.Error(),
			}).Warn("Failed to read uploaded file")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read uploaded file"})
			c.Abort()
			return
		}
		defer func() {
			if cerr := src.Close(); cerr != nil {
				audit.GetBadRequestLogger().WithFields(map[string]interface{}{
					"operation": "VALIDATE_IMAGE_UPLOAD_FILE_CLOSE_ERROR",
					"where":     "backend/internal/shared/middleware/image_validator.go",
					"function":  "ValidateImageUpload",
					"error":     cerr.Error(),
				}).Error("Failed to close uploaded image file")
			}
		}()

		// Try to decode the image to ensure it's valid and detect format
		_, format, err := image.DecodeConfig(src)
		if err != nil {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_UPLOAD_DECODE_ERROR",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageUpload",
				"error":     err.Error(),
			}).Warn("Invalid image file")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
			c.Abort()
			return
		}

		// Validate format is allowed (jpeg, png, webp)
		allowedFormats := map[string]bool{
			"jpeg": true,
			"png":  true,
			"webp": true,
		}

		if !allowedFormats[format] {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":     "VALIDATE_IMAGE_UPLOAD_INVALID_FORMAT",
				"where":         "backend/internal/shared/middleware/image_validator.go",
				"function":      "ValidateImageUpload",
				"received_type": format,
			}).Warn("Invalid file type. Only JPEG, PNG, and WebP images are allowed")
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Invalid file type. Only JPEG, PNG, and WebP images are allowed",
				"received_type": format,
				"allowed_types": []string{"jpeg", "png", "webp"},
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
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":  "VALIDATE_IMAGE_OWNERSHIP_INVALID_ID",
				"where":      "backend/internal/shared/middleware/image_validator.go",
				"function":   "ValidateImageOwnership",
				"imageIDStr": imageIDStr,
				"error":      err.Error(),
			}).Warn("Invalid image ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image ID"})
			c.Abort()
			return
		}

		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("userID")
		if !exists {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_OWNERSHIP_NO_USER",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageOwnership",
				"imageID":   imageID,
			}).Warn("User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check ownership
		isOwner, err := imageRepo.CheckOwnership(uint(imageID), userID.(string))
		if err != nil {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_OWNERSHIP_DB_ERROR",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageOwnership",
				"imageID":   imageID,
				"userID":    userID,
				"error":     err.Error(),
			}).Error("Failed to verify image ownership")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify image ownership"})
			c.Abort()
			return
		}

		if !isOwner {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_IMAGE_OWNERSHIP_FORBIDDEN",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateImageOwnership",
				"imageID":   imageID,
				"userID":    userID,
			}).Warn("Permission denied to modify image")
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
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":   "VALIDATE_ENTITY_OWNERSHIP_MISSING_PARAMS",
				"where":       "backend/internal/shared/middleware/image_validator.go",
				"function":    "ValidateEntityOwnership",
				"entityType":  entityType,
				"entityIDStr": entityIDStr,
			}).Warn("entity_type and entity_id are required")
			c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type and entity_id are required"})
			c.Abort()
			return
		}

		entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
		if err != nil {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":   "VALIDATE_ENTITY_OWNERSHIP_INVALID_ID",
				"where":       "backend/internal/shared/middleware/image_validator.go",
				"function":    "ValidateEntityOwnership",
				"entityIDStr": entityIDStr,
				"error":       err.Error(),
			}).Warn("Invalid entity_id")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid entity_id"})
			c.Abort()
			return
		}

		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation": "VALIDATE_ENTITY_OWNERSHIP_NO_USER",
				"where":     "backend/internal/shared/middleware/image_validator.go",
				"function":  "ValidateEntityOwnership",
				"entityID":  entityID,
			}).Warn("User not authenticated")
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
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":  "VALIDATE_ENTITY_OWNERSHIP_INVALID_TYPE",
				"where":      "backend/internal/shared/middleware/image_validator.go",
				"function":   "ValidateEntityOwnership",
				"entityType": entityType,
			}).Warn("Invalid entity_type")
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
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":  "VALIDATE_ENTITY_OWNERSHIP_DB_ERROR",
				"where":      "backend/internal/shared/middleware/image_validator.go",
				"function":   "ValidateEntityOwnership",
				"entityID":   entityID,
				"entityType": entityType,
				"userID":     userID,
				"error":      err.Error(),
			}).Error("Failed to verify entity ownership")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify entity ownership"})
			c.Abort()
			return
		}

		if !isOwner {
			audit.GetBadRequestLogger().WithFields(map[string]interface{}{
				"operation":  "VALIDATE_ENTITY_OWNERSHIP_FORBIDDEN",
				"where":      "backend/internal/shared/middleware/image_validator.go",
				"function":   "ValidateEntityOwnership",
				"entityID":   entityID,
				"entityType": entityType,
				"userID":     userID,
			}).Warn("Permission denied to add images to entity")
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to add images to this entity"})
			c.Abort()
			return
		}

		c.Next()
	}
}
