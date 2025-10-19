package handler

import (
	"strconv"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/validator"
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
		response.InternalError(c, "Failed to retrieve projects")
		return
	}

	response.OK(c, "projects", projects, "Success")
}

func (h *ProjectHandler) GetByID(c *gin.Context) {
	projectID := c.Param("id")

	// Parse project ID
	id, err := strconv.Atoi(projectID)
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	response.OK(c, "project", project, "Success")
}

func (h *ProjectHandler) Create(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse request body
	var newProject models.Project
	if err := c.ShouldBindJSON(&newProject); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the owner
	newProject.OwnerID = userID

	// Validate project data
	if err := validator.ValidateProject(&newProject); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create a project
	if err := h.repo.Create(&newProject); err != nil {
		response.InternalError(c, "Failed to create project")
		return
	}

	response.Created(c, "project", &newProject, "Project created successfully")
}

func (h *ProjectHandler) Update(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	projectID := c.Param("id")

	// Parse project ID
	id, err := strconv.Atoi(projectID)
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	// Parse request body
	var updateData models.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Validate project data
	if err := validator.ValidateProject(&updateData); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update project
	if err := h.repo.Update(&updateData); err != nil {
		response.InternalError(c, "Failed to update project")
		return
	}

	response.OK(c, "project", &updateData, "Project updated successfully")
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	userID := c.GetString("userID")
	projectID := c.Param("id")

	id, err := strconv.Atoi(projectID)
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	// Get a project to check ownership
	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	if project.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		response.InternalError(c, "Failed to delete project")
		return
	}

	response.OK(c, "message", "Project deleted successfully", "Success")
}

func (h *ProjectHandler) GetBySkills(c *gin.Context) {
	skills := c.QueryArray("skills")
	if len(skills) == 0 {
		response.BadRequest(c, "At least one skill is required")
		return
	}

	projects, err := h.repo.GetBySkills(skills)
	if err != nil {
		response.InternalError(c, "Failed to retrieve projects")
		return
	}

	response.OK(c, "projects", projects, "Success")
}

func (h *ProjectHandler) GetByClient(c *gin.Context) {
	client := c.Query("client")
	if client == "" {
		response.BadRequest(c, "Client name is required")
		return
	}

	projects, err := h.repo.GetByClient(client)
	if err != nil {
		response.InternalError(c, "Failed to retrieve projects")
		return
	}

	response.OK(c, "projects", projects, "Success")
}
