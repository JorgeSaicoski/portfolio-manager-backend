package handler

import (
	"strconv"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/validator"
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

func (h *ProjectHandler) GetByUser(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware

	// Parse pagination parameters - using default values if not provided
	page := 1
	limit := 10
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	projects, err := h.repo.GetByOwnerIDBasic(userID, limit, offset)
	if err != nil {
		response.InternalError(c, "Failed to retrieve projects")
		return
	}

	response.SuccessWithPagination(c, 200, "projects", projects, page, limit)
}

func (h *ProjectHandler) GetByCategory(c *gin.Context) {
	// Try both parameter names for flexibility
	categoryID := c.Param("categoryId")
	if categoryID == "" {
		categoryID = c.Param("id")
	}

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

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(newProject.Title, newProject.CategoryID, 0)
	if err != nil {
		response.InternalError(c, "Failed to check for duplicate project")
		return
	}
	if isDuplicate {
		response.BadRequest(c, "Project with this title already exists in this category")
		return
	}

	// Create a project
	if err := h.repo.Create(&newProject); err != nil {
		// Check if error is due to foreign key constraint (invalid category_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_categories_projects") || strings.Contains(errMsg, "23503") {
			response.NotFound(c, "Category not found")
			return
		}
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

	// Check if project exists and belongs to user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Project not found")
		return
	}
	if existing.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(updateData.Title, updateData.CategoryID, updateData.ID)
	if err != nil {
		response.InternalError(c, "Failed to check for duplicate project")
		return
	}
	if isDuplicate {
		response.BadRequest(c, "Project with this title already exists in this category")
		return
	}

	// Update project
	if err := h.repo.Update(&updateData); err != nil {
		// Check if error is due to foreign key constraint (invalid category_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_categories_projects") || strings.Contains(errMsg, "23503") {
			response.NotFound(c, "Category not found")
			return
		}
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

func (h *ProjectHandler) GetByIDPublic(c *gin.Context) {
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

// UpdatePosition updates the position field of a project
func (h *ProjectHandler) UpdatePosition(c *gin.Context) {
	userID := c.GetString("userID") // From auth middleware
	projectID := c.Param("id")

	// Parse project ID
	id, err := strconv.Atoi(projectID)
	if err != nil {
		response.BadRequest(c, "Invalid project ID")
		return
	}

	// Parse request body
	var req struct {
		Position uint `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check if project exists and belongs to user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Project not found")
		return
	}

	if existing.OwnerID != userID {
		response.Forbidden(c, "Access denied")
		return
	}

	// Update position
	if err := h.repo.UpdatePosition(uint(id), req.Position); err != nil {
		response.InternalError(c, "Failed to update project position")
		return
	}

	response.OK(c, "message", "Project position updated successfully", "Success")
}
