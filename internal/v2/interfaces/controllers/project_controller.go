package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/contracts"
	appdto "github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/dto"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/application/usecases/project"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/request"
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/v2/interfaces/dto/response"
	pkgerrors "github.com/JorgeSaicoski/portfolio-manager/backend/pkg/errors"
)

// ProjectController handles HTTP requests for project operations
type ProjectController struct {
	createUseCase    *project.CreateProjectUseCase
	getUseCase       *project.GetProjectUseCase
	getPublicUseCase *project.GetProjectPublicUseCase
	listUseCase      *project.ListProjectsUseCase
	updateUseCase    *project.UpdateProjectUseCase
	deleteUseCase    *project.DeleteProjectUseCase
	projectRepo      contracts.ProjectRepository
}

// NewProjectController creates a new project controller instance
func NewProjectController(
	createUC *project.CreateProjectUseCase,
	getUC *project.GetProjectUseCase,
	getPublicUC *project.GetProjectPublicUseCase,
	listUC *project.ListProjectsUseCase,
	updateUC *project.UpdateProjectUseCase,
	deleteUC *project.DeleteProjectUseCase,
	projectRepo contracts.ProjectRepository,
) *ProjectController {
	return &ProjectController{
		createUseCase:    createUC,
		getUseCase:       getUC,
		getPublicUseCase: getPublicUC,
		listUseCase:      listUC,
		updateUseCase:    updateUC,
		deleteUseCase:    deleteUC,
		projectRepo:      projectRepo,
	}
}

// Create handles POST /api/projects/own
func (ctrl *ProjectController) Create(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map HTTP request DTO to application DTO
	input := appdto.CreateProjectInput{
		Title:       req.Title,
		Description: req.Description,
		MainImage:   req.MainImage,
		Images:      req.Images,
		Skills:      req.Skills,
		Client:      req.Client,
		Link:        req.Link,
		CategoryID:  req.CategoryID,
		OwnerID:     userID,
	}

	// Execute use case
	projectDTO, err := ctrl.createUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map application DTO to HTTP response DTO
	resp := response.ProjectResponse{
		ID:          projectDTO.ID,
		Title:       projectDTO.Title,
		Description: projectDTO.Description,
		MainImage:   projectDTO.MainImage,
		Images:      projectDTO.Images,
		Skills:      projectDTO.Skills,
		Client:      projectDTO.Client,
		Link:        projectDTO.Link,
		CategoryID:  projectDTO.CategoryID,
		OwnerID:     projectDTO.OwnerID,
		CreatedAt:   projectDTO.CreatedAt,
		UpdatedAt:   projectDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusCreated, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// List handles GET /api/projects/own
func (ctrl *ProjectController) List(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Bind and validate query parameters
	var req request.ListProjectsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Set default pagination values if not provided
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Map to application DTO
	input := appdto.ListProjectsInput{
		OwnerID: userID,
		Pagination: appdto.PaginationDTO{
			Page:  req.Page,
			Limit: req.Limit,
		},
	}

	// Execute use case
	output, err := ctrl.listUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTOs
	projects := make([]response.ProjectResponse, len(output.Projects))
	for i, proj := range output.Projects {
		projects[i] = response.ProjectResponse{
			ID:          proj.ID,
			Title:       proj.Title,
			Description: proj.Description,
			MainImage:   proj.MainImage,
			Images:      proj.Images,
			Skills:      proj.Skills,
			Client:      proj.Client,
			Link:        proj.Link,
			CategoryID:  proj.CategoryID,
			OwnerID:     proj.OwnerID,
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.PaginatedDataResponse{
		Data:    projects,
		Page:    output.Pagination.Page,
		Limit:   output.Pagination.Limit,
		Total:   output.Pagination.Total,
		Message: "Success",
	})
}

// GetByID handles GET /api/projects/own/:id
func (ctrl *ProjectController) GetByID(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse project ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project ID"})
		return
	}

	// Execute use case
	projectDTO, err := ctrl.getUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO
	resp := response.ProjectResponse{
		ID:          projectDTO.ID,
		Title:       projectDTO.Title,
		Description: projectDTO.Description,
		MainImage:   projectDTO.MainImage,
		Images:      projectDTO.Images,
		Skills:      projectDTO.Skills,
		Client:      projectDTO.Client,
		Link:        projectDTO.Link,
		CategoryID:  projectDTO.CategoryID,
		OwnerID:     projectDTO.OwnerID,
		CreatedAt:   projectDTO.CreatedAt,
		UpdatedAt:   projectDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// Update handles PUT /api/projects/own/:id
func (ctrl *ProjectController) Update(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse project ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project ID"})
		return
	}

	// Bind and validate HTTP request DTO
	var req request.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to application DTO
	input := appdto.UpdateProjectInput{
		ID:          uint(id),
		Title:       req.Title,
		Description: req.Description,
		MainImage:   req.MainImage,
		Images:      req.Images,
		Skills:      req.Skills,
		Client:      req.Client,
		Link:        req.Link,
		OwnerID:     userID,
	}

	// Execute use case
	err = ctrl.updateUseCase.Execute(c.Request.Context(), input)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Project updated successfully",
	})
}

// Delete handles DELETE /api/projects/own/:id
func (ctrl *ProjectController) Delete(c *gin.Context) {
	// Extract userID from context (set by auth middleware)
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.ErrorResponse{Error: "unauthorized: missing user ID"})
		return
	}

	// Parse project ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project ID"})
		return
	}

	// Execute use case
	err = ctrl.deleteUseCase.Execute(c.Request.Context(), uint(id), userID)
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Return success response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    nil,
		Message: "Project deleted successfully",
	})
}

