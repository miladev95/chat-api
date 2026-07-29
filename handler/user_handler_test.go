package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chat/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupUserRoutes(svc *mockUserSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewUserHandler(svc)
	r := gin.New()
	r.POST("/users", h.CreateUser)
	r.GET("/users", h.GetAllUsers)
	r.GET("/users/:id", h.GetUserByID)
	return r
}

func TestUserHandler_CreateUser(t *testing.T) {
	t.Run("created successfully", func(t *testing.T) {
		svc := &mockUserSvc{
			createUserFn: func(u *models.User) error {
				u.ID = 1
				return nil
			},
		}
		r := setupUserRoutes(svc)

		body := `{"username":"john","email":"john@example.com"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.User
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, uint(1), resp.ID)
		assert.Equal(t, "john", resp.Username)
		assert.Equal(t, "john@example.com", resp.Email)
	})

	t.Run("missing username", func(t *testing.T) {
		svc := &mockUserSvc{}
		r := setupUserRoutes(svc)

		body := `{"email":"john@example.com"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid email", func(t *testing.T) {
		svc := &mockUserSvc{}
		r := setupUserRoutes(svc)

		body := `{"username":"john","email":"not-an-email"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockUserSvc{
			createUserFn: func(u *models.User) error {
				return gorm.ErrDuplicatedKey
			},
		}
		r := setupUserRoutes(svc)

		body := `{"username":"john","email":"john@example.com"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	t.Run("returns users", func(t *testing.T) {
		expected := []models.User{
			{ID: 1, Username: "alice", Email: "alice@example.com"},
			{ID: 2, Username: "bob", Email: "bob@example.com"},
		}
		svc := &mockUserSvc{
			getAllUsersFn: func() ([]models.User, error) {
				return expected, nil
			},
		}
		r := setupUserRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.User
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 2)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockUserSvc{
			getAllUsersFn: func() ([]models.User, error) {
				return nil, gorm.ErrInvalidDB
			},
		}
		r := setupUserRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUserHandler_GetUserByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc := &mockUserSvc{
			getUserByIDFn: func(id uint) (*models.User, error) {
				return &models.User{ID: 1, Username: "jane", Email: "jane@example.com"}, nil
			},
		}
		r := setupUserRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.User
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, uint(1), resp.ID)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockUserSvc{
			getUserByIDFn: func(id uint) (*models.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		r := setupUserRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/999", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := &mockUserSvc{}
		r := setupUserRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/users/abc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
