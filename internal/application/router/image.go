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
		protected.POST("/upload", r.imageHandler.UploadImage)
		protected.GET("", r.imageHandler.GetImages) // Query params: entity_type, entity_id
		protected.GET("/:id", r.imageHandler.GetImageByID)
		protected.PUT("/:id", r.imageHandler.UpdateImage)
		protected.DELETE("/:id", r.imageHandler.DeleteImage)
	}
}

// RegisterStaticRoutes sets up static file serving for uploaded images
func (r *Router) RegisterStaticRoutes(router *gin.Engine) {
	// Serve uploaded images
	router.Static("/uploads", "./uploads")
}