// GetPublicByID handles GET /api/projects/public/:id
func (ctrl *ProjectController) GetPublicByID(c *gin.Context) {
	// Parse project ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid project ID"})
		return
	}

	// Execute use case (no auth required for public access)
	projectDTO, err := ctrl.getPublicUseCase.Execute(c.Request.Context(), uint(id))
	if err != nil {
		status := pkgerrors.ToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Map to HTTP response DTO (don't include OwnerID in public response)
	resp := response.ProjectResponse{
		ID:          projectDTO.ID,
		Title:       projectDTO.Title,
		Description: projectDTO.Description,
		MainImage:   projectDTO.MainImage,
		Images:      projectDTO.Images,
		Skills:      projectDTO.Skills,
		Client:      projectDTO.Client,
		Link:        projectDTO.Link,
		CategoryID:  projectDTO.CategoryID,
		CreatedAt:   projectDTO.CreatedAt,
		UpdatedAt:   projectDTO.UpdatedAt,
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    resp,
		Message: "Success",
	})
}

// GetByCategory handles GET /api/projects/category/:categoryId
func (ctrl *ProjectController) GetByCategory(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("categoryId")
	categoryID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Get all projects for the category
	projects, err := ctrl.projectRepo.GetByCategoryID(c.Request.Context(), uint(categoryID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "failed to retrieve projects"})
		return
	}

	// Map to HTTP response DTOs
	projectResponses := make([]response.ProjectResponse, len(projects))
	for i, proj := range projects {
		projectResponses[i] = response.ProjectResponse{
			ID:          proj.ID,
			Title:       proj.Title,
			Description: proj.Description,
			MainImage:   proj.MainImage,
			Images:      proj.Images,
			Skills:      proj.Skills,
			Client:      proj.Client,
			Link:        proj.Link,
			CategoryID:  proj.CategoryID,
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    projectResponses,
		Message: "Success",
	})
}

// SearchBySkills handles GET /api/projects/search/skills?skills=React&skills=Node.js
func (ctrl *ProjectController) SearchBySkills(c *gin.Context) {
	// Bind and validate query parameters
	var req request.SearchProjectsBySkillsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Search projects by skills
	projects, err := ctrl.projectRepo.SearchBySkills(c.Request.Context(), req.Skills)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "failed to search projects"})
		return
	}

	// Map to HTTP response DTOs
	projectResponses := make([]response.ProjectResponse, len(projects))
	for i, proj := range projects {
		projectResponses[i] = response.ProjectResponse{
			ID:          proj.ID,
			Title:       proj.Title,
			Description: proj.Description,
			MainImage:   proj.MainImage,
			Images:      proj.Images,
			Skills:      proj.Skills,
			Client:      proj.Client,
			Link:        proj.Link,
			CategoryID:  proj.CategoryID,
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    projectResponses,
		Message: "Success",
	})
}

// SearchByClient handles GET /api/projects/search/client?client=ABC%20Company
func (ctrl *ProjectController) SearchByClient(c *gin.Context) {
	// Bind and validate query parameters
	var req request.SearchProjectsByClientRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{Error: err.Error()})
		return
	}

	// Search projects by client
	projects, err := ctrl.projectRepo.SearchByClient(c.Request.Context(), req.Client)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{Error: "failed to search projects"})
		return
	}

	// Map to HTTP response DTOs
	projectResponses := make([]response.ProjectResponse, len(projects))
	for i, proj := range projects {
		projectResponses[i] = response.ProjectResponse{
			ID:          proj.ID,
			Title:       proj.Title,
			Description: proj.Description,
			MainImage:   proj.MainImage,
			Images:      proj.Images,
			Skills:      proj.Skills,
			Client:      proj.Client,
			Link:        proj.Link,
			CategoryID:  proj.CategoryID,
			CreatedAt:   proj.CreatedAt,
			UpdatedAt:   proj.UpdatedAt,
		}
	}

	// Return HTTP response with API_OVERVIEW.md format
	c.JSON(http.StatusOK, response.DataResponse{
		Data:    projectResponses,
		Message: "Success",
	})
}
