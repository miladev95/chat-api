package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupFileRoutes(svc *mockFileSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewFileHandler(svc)
	r := gin.New()
	r.POST("/upload", h.Upload)
	return r
}

func TestFileHandler_Upload(t *testing.T) {
	t.Run("upload successfully", func(t *testing.T) {
		svc := &mockFileSvc{
			uploadFileFn: func(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
				return &models.File{
					ID:   1,
					URL:  "/uploads/test.png",
					Size: header.Size,
					Type: "image/png",
				}, nil
			},
		}
		r := setupFileRoutes(svc)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, _ := w.CreateFormFile("file", "test.png")
		fw.Write([]byte("fake-image-data"))
		w.Close()

		req := httptest.NewRequest("POST", "/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var result models.File
		assert.NoError(t, decodeJSON(resp.Body, &result))
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, "/uploads/test.png", result.URL)
		assert.Equal(t, "image/png", result.Type)
	})

	t.Run("missing file", func(t *testing.T) {
		svc := &mockFileSvc{}
		r := setupFileRoutes(svc)

		req := httptest.NewRequest("POST", "/upload", nil)
		req.Header.Set("Content-Type", "multipart/form-data; boundary=test")
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockFileSvc{
			uploadFileFn: func(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
				return nil, assert.AnError
			},
		}
		r := setupFileRoutes(svc)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, _ := w.CreateFormFile("file", "test.png")
		fw.Write([]byte("data"))
		w.Close()

		req := httptest.NewRequest("POST", "/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

// decodeJSON unmarshals JSON response body into the given value.
func decodeJSON(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
