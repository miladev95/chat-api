package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"chat/models"
	"chat/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func uintPtr(v uint) *uint { return &v }

func setupMessageRoutes(svc *mockMsgSvc, fileSvc *mockFileSvc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewMessageHandler(svc, fileSvc)
	r := gin.New()
	r.POST("/messages", h.SendMessage)
	r.POST("/messages/upload", h.SendFileMessage)
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
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

		body := `{"receiver_id":2,"type":"text"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing type", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
	r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

		body := `{"user_id":2}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages/1/seen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid message id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc, &mockFileSvc{})

		body := `{"user_id":2}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/messages/abc/seen", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing user_id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/messages/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid message id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages/unseen/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid user id", func(t *testing.T) {
		svc := &mockMsgSvc{}
		r := setupMessageRoutes(svc, &mockFileSvc{})

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
		r := setupMessageRoutes(svc, &mockFileSvc{})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/messages/unseen/1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func filePart(w *multipart.Writer, fieldname, filename, contentType, data string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldname, filename))
	h.Set("Content-Type", contentType)
	fw, err := w.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = fw.Write([]byte(data))
	return err
}

func TestMessageHandler_SendFileMessage(t *testing.T) {
	t.Run("sends image to direct conversation", func(t *testing.T) {
		fileSvc := &mockFileSvc{
			uploadFileFn: func(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
				return &models.File{ID: 10, URL: "/uploads/test.png", Size: 100, Type: "image/png"}, nil
			},
		}
		msgSvc := &mockMsgSvc{
			sendMessageFn: func(msg *models.Message) error {
				msg.ID = 1
				return nil
			},
		}
		r := setupMessageRoutes(msgSvc, fileSvc)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		require.NoError(t, filePart(w, "file", "photo.png", "image/png", "fake-image-data"))
		w.WriteField("sender_id", "1")
		w.WriteField("receiver_id", "2")
		w.WriteField("content", "Check out this photo!")
		w.Close()

		req := httptest.NewRequest("POST", "/messages/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var result models.Message
		decodeJSON(resp.Body, &result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, uint(1), result.SenderID)
		assert.NotNil(t, result.FileID)
		assert.Equal(t, uint(10), *result.FileID)
		assert.Equal(t, models.MessageTypeImage, result.Type)
	})

	t.Run("sends file to group conversation", func(t *testing.T) {
		fileSvc := &mockFileSvc{
			uploadFileFn: func(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
				return &models.File{ID: 11, URL: "/uploads/doc.pdf", Size: 500, Type: "application/pdf"}, nil
			},
		}
		msgSvc := &mockMsgSvc{
			sendMessageFn: func(msg *models.Message) error {
				msg.ID = 2
				return nil
			},
		}
		r := setupMessageRoutes(msgSvc, fileSvc)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		require.NoError(t, filePart(w, "file", "document.pdf", "application/pdf", "fake-pdf-data"))
		w.WriteField("sender_id", "1")
		w.WriteField("group_id", "5")
		w.Close()

		req := httptest.NewRequest("POST", "/messages/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusCreated, resp.Code)

		var result models.Message
		decodeJSON(resp.Body, &result)
		assert.Equal(t, uint(2), result.ID)
		assert.Equal(t, uint(11), *result.FileID)
		assert.Equal(t, models.MessageTypeFile, result.Type)
	})

	t.Run("missing sender_id", func(t *testing.T) {
		r := setupMessageRoutes(&mockMsgSvc{}, &mockFileSvc{})

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		require.NoError(t, filePart(w, "file", "test.png", "image/png", "data"))
		w.Close()

		req := httptest.NewRequest("POST", "/messages/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("missing file", func(t *testing.T) {
		r := setupMessageRoutes(&mockMsgSvc{}, &mockFileSvc{})

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		w.WriteField("sender_id", "1")
		w.Close()

		req := httptest.NewRequest("POST", "/messages/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})

	t.Run("file upload fails", func(t *testing.T) {
		fileSvc := &mockFileSvc{
			uploadFileFn: func(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
				return nil, assert.AnError
			},
		}
		r := setupMessageRoutes(&mockMsgSvc{}, fileSvc)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		require.NoError(t, filePart(w, "file", "test.png", "image/png", "data"))
		w.WriteField("sender_id", "1")
		w.WriteField("receiver_id", "2")
		w.Close()

		req := httptest.NewRequest("POST", "/messages/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})

	t.Run("message send fails", func(t *testing.T) {
		fileSvc := &mockFileSvc{
			uploadFileFn: func(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
				return &models.File{ID: 99, URL: "/uploads/cleanup.txt", Size: 10, Type: "text/plain"}, nil
			},
		}
		msgSvc := &mockMsgSvc{
			sendMessageFn: func(msg *models.Message) error {
				return service.ErrInvalidConversation
			},
		}
		r := setupMessageRoutes(msgSvc, fileSvc)

		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		require.NoError(t, filePart(w, "file", "test.txt", "text/plain", "data"))
		w.WriteField("sender_id", "1")
		w.Close()

		req := httptest.NewRequest("POST", "/messages/upload", &buf)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusBadRequest, resp.Code)
	})
}
