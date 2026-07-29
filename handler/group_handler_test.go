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
)

func setupGroupRoutes(svc *mockGroupSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewGroupHandler(svc)
	r := gin.New()
	r.POST("/groups", h.CreateGroup)
	r.DELETE("/groups/:id", h.DeleteGroup)
	r.POST("/groups/:id/members", h.AddMember)
	r.DELETE("/groups/:id/members/:user_id", h.RemoveMember)
	r.GET("/groups/:id/members", h.GetMembers)
	return r
}

func TestGroupHandler_CreateGroup(t *testing.T) {
	t.Run("created successfully", func(t *testing.T) {
		svc := &mockGroupSvc{
			createGroupFn: func(name string) (*models.Group, error) {
				return &models.Group{ID: 1, Name: name}, nil
			},
		}
		r := setupGroupRoutes(svc)

		body := `{"name":"Test Group"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.Group
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Test Group", resp.Name)
	})

	t.Run("missing name", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockGroupSvc{
			createGroupFn: func(name string) (*models.Group, error) {
				return nil, assert.AnError
			},
		}
		r := setupGroupRoutes(svc)

		body := `{"name":"Test Group"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGroupHandler_DeleteGroup(t *testing.T) {
	t.Run("deleted successfully", func(t *testing.T) {
		var capturedReqID uint
		svc := &mockGroupSvc{
			deleteGroupFn: func(id, requesterID uint) error {
				capturedReqID = requesterID
				return nil
			},
		}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint(1), capturedReqID)
	})

	t.Run("invalid group id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/abc", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing requester_id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("forbidden error", func(t *testing.T) {
		svc := &mockGroupSvc{
			deleteGroupFn: func(id, requesterID uint) error {
				return assert.AnError
			},
		}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGroupHandler_AddMember(t *testing.T) {
	t.Run("owner adds member successfully", func(t *testing.T) {
		var capturedRole string
		svc := &mockGroupSvc{
			addMemberFn: func(gid, uid uint, role string, requesterID uint) error {
				capturedRole = role
				return nil
			},
		}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1,"user_id":10,"role":"admin"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups/1/members", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "admin", capturedRole)
	})

	t.Run("invalid group id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1,"user_id":10}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups/abc/members", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing user_id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups/1/members", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing requester_id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{"user_id":10}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/groups/1/members", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGroupHandler_RemoveMember(t *testing.T) {
	t.Run("removed successfully", func(t *testing.T) {
		var capturedReqID uint
		svc := &mockGroupSvc{
			removeMemberFn: func(gid, uid uint, requesterID uint) error {
				capturedReqID = requesterID
				return nil
			},
		}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1/members/10", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint(1), capturedReqID)
	})

	t.Run("invalid group id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/abc/members/10", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{"requester_id":1}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1/members/xyz", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing requester_id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		body := `{}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/groups/1/members/10", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGroupHandler_GetMembers(t *testing.T) {
	t.Run("returns members", func(t *testing.T) {
		expected := []models.GroupMember{
			{GroupID: 1, UserID: 10, Role: models.RoleMember},
			{GroupID: 1, UserID: 20, Role: models.RoleAdmin},
		}
		svc := &mockGroupSvc{
			getMembersFn: func(gid uint) ([]models.GroupMember, error) {
				return expected, nil
			},
		}
		r := setupGroupRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/groups/1/members", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.GroupMember
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 2)
	})

	t.Run("invalid group id", func(t *testing.T) {
		svc := &mockGroupSvc{}
		r := setupGroupRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/groups/abc/members", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockGroupSvc{
			getMembersFn: func(gid uint) ([]models.GroupMember, error) {
				return nil, assert.AnError
			},
		}
		r := setupGroupRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/groups/1/members", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
