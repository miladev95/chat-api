package repository

import (
	"gorm.io/gorm"

	"chat/models"
)

// FileRepository defines data access operations for file metadata.
type FileRepository interface {
	CreateFile(file *models.File) error
	GetFileByID(fileID uint) (*models.File, error)
}

type fileRepo struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepo{db: db}
}

func (r *fileRepo) CreateFile(file *models.File) error {
	return r.db.Create(file).Error
}

func (r *fileRepo) GetFileByID(fileID uint) (*models.File, error) {
	var file models.File
	if err := r.db.First(&file, fileID).Error; err != nil {
		return nil, err
	}
	return &file, nil
}
