package service

import (
	"errors"
	"time"

	"chat/models"
	"chat/repository"
)

type MessageService interface {
	SendMessage(msg *models.Message) error
	GetConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error)
	MarkSeen(messageID, userID uint) error
	DeleteMessage(messageID uint) error
	GetUnseenCount(userID uint) (map[string]interface{}, error)
}

type messageService struct {
	msgRepo   repository.MessageRepository
	groupRepo repository.GroupRepository
}

func NewMessageService(msgRepo repository.MessageRepository, groupRepo repository.GroupRepository) MessageService {
	return &messageService{msgRepo: msgRepo, groupRepo: groupRepo}
}

// SendMessage stores a new message after validating the conversation target.
func (s *messageService) SendMessage(msg *models.Message) error {
	if msg.ReceiverID == nil && msg.GroupID == nil {
		return ErrInvalidConversation
	}
	// For group messages, verify the sender is a member of the group
	if msg.GroupID != nil {
		isMember, err := s.groupRepo.IsMember(*msg.GroupID, msg.SenderID)
		if err != nil {
			return err
		}
		if !isMember {
			return ErrNotGroupMember
		}
	}
	msg.CreatedAt = time.Now()
	return s.msgRepo.CreateMessage(msg)
}

// GetConversation fetches paginated messages for either a direct or group chat.
func (s *messageService) GetConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error) {
	return s.msgRepo.GetMessagesByConversation(receiverID, groupID, limit, offset)
}

// MarkSeen sets the seen flag for a user on a specific message.
func (s *messageService) MarkSeen(messageID, userID uint) error {
	return s.msgRepo.MarkMessageSeen(messageID, userID)
}

// DeleteMessage removes a message by ID.
func (s *messageService) DeleteMessage(messageID uint) error {
	return s.msgRepo.DeleteMessage(messageID)
}

// GetUnseenCount fetches the count of unseen direct and group messages for a user, grouped by sender/group.
func (s *messageService) GetUnseenCount(userID uint) (map[string]interface{}, error) {
	return s.msgRepo.GetUnseenCountDetailed(userID)
}

var (
	// ErrInvalidConversation signals a missing conversation target.
	ErrInvalidConversation = errors.New("either receiverID or groupID must be provided")
	// ErrNotGroupMember signals the sender is not a member of the target group.
	ErrNotGroupMember = errors.New("sender is not a member of this group")
)
