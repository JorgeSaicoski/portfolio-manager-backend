package handler

import (
	"strconv"
	"strings"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/audit"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/metrics"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/infrastructure/repo"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/response"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/shared/validator"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProjectHandler struct {
	repo          repo.ProjectRepository
	categoryRepo  repo.CategoryRepository
	portfolioRepo repo.PortfolioRepository
	metrics       *metrics.Collector
}

func NewProjectHandler(repo repo.ProjectRepository, categoryRepo repo.CategoryRepository, portfolioRepo repo.PortfolioRepository, metrics *metrics.Collector) *ProjectHandler {
	return &ProjectHandler{
		repo:          repo,
		categoryRepo:  categoryRepo,
		portfolioRepo: portfolioRepo,
		metrics:       metrics,
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECTS_BY_USER_DB_ERROR",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByUser",
			"userID":    userID,
			"error":     err.Error(),
		}).Error("Failed to retrieve projects")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "GET_PROJECTS_BY_CATEGORY_DB_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "GetByCategory",
			"categoryID": categoryID,
			"error":      err.Error(),
		}).Error("Failed to retrieve projects")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECT_BY_ID_INVALID_ID",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByID",
			"projectID": projectID,
			"error":     err.Error(),
		}).Warn("Invalid project ID")
		response.BadRequest(c, "Invalid project ID")
		return
	}

	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECT_BY_ID_NOT_FOUND",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByID",
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Project not found")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "CREATE_PROJECT_BAD_REQUEST",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Create",
			"userID":    userID,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the owner
	newProject.OwnerID = userID

	// Validate project data first (includes categoryID check)
	if err := validator.ValidateProject(&newProject); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "CREATE_PROJECT_VALIDATION_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Create",
			"userID":     userID,
			"categoryID": newProject.CategoryID,
			"error":      err.Error(),
		}).Warn("Project validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Validate category exists and belongs to user's portfolio
	category, err := h.categoryRepo.GetByID(newProject.CategoryID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "CREATE_PROJECT_CATEGORY_NOT_FOUND",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Create",
			"userID":     userID,
			"categoryID": newProject.CategoryID,
			"error":      err.Error(),
		}).Warn("Category not found")
		response.NotFound(c, "Category not found")
		return
	}

	portfolio, err := h.portfolioRepo.GetByIDBasic(category.PortfolioID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_PROJECT_PORTFOLIO_NOT_FOUND",
			"where":       "backend/internal/application/handler/project.go",
			"function":    "Create",
			"userID":      userID,
			"categoryID":  newProject.CategoryID,
			"portfolioID": category.PortfolioID,
			"error":       err.Error(),
		}).Warn("Portfolio not found")
		response.NotFound(c, "Portfolio not found")
		return
	}

	if portfolio.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":   "CREATE_PROJECT_FORBIDDEN",
			"where":       "backend/internal/application/handler/project.go",
			"function":    "Create",
			"userID":      userID,
			"categoryID":  newProject.CategoryID,
			"portfolioID": category.PortfolioID,
			"ownerID":     portfolio.OwnerID,
		}).Warn("Access denied: category belongs to another user's portfolio")
		response.ForbiddenWithDetails(c, "Access denied: category belongs to another user's portfolio", map[string]interface{}{
			"resource_type": "category",
			"resource_id":   category.ID,
			"owner_id":      portfolio.OwnerID,
			"action":        "create_project",
		})
		return
	}

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(newProject.Title, newProject.CategoryID, 0)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "CREATE_PROJECT_DUPLICATE_CHECK_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Create",
			"userID":     userID,
			"categoryID": newProject.CategoryID,
			"title":      newProject.Title,
			"error":      err.Error(),
		}).Error("Failed to check for duplicate project")
		response.InternalError(c, "Failed to check for duplicate project")
		return
	}
	if isDuplicate {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "CREATE_PROJECT_DUPLICATE_TITLE",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Create",
			"userID":     userID,
			"categoryID": newProject.CategoryID,
			"title":      newProject.Title,
		}).Warn("Project with this title already exists in this category")
		response.BadRequest(c, "Project with this title already exists in this category")
		return
	}

	// Create a project
	if err := h.repo.Create(&newProject); err != nil {
		// Check if error is due to foreign key constraint (invalid category_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_categories_projects") || strings.Contains(errMsg, "23503") {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation":  "CREATE_PROJECT_FK_CONSTRAINT_ERROR",
				"where":      "backend/internal/application/handler/project.go",
				"function":   "Create",
				"userID":     userID,
				"categoryID": newProject.CategoryID,
				"error":      err.Error(),
			}).Warn("Category not found during project creation")
			response.NotFound(c, "Category not found")
			return
		}
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "CREATE_PROJECT_DB_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Create",
			"userID":     userID,
			"categoryID": newProject.CategoryID,
			"error":      err.Error(),
		}).Error("Failed to create project")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_INVALID_ID",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Update",
			"userID":    userID,
			"projectID": projectID,
			"error":     err.Error(),
		}).Warn("Invalid project ID")
		response.BadRequest(c, "Invalid project ID")
		return
	}

	// Parse request body
	var updateData models.Project
	if err := c.ShouldBindJSON(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_BAD_REQUEST",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Update",
			"userID":    userID,
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Set the ID and owner
	updateData.ID = uint(id)
	updateData.OwnerID = userID

	// Validate project data
	if err := validator.ValidateProject(&updateData); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_PROJECT_VALIDATION_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Update",
			"userID":     userID,
			"projectID":  id,
			"categoryID": updateData.CategoryID,
			"error":      err.Error(),
		}).Warn("Project validation failed")
		response.BadRequest(c, err.Error())
		return
	}

	// Check if project exists and belongs to user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_NOT_FOUND",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Update",
			"userID":    userID,
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Project not found")
		response.NotFound(c, "Project not found")
		return
	}
	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_FORBIDDEN",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Update",
			"userID":    userID,
			"projectID": id,
			"ownerID":   existing.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "project",
			"resource_id":   existing.ID,
			"owner_id":      existing.OwnerID,
			"action":        "update",
		})
		return
	}

	// Check for duplicate title
	isDuplicate, err := h.repo.CheckDuplicate(updateData.Title, updateData.CategoryID, updateData.ID)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_PROJECT_DUPLICATE_CHECK_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Update",
			"userID":     userID,
			"projectID":  id,
			"categoryID": updateData.CategoryID,
			"title":      updateData.Title,
			"error":      err.Error(),
		}).Error("Failed to check for duplicate project")
		response.InternalError(c, "Failed to check for duplicate project")
		return
	}
	if isDuplicate {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_PROJECT_DUPLICATE_TITLE",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Update",
			"userID":     userID,
			"projectID":  id,
			"categoryID": updateData.CategoryID,
			"title":      updateData.Title,
		}).Warn("Project with this title already exists in this category")
		response.BadRequest(c, "Project with this title already exists in this category")
		return
	}

	// Update project
	if err := h.repo.Update(&updateData); err != nil {
		// Check if error is due to foreign key constraint (invalid category_id)
		errMsg := err.Error()
		if strings.Contains(errMsg, "fk_categories_projects") || strings.Contains(errMsg, "23503") {
			audit.GetErrorLogger().WithFields(logrus.Fields{
				"operation":  "UPDATE_PROJECT_FK_CONSTRAINT_ERROR",
				"where":      "backend/internal/application/handler/project.go",
				"function":   "Update",
				"userID":     userID,
				"projectID":  id,
				"categoryID": updateData.CategoryID,
				"error":      err.Error(),
			}).Warn("Category not found")
			response.NotFound(c, "Category not found")
			return
		}
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "UPDATE_PROJECT_DB_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Update",
			"userID":     userID,
			"projectID":  id,
			"categoryID": updateData.CategoryID,
			"error":      err.Error(),
		}).Error("Failed to update project")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_PROJECT_INVALID_ID",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Delete",
			"userID":    userID,
			"projectID": projectID,
			"error":     err.Error(),
		}).Warn("Invalid project ID")
		response.BadRequest(c, "Invalid project ID")
		return
	}

	// Get a project to check ownership
	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_PROJECT_NOT_FOUND",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Delete",
			"userID":    userID,
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Project not found")
		response.NotFound(c, "Project not found")
		return
	}

	if project.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "DELETE_PROJECT_FORBIDDEN",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "Delete",
			"userID":    userID,
			"projectID": id,
			"ownerID":   project.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "project",
			"resource_id":   project.ID,
			"owner_id":      project.OwnerID,
			"action":        "delete",
		})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation":  "DELETE_PROJECT_DB_ERROR",
			"where":      "backend/internal/application/handler/project.go",
			"function":   "Delete",
			"projectID":  id,
			"userID":     userID,
			"categoryID": project.CategoryID,
			"error":      err.Error(),
		}).Error("Failed to delete project")

		response.InternalError(c, "Failed to delete project")
		return
	}

	audit.GetDeleteLogger().WithFields(logrus.Fields{
		"operation":  "DELETE_PROJECT",
		"projectID":  id,
		"userID":     userID,
		"categoryID": project.CategoryID,
		"title":      project.Title,
	}).Info("Project deleted successfully")

	response.OK(c, "message", "Project deleted successfully", "Success")
}

