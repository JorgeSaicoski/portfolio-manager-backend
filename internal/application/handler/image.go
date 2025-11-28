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
	"github.com/sirupsen/logrus"
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPLOAD_IMAGE_PARSE_FORM_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UploadImage",
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("Failed to parse form data")
		response.BadRequest(c, "Failed to parse form data")
		return
	}

	// Get the file from the request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPLOAD_IMAGE_NO_FILE_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UploadImage",
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("No file uploaded")
		response.BadRequest(c, "No file uploaded")
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "UPLOAD_IMAGE_FILE_CLOSE_ERROR",
				"where":     "backend/internal/application/handler/image.go",
				"function":  "UploadImage",
				"userID":    userID,
				"filename":  header.Filename,
				"error":     err.Error(),
			}).Error("Failed to close uploaded file")
		}
	}()

	// Parse form data
	var req dto.CreateImageRequest
	if err := c.ShouldBind(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_IMAGE_BAD_REQUEST",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UploadImage",
			"userID":    userID,
			"data":      req,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, err.Error())
		return
	}

	// Validate the image
	if err := ValidateImage(file, header); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPLOAD_IMAGE_VALIDATION_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UploadImage",
			"userID":    userID,
			"filename":  header.Filename,
			"error":     err.Error(),
		}).Warn("Invalid image file")
		response.BadRequest(c, err.Error())
		return
	}

	// Generate unique filename
	filename := GenerateUniqueFilename(header.Filename)

	// Save the image (optimized original + thumbnail)
	originalURL, thumbnailURL, err := SaveImage(file, filename)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPLOAD_IMAGE_SAVE_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UploadImage",
			"userID":    userID,
			"filename":  filename,
			"error":     err.Error(),
		}).Error("Failed to save image")
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
		if delErr := DeleteImageFiles(originalURL, thumbnailURL); delErr != nil {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation": "UPLOAD_IMAGE_FILE_DELETE_ERROR",
				"where":     "backend/internal/application/handler/image.go",
				"function":  "UploadImage",
				"userID":    userID,
				"filename":  filename,
				"error":     delErr.Error(),
			}).Error("Failed to delete image files after DB error")
		}
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPLOAD_IMAGE_DB_CREATE_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UploadImage",
			"userID":    userID,
			"filename":  filename,
			"error":     err.Error(),
		}).Error("Failed to create image record")
		response.InternalError(c, "Failed to create image record")
		return
	}

	// Audit log
	audit.GetCreateLogger().WithFields(logrus.Fields{
		"operation":  "CREATE_IMAGE",
		"userID":     userID,
		"imageID":    image.ID,
		"entityType": req.EntityType,
		"entityID":   req.EntityID,
		"filename":   header.Filename,
		"fileSize":   header.Size,
		"mimeType":   header.Header.Get("Content-Type"),
		"alt":        req.Alt,
		"isMain":     req.IsMain,
	}).Info("Image uploaded successfully")

	// Metrics
	h.metrics.IncImagesUploaded()

	response.Created(c, "image", dtoResponse.ToImageResponse(image), "Image uploaded successfully")
}

// GetImages retrieves images for a specific entity (query params)
func (h *ImageHandler) GetImages(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	// Validate query parameters
	if entityType == "" || entityIDStr == "" {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_IMAGES_MISSING_PARAMS",
			"where":       "backend/internal/application/handler/image.go",
			"function":    "GetImages",
			"entityType":  entityType,
			"entityIDStr": entityIDStr,
		}).Warn("Missing required query parameters")
		response.BadRequest(c, "entity_type and entity_id are required")
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_IMAGES_INVALID_ID",
			"where":       "backend/internal/application/handler/image.go",
			"function":    "GetImages",
			"entityIDStr": entityIDStr,
			"error":       err.Error(),
		}).Warn("Invalid entity_id")
		response.BadRequest(c, "Invalid entity_id")
		return
	}

	images, err := h.repo.GetByEntity(entityType, uint(entityID))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "GET_IMAGES_DB_ERROR",
			"where":      "backend/internal/application/handler/image.go",
			"function":   "GetImages",
			"entityType": entityType,
			"entityID":   entityID,
			"error":      err.Error(),
		}).Error("Failed to retrieve images from database")
		response.InternalError(c, "Failed to retrieve images")
		return
	}

	response.OK(c, "images", dtoResponse.ToImageResponses(images), "Success")
}

