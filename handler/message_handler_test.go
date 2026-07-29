package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"chat/models"
	"chat/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func uintPtr(v uint) *uint { return &v }

func setupMessageRoutes(svc *mockMsgSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewMessageHandler(svc)
	r := gin.New()
	r.POST("/messages", h.SendMessage)
	r.GET("/messages", h.GetConversation)
	r.POST("/messages/:id/seen", h.MarkSeen)
	r.DELETE("/messages/:id", h.DeleteMessage)
	r.GET("/messages/unseen/:user_id", h.GetUnseenCount)
	return r
}

func TestMessageHandler_SendMessage(t *testing.T) {
	t.Run("direct message", func(t *testing.T) {
		svc := &mockMsgSvc{
			sendMessageFn: func(msg *models.Message) error {
				msg.ID = 1
				return nil
			},
		}
		r := setupMessageRoutes(svc)

		body := `{"sender_id":1,"receiver_id":2,"type":"text","content":"Hello!"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp models.Message
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, uint(1), resp.ID)
	})

	t.Run("missing sender_id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc)

		body := `{"receiver_id":2,"type":"text"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing type", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc)

		body := `{"sender_id":1,"receiver_id":2,"content":"Hello!"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service validation error", func(t *testing.T) {
		svc := &mockMsgSvc{
			sendMessageFn: func(msg *models.Message) error {
				return service.ErrInvalidConversation
			},
		}
		r := setupMessageRoutes(svc)

		body := `{"sender_id":1,"type":"text","content":"Hi"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestMessageHandler_GetConversation(t *testing.T) {
	t.Run("returns messages", func(t *testing.T) {
		svc := &mockMsgSvc{
			getConversationFn: func(rid, gid *uint, limit, offset int) ([]models.Message, error) {
				return []models.Message{
					{ID: 1, SenderID: 1, ReceiverID: uintPtr(2)},
				}, nil
			},
		}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages?receiver_id=2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []models.Message
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Len(t, resp, 1)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockMsgSvc{
			getConversationFn: func(rid, gid *uint, limit, offset int) ([]models.Message, error) {
				return nil, assert.AnError
			},
		}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages?receiver_id=2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMessageHandler_MarkSeen(t *testing.T) {
	t.Run("marked successfully", func(t *testing.T) {
		svc := &mockMsgSvc{
			markSeenFn: func(mid, uid uint) error {
				return nil
			},
		}
		r := setupMessageRoutes(svc)

		body := `{"user_id":2}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages/1/seen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid message id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc)

		body := `{"user_id":2}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages/abc/seen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing user_id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc)

		body := `{}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages/1/seen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestMessageHandler_DeleteMessage(t *testing.T) {
	t.Run("deleted successfully", func(t *testing.T) {
		svc := &mockMsgSvc{
			deleteMessageFn: func(mid uint) error {
				return nil
			},
		}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/messages/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid message id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/messages/abc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockMsgSvc{
			deleteMessageFn: func(mid uint) error {
				return assert.AnError
			},
		}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/messages/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMessageHandler_GetUnseenCount(t *testing.T) {
	t.Run("returns counts", func(t *testing.T) {
		svc := &mockMsgSvc{
			getUnseenCountFn: func(uid uint) (map[string]interface{}, error) {
				return map[string]interface{}{
					"direct_messages": []interface{}{},
					"group_messages":  []interface{}{},
					"total":           float64(3),
				}, nil
			},
		}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages/unseen/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages/unseen/abc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &mockMsgSvc{
			getUnseenCountFn: func(uid uint) (map[string]interface{}, error) {
				return nil, assert.AnError
			},
		}
		r := setupMessageRoutes(svc)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages/unseen/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