func (h *ProjectHandler) GetBySkills(c *gin.Context) {
	skills := c.QueryArray("skills")
	if len(skills) == 0 {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECTS_BY_SKILLS_MISSING_SKILLS",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetBySkills",
		}).Warn("At least one skill is required")
		response.BadRequest(c, "At least one skill is required")
		return
	}

	projects, err := h.repo.GetBySkills(skills)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECTS_BY_SKILLS_DB_ERROR",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetBySkills",
			"skills":    skills,
			"error":     err.Error(),
		}).Error("Failed to retrieve projects")
		response.InternalError(c, "Failed to retrieve projects")
		return
	}

	response.OK(c, "projects", projects, "Success")
}

func (h *ProjectHandler) GetByClient(c *gin.Context) {
	client := c.Query("client")
	if client == "" {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECTS_BY_CLIENT_MISSING_CLIENT",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByClient",
		}).Warn("Client name is required")
		response.BadRequest(c, "Client name is required")
		return
	}

	projects, err := h.repo.GetByClient(client)
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECTS_BY_CLIENT_DB_ERROR",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByClient",
			"client":    client,
			"error":     err.Error(),
		}).Error("Failed to retrieve projects")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECT_BY_ID_PUBLIC_INVALID_ID",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByIDPublic",
			"projectID": projectID,
			"error":     err.Error(),
		}).Warn("Invalid project ID")
		response.BadRequest(c, "Invalid project ID")
		return
	}

	project, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "GET_PROJECT_BY_ID_PUBLIC_NOT_FOUND",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "GetByIDPublic",
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Project not found")
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
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_POSITION_INVALID_ID",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"projectID": projectID,
			"error":     err.Error(),
		}).Warn("Invalid project ID")
		response.BadRequest(c, "Invalid project ID")
		return
	}

	// Parse request body
	var req struct {
		Position uint `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_POSITION_BAD_REQUEST",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Invalid request data")
		response.BadRequest(c, "Invalid request data")
		return
	}

	// Check if project exists and belongs to user
	existing, err := h.repo.GetByID(uint(id))
	if err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_POSITION_NOT_FOUND",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"projectID": id,
			"error":     err.Error(),
		}).Warn("Project not found")
		response.NotFound(c, "Project not found")
		return
	}

	if existing.OwnerID != userID {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_POSITION_FORBIDDEN",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"projectID": id,
			"ownerID":   existing.OwnerID,
		}).Warn("Access denied")
		response.ForbiddenWithDetails(c, "Access denied", map[string]interface{}{
			"resource_type": "project",
			"resource_id":   existing.ID,
			"owner_id":      existing.OwnerID,
			"action":        "update_position",
		})
		return
	}

	// Update position
	if err := h.repo.UpdatePosition(uint(id), req.Position); err != nil {
		audit.GetErrorLogger().WithFields(logrus.Fields{
			"operation": "UPDATE_PROJECT_POSITION_DB_ERROR",
			"where":     "backend/internal/application/handler/project.go",
			"function":  "UpdatePosition",
			"userID":    userID,
			"projectID": id,
			"position":  req.Position,
			"error":     err.Error(),
		}).Error("Failed to update project position")
		response.InternalError(c, "Failed to update project position")
		return
	}

	response.OK(c, "message", "Project position updated successfully", "Success")
}
