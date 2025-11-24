package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	dto "github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto/request"
	dtoResponse "github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/dto/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	repo    repo.ImageRepository
	metrics *metrics.Collector
}

func NewImageHandler(repo repo.ImageRepository, metrics *metrics.Collector) *ImageHandler {
	return &ImageHandler{
		repo:    repo,
		metrics: metrics,
	}
}

// UploadImage handles image upload
func (h *ImageHandler) UploadImage(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(MaxFileSize); err != nil {
		response.BadRequest(c, "Failed to parse form data")
		return
	}

	// Get the file from the request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file uploaded")
		return
	}
	defer file.Close()

	// Parse form data
	var req dto.CreateImageRequest
	if err := c.ShouldBind(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Validate the image
	if err := ValidateImage(file, header); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Generate unique filename
	filename := GenerateUniqueFilename(header.Filename)

	// Save the image (optimized original + thumbnail)
	originalURL, thumbnailURL, err := SaveImage(file, filename)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("Failed to save image: %v", err))
		return
	}

	// Create image record in database
	image := &models.Image{
		URL:          originalURL,
		ThumbnailURL: thumbnailURL,
		FileName:     header.Filename,
		FileSize:     header.Size,
		MimeType:     header.Header.Get("Content-Type"),
		Alt:          req.Alt,
		OwnerID:      userID,
		Type:         req.Type,
		EntityID:     req.EntityID,
		EntityType:   req.EntityType,
		IsMain:       req.IsMain,
	}

	if err := h.repo.Create(image); err != nil {
		// Clean up files if database creation fails
		DeleteImageFiles(originalURL, thumbnailURL)
		response.InternalError(c, "Failed to create image record")
		return
	}

	// Audit log
	audit.Logger.Info("Image uploaded",
		"user_id", userID,
		"image_id", image.ID,
		"entity_type", req.EntityType,
		"entity_id", req.EntityID,
		"filename", header.Filename,
		"file_size", header.Size,
		"mime_type", header.Header.Get("Content-Type"))

	// Metrics
	h.metrics.IncImagesUploaded()

	response.Created(c, "image", dtoResponse.ToImageResponse(image), "Image uploaded successfully")
}

// GetImages retrieves images for a specific entity
func (h *ImageHandler) GetImages(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	// Validate query parameters
	if entityType == "" || entityIDStr == "" {
		response.BadRequest(c, "entity_type and entity_id are required")
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid entity_id")
		return
	}

	images, err := h.repo.GetByEntity(entityType, uint(entityID))
	if err != nil {
		response.InternalError(c, "Failed to retrieve images")
		return
	}

	response.OK(c, "images", dtoResponse.ToImageResponses(images), "Success")
}

// GetImageByID retrieves a single image by ID
func (h *ImageHandler) GetImageByID(c *gin.Context) {
	imageIDStr := c.Param("id")

	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid image ID")
		return
	}

	image, err := h.repo.GetByID(uint(imageID))
	if err != nil {
		response.NotFound(c, "Image not found")
		return
	}

	response.OK(c, "image", dtoResponse.ToImageResponse(image), "Success")
}

// UpdateImage updates image metadata (alt text, is_main)
func (h *ImageHandler) UpdateImage(c *gin.Context) {
	userID := c.GetString("userID")
	imageIDStr := c.Param("id")

	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid image ID")
		return
	}

	// Check ownership
	isOwner, err := h.repo.CheckOwnership(uint(imageID), userID)
	if err != nil {
		response.InternalError(c, "Failed to verify ownership")
		return
	}
	if !isOwner {
		response.Forbidden(c, "You don't have permission to update this image")
		return
	}

	// Parse request body
	var req dto.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get existing image
	image, err := h.repo.GetByID(uint(imageID))
	if err != nil {
		response.NotFound(c, "Image not found")
		return
	}

	// Track changes for audit log
	changes := make(map[string]interface{})

	// Update fields if provided
	if req.Alt != "" && req.Alt != image.Alt {
		changes["alt"] = map[string]string{"from": image.Alt, "to": req.Alt}
		image.Alt = req.Alt
	}

	if req.IsMain != nil && *req.IsMain != image.IsMain {
		changes["is_main"] = map[string]bool{"from": image.IsMain, "to": *req.IsMain}
		image.IsMain = *req.IsMain

		// If setting as main, unset other images for the same entity
		if *req.IsMain {
			// This could be improved by adding a method to unset other main images
			// For now, the frontend should handle this
		}
	}

	// Update in database
	if err := h.repo.Update(image); err != nil {
		response.InternalError(c, "Failed to update image")
		return
	}

	// Audit log
	if len(changes) > 0 {
		audit.Logger.Info("Image updated",
			"user_id", userID,
			"image_id", imageID,
			"changes", changes)
	}

	response.OK(c, "image", dtoResponse.ToImageResponse(image), "Image updated successfully")
}

// DeleteImage deletes an image
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	userID := c.GetString("userID")
	imageIDStr := c.Param("id")

	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid image ID")
		return
	}

	// Check ownership
	isOwner, err := h.repo.CheckOwnership(uint(imageID), userID)
	if err != nil {
		response.InternalError(c, "Failed to verify ownership")
		return
	}
	if !isOwner {
		response.Forbidden(c, "You don't have permission to delete this image")
		return
	}

	// Get image details for audit log and file cleanup
	image, err := h.repo.GetByID(uint(imageID))
	if err != nil {
		response.NotFound(c, "Image not found")
		return
	}

	// Delete from database first
	if err := h.repo.Delete(uint(imageID)); err != nil {
		response.InternalError(c, "Failed to delete image")
		return
	}

	// Delete files from filesystem
	if err := DeleteImageFiles(image.URL, image.ThumbnailURL); err != nil {
		// Log error but don't fail the request since DB record is already deleted
		audit.Logger.Error("Failed to delete image files",
			"user_id", userID,
			"image_id", imageID,
			"error", err)
	}

	// Audit log
	audit.Logger.Info("Image deleted",
		"user_id", userID,
		"image_id", imageID,
		"filename", image.FileName,
		"entity_type", image.EntityType,
		"entity_id", image.EntityID)

	// Metrics
	h.metrics.IncImagesDeleted()

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}
