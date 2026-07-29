package service

import (
	"errors"
	"testing"

	"chat/models"

	"github.com/stretchr/testify/assert"
)

func uintPtr(v uint) *uint { return &v }

func TestMessageService_SendMessage(t *testing.T) {
	t.Run("direct message", func(t *testing.T) {
		var savedMsg *models.Message
		msgRepo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				msg.ID = 1
				savedMsg = msg
				return nil
			},
		}
		groupRepo := &mockGroupRepo{}
		svc := NewMessageService(msgRepo, groupRepo)

		msg := &models.Message{
			SenderID:   1,
			ReceiverID: uintPtr(2),
			Type:       models.MessageTypeText,
			Content:    "Hello!",
		}
		err := svc.SendMessage(msg)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), msg.ID)
		assert.NotZero(t, savedMsg.CreatedAt)
	})

	t.Run("group message as member succeeds", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				msg.ID = 2
				return nil
			},
		}
		groupRepo := &mockGroupRepo{
			isMemberFn: func(gid, uid uint) (bool, error) {
				return true, nil
			},
		}
		svc := NewMessageService(msgRepo, groupRepo)

		msg := &models.Message{
			SenderID: 1,
			GroupID:  uintPtr(1),
			Type:     models.MessageTypeText,
			Content:  "Hello team!",
		}
		err := svc.SendMessage(msg)

		assert.NoError(t, err)
		assert.Equal(t, uint(2), msg.ID)
	})

	t.Run("group message as non-member fails", func(t *testing.T) {
		msgRepo := &mockMessageRepo{}
		groupRepo := &mockGroupRepo{
			isMemberFn: func(gid, uid uint) (bool, error) {
				return false, nil
			},
		}
		svc := NewMessageService(msgRepo, groupRepo)

		msg := &models.Message{
			SenderID: 99,
			GroupID:  uintPtr(1),
			Type:     models.MessageTypeText,
			Content:  "Unauthorized!",
		}
		err := svc.SendMessage(msg)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotGroupMember)
	})

	t.Run("missing receiver and group", func(t *testing.T) {
		msgRepo := &mockMessageRepo{}
		groupRepo := &mockGroupRepo{}
		svc := NewMessageService(msgRepo, groupRepo)

		msg := &models.Message{
			SenderID: 1,
			Type:     models.MessageTypeText,
			Content:  "Orphan message",
		}
		err := svc.SendMessage(msg)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConversation)
	})

	t.Run("repo error on create", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				return errors.New("db error")
			},
		}
		groupRepo := &mockGroupRepo{}
		svc := NewMessageService(msgRepo, groupRepo)

		msg := &models.Message{
			SenderID:   1,
			ReceiverID: uintPtr(2),
			Type:       models.MessageTypeText,
			Content:    "Hi",
		}
		err := svc.SendMessage(msg)
		assert.Error(t, err)
	})
}

func TestMessageService_GetConversation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := []models.Message{
			{ID: 1, SenderID: 1, ReceiverID: uintPtr(2)},
			{ID: 2, SenderID: 1, ReceiverID: uintPtr(2)},
		}
		msgRepo := &mockMessageRepo{
			getMessagesByConversationFn: func(rid, gid *uint, limit, offset int) ([]models.Message, error) {
				return expected, nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		msgs, err := svc.GetConversation(uintPtr(2), nil, 50, 0)
		assert.NoError(t, err)
		assert.Equal(t, expected, msgs)
	})

	t.Run("repo error", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			getMessagesByConversationFn: func(rid, gid *uint, limit, offset int) ([]models.Message, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		msgs, err := svc.GetConversation(uintPtr(2), nil, 50, 0)
		assert.Error(t, err)
		assert.Nil(t, msgs)
	})
}

func TestMessageService_MarkSeen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			markMessageSeenFn: func(mid, uid uint) error {
				return nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.MarkSeen(1, 2)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			markMessageSeenFn: func(mid, uid uint) error {
				return errors.New("db error")
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.MarkSeen(1, 2)
		assert.Error(t, err)
	})
}

func TestMessageService_DeleteMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			deleteMessageFn: func(mid uint) error {
				return nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.DeleteMessage(1)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			deleteMessageFn: func(mid uint) error {
				return errors.New("db error")
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.DeleteMessage(1)
		assert.Error(t, err)
	})
}

func TestMessageService_GetUnseenCount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := map[string]interface{}{
			"direct_messages": []interface{}{},
			"group_messages":  []interface{}{},
			"total":           int64(0),
		}
		msgRepo := &mockMessageRepo{
			getUnseenCountDetailedFn: func(uid uint) (map[string]interface{}, error) {
				return expected, nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		result, err := svc.GetUnseenCount(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("repo error", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			getUnseenCountDetailedFn: func(uid uint) (map[string]interface{}, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		result, err := svc.GetUnseenCount(1)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
