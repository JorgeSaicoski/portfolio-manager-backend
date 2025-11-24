package router

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

func (r *Router) RegisterImageRoutes(apiGroup *gin.RouterGroup) {
	images := apiGroup.Group("/images")

	// Protected routes - require authentication
	protected := images.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Upload image with validation
		protected.POST("/upload",
			middleware.ValidateImageUpload(),
			middleware.ValidateEntityOwnership(r.imageRepo),
			r.imageHandler.UploadImage)

		// Get images - no additional validation needed
		protected.GET("", r.imageHandler.GetImages)
		protected.GET("/:id", r.imageHandler.GetImageByID)

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
	// Serve uploaded images
	router.Static("/uploads", "./uploads")
}
