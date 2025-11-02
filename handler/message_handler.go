package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chat/models"
	"chat/service"
)

type MessageHandler struct {
	msgService service.MessageService
}

func NewMessageHandler(msgService service.MessageService) *MessageHandler {
	return &MessageHandler{msgService: msgService}
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var receiverID, groupID *uint
	if receiverIDStr != "" {
		id, _ := strconv.Atoi(receiverIDStr)
		tmp := uint(id)
		receiverID = &tmp
	}
	if groupIDStr != "" {
		id, _ := strconv.Atoi(groupIDStr)
		tmp := uint(id)
		groupID = &tmp
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
	messageID, _ := strconv.Atoi(c.Param("id"))
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

// DELETE /messages/:id
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	messageID, _ := strconv.Atoi(c.Param("id"))
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
