package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"chat/models"
	"chat/service"
)

type MessageHandler struct {
	msgService  service.MessageService
	fileService service.FileService
}

func NewMessageHandler(msgService service.MessageService, fileService service.FileService) *MessageHandler {
	return &MessageHandler{msgService: msgService, fileService: fileService}
}

// POST /messages
// validMessageTypes contains the allowed message type values.
var validMessageTypes = map[string]bool{
	string(models.MessageTypeText):  true,
	string(models.MessageTypeImage): true,
	string(models.MessageTypeFile):  true,
}

// POST /messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req struct {
		SenderID   uint   `json:"sender_id" binding:"required"`
		ReceiverID *uint  `json:"receiver_id"`
		GroupID    *uint  `json:"group_id"`
		Type       string `json:"type" binding:"required"`
		Content    string `json:"content"`
		FileID     *uint  `json:"file_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate message type is one of the allowed values
	if !validMessageTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message type: must be text, image, or file"})
		return
	}

	// For text messages, content is required
	if req.Type == string(models.MessageTypeText) && req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required for text messages"})
		return
	}

	// For image/file messages, file_id is required
	if req.Type != string(models.MessageTypeText) && req.FileID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file_id is required for image and file messages"})
		return
	}

	msg := &models.Message{
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		GroupID:    req.GroupID,
		Type:       models.MessageType(req.Type),
		Content:    req.Content,
		FileID:     req.FileID,
	}
	if err := h.msgService.SendMessage(msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

// GET /messages?receiver_id=&group_id=&limit=&offset=
func (h *MessageHandler) GetConversation(c *gin.Context) {
	receiverIDStr := c.Query("receiver_id")
	groupIDStr := c.Query("group_id")

	var receiverID, groupID *uint
	if receiverIDStr != "" {
		id, err := strconv.Atoi(receiverIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiver_id"})
			return
		}
		tmp := uint(id)
		receiverID = &tmp
	}
	if groupIDStr != "" {
		id, err := strconv.Atoi(groupIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		tmp := uint(id)
		groupID = &tmp
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		limit = 50
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Validate at least one conversation target was provided
	if receiverID == nil && groupID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either receiver_id or group_id is required"})
		return
	}

	msgs, err := h.msgService.GetConversation(receiverID, groupID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

// POST /messages/:id/seen
func (h *MessageHandler) MarkSeen(c *gin.Context) {
	messageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.msgService.MarkSeen(uint(messageID), req.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked as seen"})
}

// POST /messages/upload — multipart: upload a file and send it as a chat message in one request.
// Fields: file, sender_id, receiver_id OR group_id, content (optional)
func (h *MessageHandler) SendFileMessage(c *gin.Context) {
	senderID, err := strconv.Atoi(c.PostForm("sender_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sender_id is required and must be numeric"})
		return
	}

	// Parse the uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Upload file and create metadata record
	fileRecord, err := h.fileService.UploadFile(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse conversation targets
	var receiverID, groupID *uint
	if rid := c.PostForm("receiver_id"); rid != "" {
		id, err := strconv.Atoi(rid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiver_id"})
			return
		}
		tmp := uint(id)
		receiverID = &tmp
	}
	if gid := c.PostForm("group_id"); gid != "" {
		id, err := strconv.Atoi(gid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group_id"})
			return
		}
		tmp := uint(id)
		groupID = &tmp
	}

	// Validate at least one conversation target
	if receiverID == nil && groupID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either receiver_id or group_id is required"})
		return
	}

	// Determine message type based on content type
	msgType := models.MessageTypeFile
	if strings.HasPrefix(header.Header.Get("Content-Type"), "image/") {
		msgType = models.MessageTypeImage
	}

	// Build and send the message
	msg := &models.Message{
		SenderID:   uint(senderID),
		ReceiverID: receiverID,
		GroupID:    groupID,
		Type:       msgType,
		Content:    c.PostForm("content"),
		FileID:     &fileRecord.ID,
	}

	if err := h.msgService.SendMessage(msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

// DELETE /messages/:id
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	messageID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message id"})
		return
	}
	if err := h.msgService.DeleteMessage(uint(messageID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "message deleted"})
}

// GET /messages/unseen/:user_id
func (h *MessageHandler) GetUnseenCount(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	counts, err := h.msgService.GetUnseenCount(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, counts)
}
