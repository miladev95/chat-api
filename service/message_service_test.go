package service

import (
	"errors"
	"testing"

	"chat/models"
	"chat/repository"

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

func TestMessageService_GetConversations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := []repository.ConversationItem{
			{UserID: 2, Username: "bob", UnseenCount: 3},
			{UserID: 3, Username: "carol", UnseenCount: 1},
		}
		msgRepo := &mockMessageRepo{
			getConversationsFn: func(uid uint) ([]repository.ConversationItem, error) {
				return expected, nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		convs, err := svc.GetConversations(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, convs)
	})

	t.Run("repo error", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			getConversationsFn: func(uid uint) ([]repository.ConversationItem, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		convs, err := svc.GetConversations(1)
		assert.Error(t, err)
		assert.Nil(t, convs)
	})
}

func TestMessageService_EditMessage(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var capturedContent string
		msgRepo := &mockMessageRepo{
			getMessageByIDFn: func(mid uint) (*models.Message, error) {
				return &models.Message{ID: 1, SenderID: 5, Content: "Original"}, nil
			},
			updateMessageContentFn: func(mid uint, content string) error {
				capturedContent = content
				return nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.EditMessage(1, 5, "Updated content")
		assert.NoError(t, err)
		assert.Equal(t, "Updated content", capturedContent)
	})

	t.Run("empty content rejected", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			getMessageByIDFn: func(mid uint) (*models.Message, error) {
				return &models.Message{ID: 1, SenderID: 5}, nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.EditMessage(1, 5, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content cannot be empty")
	})

	t.Run("message not found", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			getMessageByIDFn: func(mid uint) (*models.Message, error) {
				return nil, errors.New("not found")
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.EditMessage(999, 5, "new content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "message not found")
	})

	t.Run("cannot edit other user's message", func(t *testing.T) {
		msgRepo := &mockMessageRepo{
			getMessageByIDFn: func(mid uint) (*models.Message, error) {
				return &models.Message{ID: 1, SenderID: 5, Content: "Original"}, nil
			},
		}
		svc := NewMessageService(msgRepo, &mockGroupRepo{})

		err := svc.EditMessage(1, 99, "hacked content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "only edit your own messages")
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
