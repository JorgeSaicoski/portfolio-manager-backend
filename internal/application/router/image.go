package router

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterImageRoutes(apiGroup *gin.RouterGroup) {
	images := apiGroup.Group("/images")

	// Public routes
	{
		// Get images by entity (public)
		images.GET("/entity/:type/:id", r.imageHandler.GetImagesByEntity)
	}

	// Protected routes - require authentication
	protected := images.Group("/own")
	protected.Use(middleware.AuthMiddleware())
	{
		// Upload image with validation
		protected.POST("",
			middleware.ValidateImageUpload(),
			middleware.ValidateEntityOwnership(r.imageRepo),
			r.imageHandler.UploadImage)

		// Get user's own images
		protected.GET("", r.imageHandler.GetImages)

		// Update and delete require ownership validation
		protected.PUT("/:id",
			middleware.ValidateImageOwnership(r.imageRepo),
			r.imageHandler.UpdateImage)
		protected.DELETE("/:id",
			middleware.ValidateImageOwnership(r.imageRepo),
			r.imageHandler.DeleteImage)
	}
}

// RegisterStaticRoutes sets up static file serving for uploaded images
func (r *Router) RegisterStaticRoutes(router *gin.Engine) {
	// Use a custom handler with absolute paths and MIME type headers for better compatibility

	// Compute absolute upload directory path once at initialization
	uploadDir, err := filepath.Abs("./uploads")
	if err != nil {
		log.Printf("[uploads] ERROR: Failed to resolve upload directory: %v", err)
		uploadDir = "./uploads" // fallback
	} else {
		log.Printf("[uploads] Serving files from: %s", uploadDir)
	}

	router.GET("/uploads/*filepath", func(c *gin.Context) {
		reqPath := c.Param("filepath") // e.g. /images/thumbnail/foo.png
		relPath := strings.TrimPrefix(reqPath, "/")
		fsPath := filepath.Join(uploadDir, relPath)

		// Stat the file to log existence/permissions
		if info, err := os.Stat(fsPath); err != nil {
			if os.IsNotExist(err) {
				log.Printf("[uploads] NOT FOUND: request=%s -> fs=%s", c.Request.URL.Path, fsPath)
				c.Status(http.StatusNotFound)
				return
			}
			// Other stat error
			log.Printf("[uploads] STAT ERROR: request=%s -> fs=%s err=%v", c.Request.URL.Path, fsPath, err)
			c.Status(http.StatusInternalServerError)
			return
		} else {
			// File exists — log size and serve
			log.Printf("[uploads] SERVE: request=%s -> fs=%s size=%d", c.Request.URL.Path, fsPath, info.Size())
		}

		// Handle HEAD requests: return headers only
		if c.Request.Method == http.MethodHead {
			c.Status(http.StatusOK)
			return
		}

		// Set proper MIME type based on file extension
		ext := strings.ToLower(filepath.Ext(fsPath))
		switch ext {
		case ".png":
			c.Header("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			c.Header("Content-Type", "image/jpeg")
		case ".webp":
			c.Header("Content-Type", "image/webp")
		case ".gif":
			c.Header("Content-Type", "image/gif")
		}

		c.File(fsPath)
	})

	// Also register HEAD for the same path
	router.HEAD("/uploads/*filepath", func(c *gin.Context) {
		reqPath := c.Param("filepath")
		relPath := strings.TrimPrefix(reqPath, "/")
		fsPath := filepath.Join(uploadDir, relPath)

		if _, err := os.Stat(fsPath); err != nil {
			if os.IsNotExist(err) {
				log.Printf("[uploads] NOT FOUND (HEAD): request=%s -> fs=%s", c.Request.URL.Path, fsPath)
				c.Status(http.StatusNotFound)
				return
			}
			log.Printf("[uploads] STAT ERROR (HEAD): request=%s -> fs=%s err=%v", c.Request.URL.Path, fsPath, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	})
}
