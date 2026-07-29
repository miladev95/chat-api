package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"chat/models"
	"chat/repository"
)

// FileService defines business logic for file management.
type FileService interface {
	UploadFile(file multipart.File, header *multipart.FileHeader) (*models.File, error)
}

type fileService struct {
	fileRepo    repository.FileRepository
	uploadDir   string
}

// NewFileService constructs a FileService implementation.
func NewFileService(fileRepo repository.FileRepository, uploadDir string) FileService {
	return &fileService{fileRepo: fileRepo, uploadDir: uploadDir}
}

// UploadFile saves an uploaded file to disk and creates a metadata record.
func (s *fileService) UploadFile(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate a unique filename
	ext := filepath.Ext(header.Filename)
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate filename: %w", err)
	}
	filename := hex.EncodeToString(randomBytes) + ext

	// Full path to save the file
	filePath := filepath.Join(s.uploadDir, filename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(filePath) // clean up partial write
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create metadata record
	fileRecord := &models.File{
		URL:  "/uploads/" + filename,
		Size: written,
		Type: header.Header.Get("Content-Type"),
	}

	if err := s.fileRepo.CreateFile(fileRecord); err != nil {
		os.Remove(filePath) // clean up on DB error
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return fileRecord, nil
}
