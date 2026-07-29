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
		repo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				msg.ID = 1
				savedMsg = msg
				return nil
			},
		}
		svc := NewMessageService(repo)

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

	t.Run("group message", func(t *testing.T) {
		repo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				msg.ID = 2
				return nil
			},
		}
		svc := NewMessageService(repo)

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

	t.Run("missing receiver and group", func(t *testing.T) {
		repo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				return nil
			},
		}
		svc := NewMessageService(repo)

		msg := &models.Message{
			SenderID: 1,
			Type:     models.MessageTypeText,
			Content:  "Orphan message",
		}
		err := svc.SendMessage(msg)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidConversation)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockMessageRepo{
			createMessageFn: func(msg *models.Message) error {
				return errors.New("db error")
			},
		}
		svc := NewMessageService(repo)

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
		repo := &mockMessageRepo{
			getMessagesByConversationFn: func(rid, gid *uint, limit, offset int) ([]models.Message, error) {
				return expected, nil
			},
		}
		svc := NewMessageService(repo)

		msgs, err := svc.GetConversation(uintPtr(2), nil, 50, 0)
		assert.NoError(t, err)
		assert.Equal(t, expected, msgs)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockMessageRepo{
			getMessagesByConversationFn: func(rid, gid *uint, limit, offset int) ([]models.Message, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewMessageService(repo)

		msgs, err := svc.GetConversation(uintPtr(2), nil, 50, 0)
		assert.Error(t, err)
		assert.Nil(t, msgs)
	})
}

func TestMessageService_MarkSeen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockMessageRepo{
			markMessageSeenFn: func(mid, uid uint) error {
				return nil
			},
		}
		svc := NewMessageService(repo)

		err := svc.MarkSeen(1, 2)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockMessageRepo{
			markMessageSeenFn: func(mid, uid uint) error {
				return errors.New("db error")
			},
		}
		svc := NewMessageService(repo)

		err := svc.MarkSeen(1, 2)
		assert.Error(t, err)
	})
}

func TestMessageService_DeleteMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockMessageRepo{
			deleteMessageFn: func(mid uint) error {
				return nil
			},
		}
		svc := NewMessageService(repo)

		err := svc.DeleteMessage(1)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockMessageRepo{
			deleteMessageFn: func(mid uint) error {
				return errors.New("db error")
			},
		}
		svc := NewMessageService(repo)

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
		repo := &mockMessageRepo{
			getUnseenCountDetailedFn: func(uid uint) (map[string]interface{}, error) {
				return expected, nil
			},
		}
		svc := NewMessageService(repo)

		result, err := svc.GetUnseenCount(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockMessageRepo{
			getUnseenCountDetailedFn: func(uid uint) (map[string]interface{}, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewMessageService(repo)

		result, err := svc.GetUnseenCount(1)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
