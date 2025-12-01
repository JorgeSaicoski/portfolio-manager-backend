package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/application/models"
	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

func (r *projectRepository) Create(project *models.Project) error {
	if err := r.db.Create(project).Error; err != nil {
		return err
	}
	// Reload the record to pick up any database-side defaults or trigger modifications
	return r.db.Where("id = ?", project.ID).First(project).Error
}

// GetByID For basic project info
func (r *projectRepository) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Select("id, title, description, skills, client, link, position, owner_id, category_id, created_at, updated_at").
		Preload("Images").
		Where("id = ?", id).
		First(&project).Error
	return &project, err
}

// GetByOwnerIDBasic For list views - only basic project info for a specific owner
func (r *projectRepository) GetByOwnerIDBasic(ownerID string, limit, offset int) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	// Get total count
	if err := r.db.Model(&models.Project{}).
		Where("owner_id = ?", ownerID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.Select("id, title, description, skills, client, link, position, owner_id, category_id, created_at, updated_at").
		Preload("Images").
		Where("owner_id = ?", ownerID).
		Order("position ASC, created_at ASC").
		Limit(limit).Offset(offset).
		Find(&projects).Error

	return projects, total, err
}

// GetByCategoryID For list views - projects in a category
func (r *projectRepository) GetByCategoryID(categoryID string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Select("id, title, description, skills, client, link, position, owner_id, category_id, created_at, updated_at").
		Preload("Images").
		Where("category_id = ?", categoryID).
		Order("position ASC, created_at ASC").
		Find(&projects).Error
	return projects, err
}

func (r *projectRepository) Update(project *models.Project) error {
	return r.db.Model(project).Where("id = ?", project.ID).Updates(project).Error
}

// UpdatePosition updates only the position field of a project
func (r *projectRepository) UpdatePosition(id uint, position uint) error {
	return r.db.Model(&models.Project{}).Where("id = ?", id).Update("position", position).Error
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

func (r *projectRepository) List(limit, offset int) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Select("id, title, description, skills, client, link, position, owner_id, category_id, created_at, updated_at").
		Preload("Images").
		Limit(limit).Offset(offset).
		Find(&projects).Error
	return projects, err
}

// GetBySkills Find projects by skills
func (r *projectRepository) GetBySkills(skills []string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Select("id, title, description, skills, client, link, position, owner_id, category_id, created_at, updated_at").
		Preload("Images").
		Where("skills && ?", skills).
		Find(&projects).Error
	return projects, err
}

// GetByClient Find projects by client name
func (r *projectRepository) GetByClient(client string) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Select("id, title, description, skills, client, link, position, owner_id, category_id, created_at, updated_at").
		Preload("Images").
		Where("client = ?", client).
		Find(&projects).Error
	return projects, err
}

// CheckDuplicate checks if a project with the same title exists for the same category
// excluding the project with the given id (useful for updates)
func (r *projectRepository) CheckDuplicate(title string, categoryID uint, id uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Project{}).Where("title = ? AND category_id = ?", title, categoryID)

	// Exclude the current project when checking for duplicates (for updates)
	if id != 0 {
		query = query.Where("id != ?", id)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
