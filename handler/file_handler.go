package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"chat/service"
)

// FileHandler exposes HTTP handlers for file upload operations.
type FileHandler struct {
	fileService service.FileService
}

// NewFileHandler wires a file service into a handler instance.
func NewFileHandler(fileService service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// POST /upload
func (h *FileHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	fileRecord, err := h.fileService.UploadFile(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fileRecord)
}
