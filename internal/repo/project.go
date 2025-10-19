package repo

import (
	"github.com/JorgeSaicoski/portfolio-manager/backend/internal/models"
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
	return r.db.Create(project).Error
}

// GetByID For basic project info
func (r *projectRepository) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	err := r.db.Where("id = ?", id).
		First(&project).Error
	return &project, err
}

// GetByCategoryID For list views - projects in a category
func (r *projectRepository) GetByCategoryID(categoryID string) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Where("category_id = ?", categoryID).
		Find(&projects).Error
	return projects, err
}

func (r *projectRepository) Update(project *models.Project) error {
	return r.db.Model(project).Where("id = ?", project.ID).Updates(project).Error
}

func (r *projectRepository) Delete(id uint) error {
	return r.db.Delete(&models.Project{}, id).Error
}

func (r *projectRepository) List(limit, offset int) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Limit(limit).Offset(offset).
		Find(&projects).Error
	return projects, err
}

// GetBySkills Find projects by skills
func (r *projectRepository) GetBySkills(skills []string) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Where("skills && ?", skills).
		Find(&projects).Error
	return projects, err
}

// GetByClient Find projects by client name
func (r *projectRepository) GetByClient(client string) ([]*models.Project, error) {
	var projects []*models.Project
	err := r.db.Where("client = ?", client).
		Find(&projects).Error
	return projects, err
}
