package repository

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"chat/models"
)

type MessageRepository interface {
	CreateMessage(msg *models.Message) error
	GetMessagesByConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error)
	MarkMessageSeen(messageID, userID uint) error
	DeleteMessage(messageID uint) error
	GetUnseenCountDetailed(userID uint) (map[string]interface{}, error)
}

type messageRepo struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepo{db: db}
}

// Create a new message
func (r *messageRepo) CreateMessage(msg *models.Message) error {
	return r.db.Create(msg).Error
}

// Get messages by conversation (private or group)
func (r *messageRepo) GetMessagesByConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error) {
	var messages []models.Message
	query := r.db.Preload("File").Order("created_at ASC").Limit(limit).Offset(offset)

	if receiverID != nil {
		query = query.Where("receiver_id = ?", *receiverID)
	} else if groupID != nil {
		query = query.Where("group_id = ?", *groupID)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// Mark a message as seen
func (r *messageRepo) MarkMessageSeen(messageID, userID uint) error {
	seen := models.Seen{
		MessageID: messageID,
		UserID:    userID,
		SeenAt:    time.Now(),
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}, {Name: "user_id"}},
		UpdateAll: true,
	}).Create(&seen).Error
}

// Soft delete a message by ID
func (r *messageRepo) DeleteMessage(messageID uint) error {
	return r.db.Delete(&models.Message{}, messageID).Error
}

// GetUnseenCountDetailed returns unseen messages count grouped by sender (direct) and by group
func (r *messageRepo) GetUnseenCountDetailed(userID uint) (map[string]interface{}, error) {
	type DirectMessage struct {
		SenderID    uint   `json:"user_id"`
		Username    string `json:"username"`
		UnseenCount int64  `json:"unseen_count"`
	}

	type GroupMessage struct {
		GroupID     uint   `json:"group_id"`
		GroupName   string `json:"group_name"`
		UnseenCount int64  `json:"unseen_count"`
	}

	var directMessages []DirectMessage
	var groupMessages []GroupMessage
	var totalCount int64

	// Get unseen direct messages per sender
	directQuery := r.db.
		Table("messages").
		Select("messages.sender_id, users.username, COUNT(messages.id) as unseen_count").
		Joins("LEFT JOIN users ON users.id = messages.sender_id").
		Where("messages.receiver_id = ?", userID).
		Where("messages.id NOT IN (SELECT message_id FROM seens WHERE user_id = ?)", userID).
		Group("messages.sender_id")

	if err := directQuery.Scan(&directMessages).Error; err != nil {
		return nil, err
	}

	// Get unseen group messages per group
	groupQuery := r.db.
		Table("messages").
		Select("messages.group_id, groups.name as group_name, COUNT(messages.id) as unseen_count").
		Joins("INNER JOIN group_members ON group_members.group_id = messages.group_id").
		Joins("LEFT JOIN groups ON groups.id = messages.group_id").
		Where("group_members.user_id = ?", userID).
		Where("messages.sender_id != ?", userID).
		Where("messages.id NOT IN (SELECT message_id FROM seens WHERE user_id = ?)", userID).
		Group("messages.group_id")

	if err := groupQuery.Scan(&groupMessages).Error; err != nil {
		return nil, err
	}

	// Calculate total
	for _, dm := range directMessages {
		totalCount += dm.UnseenCount
	}
	for _, gm := range groupMessages {
		totalCount += gm.UnseenCount
	}

	// If no data, return empty slices instead of nil
	if directMessages == nil {
		directMessages = []DirectMessage{}
	}
	if groupMessages == nil {
		groupMessages = []GroupMessage{}
	}

	return map[string]interface{}{
		"direct_messages": directMessages,
		"group_messages":  groupMessages,
		"total":           totalCount,
	}, nil
}
