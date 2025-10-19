package handler

import (
	"net/http"
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	repo    repo.ProjectRepository
	metrics *metrics.Collector
}

func NewProjectHandler(repo repo.ProjectRepository, metrics *metrics.Collector) *ProjectHandler {
	return &ProjectHandler{
		repo:    repo,
		metrics: metrics,
	}
}

func (h *ProjectHandler) GetByCategory(c *gin.Context) {
	categoryID := c.Param("categoryId")

	projects, err := h.repo.GetByCategoryID(categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve projects",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"message":  "Success",
	})
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	projectID := c.Param("id")

	// Parse project ID
	id, err := strconv.Atoi(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Project not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
		"message": "Success",
	})
}

func (h *ProjectHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var newProject models.Project
	if err := c.ShouldBindJSON(&newProject); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the owner
	newProject.OwnerID = userID

	// Create a project
	if err := h.repo.Create(&newProject); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create project",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"project": &newProject,
		"message": "Project created successfully",
	})
}

func (h *ProjectHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	projectID := c.Param("id")

	// Parse project ID
	id, err := strconv.Atoi(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// Parse request body
	var updateData models.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Update project
	if err := h.repo.Update(&updateData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": &updateData,
		"message": "Project updated successfully",
	})
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	projectID := c.Param("id")

	id, err := strconv.Atoi(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid project ID",
		})
		return
	}

	// Get a project to check ownership
	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Project not found",
		})
		return
	}

	if project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Access denied",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete project",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

func (h *ProjectHandler) GetBySkills(c *gin.Context) {
	skills := c.QueryArray("skills")
	if len(skills) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one skill is required",
		})
		return
	}

	projects, err := h.repo.GetBySkills(skills)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve projects",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"message":  "Success",
	})
}

func (h *ProjectHandler) GetByClient(c *gin.Context) {
	client := c.Query("client")
	if client == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Client name is required",
		})
		return
	}

	projects, err := h.repo.GetByClient(client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve projects",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"message":  "Success",
	})
}