// GetImagesByEntity retrieves images for a specific entity (path params - public)
func (h *ImageHandler) GetImagesByEntity(c *gin.Context) {
	entityType := c.Param("type")
	entityIDStr := c.Param("id")

	// Validate path parameters
	if entityType == "" {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_IMAGES_BY_ENTITY_MISSING_TYPE",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "GetImagesByEntity",
		}).Warn("Missing entity_type parameter")
		response.BadRequest(c, "entity_type is required")
		return
	}

	entityID, err := strconv.Atoi(entityIDStr)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "GET_IMAGES_BY_ENTITY_INVALID_ID",
			"where":       "backend/internal/application/handler/image.go",
			"function":    "GetImagesByEntity",
			"entityIDStr": entityIDStr,
			"error":       err.Error(),
		}).Warn("Invalid entity_id")
		response.BadRequest(c, "Invalid entity_id")
		return
	}

	images, err := h.repo.GetByEntity(entityType, uint(entityID))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "GET_IMAGES_BY_ENTITY_DB_ERROR",
			"where":      "backend/internal/application/handler/image.go",
			"function":   "GetImagesByEntity",
			"entityType": entityType,
			"entityID":   entityID,
			"error":      err.Error(),
		}).Error("Failed to retrieve images from database")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "GET_IMAGE_BY_ID_INVALID_ID",
			"where":      "backend/internal/application/handler/image.go",
			"function":   "GetImageByID",
			"imageIDStr": imageIDStr,
			"error":      err.Error(),
		}).Warn("Invalid image ID")
		response.BadRequest(c, "Invalid image ID")
		return
	}

	image, err := h.repo.GetByID(uint(imageID))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_IMAGE_BY_ID_NOT_FOUND",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "GetImageByID",
			"imageID":   imageID,
			"error":     err.Error(),
		}).Warn("Image not found")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_IMAGE_INVALID_ID",
			"where":      "backend/internal/application/handler/image.go",
			"function":   "UpdateImage",
			"imageIDStr": imageIDStr,
			"error":      err.Error(),
		}).Warn("Invalid image ID")
		response.BadRequest(c, "Invalid image ID")
		return
	}

	// Check ownership
	// Get image to check ownership and get details
	image, err := h.repo.GetByID(uint(imageID))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_IMAGE_NOT_FOUND",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UpdateImage",
			"imageID":   imageID,
			"error":     err.Error(),
		}).Warn("Image not found")
		response.NotFound(c, "Image not found")
		return
	}

	if image.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_IMAGE_FORBIDDEN",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UpdateImage",
			"imageID":   imageID,
			"ownerID":   image.OwnerID,
			"userID":    userID,
		}).Warn("Permission denied to update image")
		response.ForbiddenWithDetails(c, "You don't have permission to update this image", map[string]interface{}{
			"resource_type": "image",
			"resource_id":   image.ID,
			"owner_id":      image.OwnerID,
			"entity_type":   image.EntityType,
			"entity_id":     image.EntityID,
			"action":        "update",
		})
		return
	}

	// Parse request body
	var req dto.UpdateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_IMAGE_BAD_REQUEST",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UpdateImage",
			"imageID":   imageID,
			"userID":    userID,
			"data":      req,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, err.Error())
		return
	}

	// Get existing image
	image, err = h.repo.GetByID(uint(imageID))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_IMAGE_NOT_FOUND_AGAIN",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UpdateImage",
			"imageID":   imageID,
			"error":     err.Error(),
		}).Warn("Image not found (second check)")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_IMAGE_DB_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "UpdateImage",
			"imageID":   imageID,
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to update image")
		response.InternalError(c, "Failed to update image")
		return
	}

	// Audit log
	if len(changes) > 0 {
		audit.GetUpdateLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_IMAGE",
			"userID":     userID,
			"imageID":    imageID,
			"entityType": image.EntityType,
			"entityID":   image.EntityID,
			"changes":    changes,
		}).Info("Image updated successfully")
	}

	response.OK(c, "image", dtoResponse.ToImageResponse(image), "Image updated successfully")
}

// DeleteImage deletes an image
func (h *ImageHandler) DeleteImage(c *gin.Context) {
	userID := c.GetString("userID")
	imageIDStr := c.Param("id")

	imageID, err := strconv.Atoi(imageIDStr)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "DELETE_IMAGE_INVALID_ID",
			"where":      "backend/internal/application/handler/image.go",
			"function":   "DeleteImage",
			"imageIDStr": imageIDStr,
			"userID":     userID,
			"error":      err.Error(),
		}).Warn("Invalid image ID")
		response.BadRequest(c, "Invalid image ID")
		return
	}

	// Check ownership
	// Get image details for ownership check, audit log and file cleanup
	image, err := h.repo.GetByID(uint(imageID))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_IMAGE_NOT_FOUND",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "DeleteImage",
			"imageID":   imageID,
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("Image not found")
		response.NotFound(c, "Image not found")
		return
	}

	if image.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_IMAGE_FORBIDDEN",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "DeleteImage",
			"imageID":   imageID,
			"ownerID":   image.OwnerID,
			"userID":    userID,
		}).Warn("Permission denied to delete image")
		response.ForbiddenWithDetails(c, "You don't have permission to delete this image", map[string]interface{}{
			"resource_type": "image",
			"resource_id":   image.ID,
			"owner_id":      image.OwnerID,
			"entity_type":   image.EntityType,
			"entity_id":     image.EntityID,
			"action":        "delete",
		})
		return
	}

	// Delete from database first
	if err := h.repo.Delete(uint(imageID)); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_IMAGE_DB_ERROR",
			"where":     "backend/internal/application/handler/image.go",
			"function":  "DeleteImage",
			"imageID":   imageID,
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to delete image")
		response.InternalError(c, "Failed to delete image")
		return
	}

	// Delete files from filesystem
	if err := DeleteImageFiles(image.URL, image.ThumbnailURL); err != nil {
		// Log error but don't fail the request since DB record is already deleted
		audit.GetDeleteLogger().WithFields(logrus.Fields{
			"operation":  "DELETE_IMAGE_FILES_ERROR",
			"userID":     userID,
			"imageID":    imageID,
			"filename":   image.FileName,
			"entityType": image.EntityType,
			"entityID":   image.EntityID,
			"error":      err.Error(),
		}).Error("Failed to delete image files from filesystem")
	}

	// Audit log
	audit.GetDeleteLogger().WithFields(logrus.Fields{
		"operation":  "DELETE_IMAGE",
		"userID":     userID,
		"imageID":    imageID,
		"filename":   image.FileName,
		"entityType": image.EntityType,
		"entityID":   image.EntityID,
		"fileSize":   image.FileSize,
		"mimeType":   image.MimeType,
	}).Info("Image deleted successfully")

	// Metrics
	h.metrics.IncImagesDeleted()

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}
