package repository

import (
	"encoding/json"
	"testing"

	"chat/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// helpers to turn literal values into pointers
func uintPtr(v uint) *uint { return &v }

// jsonRoundTrip marshals and unmarshals a value to produce JSON-safe map structures.
// This is needed because GetUnseenCountDetailed returns function-local types inside
// map[string]interface{} that cannot be type-asserted from outside the function.
func jsonRoundTrip(t *testing.T, v interface{}) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err, "json.Marshal should succeed")
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err, "json.Unmarshal should succeed")
	return result
}

func TestMessageRepository_CreateMessage(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	msgRepo := NewMessageRepository(db)

	user1 := &models.User{Username: "sender", Email: "sender@example.com"}
	user2 := &models.User{Username: "receiver", Email: "receiver@example.com"}
	assert.NoError(t, userRepo.CreateUser(user1))
	assert.NoError(t, userRepo.CreateUser(user2))

	t.Run("creates a direct text message", func(t *testing.T) {
		msg := &models.Message{
			SenderID:   user1.ID,
			ReceiverID: uintPtr(user2.ID),
			Type:       models.MessageTypeText,
			Content:    "Hello!",
		}

		err := msgRepo.CreateMessage(msg)
		assert.NoError(t, err)
		assert.NotZero(t, msg.ID)
		assert.Equal(t, "Hello!", msg.Content)
	})

	t.Run("creates a group text message", func(t *testing.T) {
		groupRepo := NewGroupRepository(db)
		group := &models.Group{Name: "Chat Room"}
		assert.NoError(t, groupRepo.CreateGroup(group))

		msg := &models.Message{
			SenderID: user1.ID,
			GroupID:  uintPtr(group.ID),
			Type:     models.MessageTypeText,
			Content:  "Hello team!",
		}

		err := msgRepo.CreateMessage(msg)
		assert.NoError(t, err)
		assert.NotZero(t, msg.ID)
	})

	t.Run("creates a message with file attachment", func(t *testing.T) {
		file := &models.File{URL: "https://example.com/file.pdf", Size: 1024, Type: "application/pdf"}
		assert.NoError(t, db.Create(file).Error)

		msg := &models.Message{
			SenderID:   user1.ID,
			ReceiverID: uintPtr(user2.ID),
			Type:       models.MessageTypeFile,
			Content:    "Here is the document",
			FileID:     uintPtr(file.ID),
		}

		err := msgRepo.CreateMessage(msg)
		assert.NoError(t, err)
		assert.NotZero(t, msg.ID)
	})
}

func TestMessageRepository_GetMessagesByConversation(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	msgRepo := NewMessageRepository(db)

	alice := &models.User{Username: "alice", Email: "alice@test.com"}
	bob := &models.User{Username: "bob", Email: "bob@test.com"}
	assert.NoError(t, userRepo.CreateUser(alice))
	assert.NoError(t, userRepo.CreateUser(bob))

	// Create some direct messages
	for i := 0; i < 5; i++ {
		msg := &models.Message{
			SenderID:   alice.ID,
			ReceiverID: uintPtr(bob.ID),
			Type:       models.MessageTypeText,
			Content:    "Message",
		}
		assert.NoError(t, msgRepo.CreateMessage(msg))
	}

	t.Run("returns messages for direct conversation", func(t *testing.T) {
		msgs, err := msgRepo.GetMessagesByConversation(uintPtr(bob.ID), nil, 50, 0)
		assert.NoError(t, err)
		assert.Len(t, msgs, 5)
		for _, m := range msgs {
			assert.Equal(t, alice.ID, m.SenderID)
			assert.NotNil(t, m.ReceiverID)
			assert.Equal(t, bob.ID, *m.ReceiverID)
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		msgs, err := msgRepo.GetMessagesByConversation(uintPtr(bob.ID), nil, 2, 0)
		assert.NoError(t, err)
		assert.Len(t, msgs, 2)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		msgs, err := msgRepo.GetMessagesByConversation(uintPtr(bob.ID), nil, 50, 3)
		assert.NoError(t, err)
		assert.Len(t, msgs, 2) // 5 total - 3 offset = 2 remaining
	})

	t.Run("returns empty slice for conversation with no messages", func(t *testing.T) {
		carol := &models.User{Username: "carol", Email: "carol@test.com"}
		assert.NoError(t, userRepo.CreateUser(carol))

		msgs, err := msgRepo.GetMessagesByConversation(uintPtr(carol.ID), nil, 50, 0)
		assert.NoError(t, err)
		assert.Empty(t, msgs)
	})

	t.Run("returns messages for group conversation", func(t *testing.T) {
		groupRepo := NewGroupRepository(db)
		group := &models.Group{Name: "Group Chat"}
		assert.NoError(t, groupRepo.CreateGroup(group))

		for i := 0; i < 3; i++ {
			msg := &models.Message{
				SenderID: alice.ID,
				GroupID:  uintPtr(group.ID),
				Type:     models.MessageTypeText,
				Content:  "Group message",
			}
			assert.NoError(t, msgRepo.CreateMessage(msg))
		}

		msgs, err := msgRepo.GetMessagesByConversation(nil, uintPtr(group.ID), 50, 0)
		assert.NoError(t, err)
		assert.Len(t, msgs, 3)
	})

	t.Run("returns messages in ascending order by created_at", func(t *testing.T) {
		msgs, err := msgRepo.GetMessagesByConversation(uintPtr(bob.ID), nil, 50, 0)
		assert.NoError(t, err)
		for i := 1; i < len(msgs); i++ {
			assert.True(t,
				msgs[i-1].CreatedAt.Before(msgs[i].CreatedAt) ||
					msgs[i-1].CreatedAt.Equal(msgs[i].CreatedAt),
				"messages should be in ascending order")
		}
	})
}

func TestMessageRepository_MarkMessageSeen(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	msgRepo := NewMessageRepository(db)

	sender := &models.User{Username: "sender", Email: "sender@test.com"}
	receiver := &models.User{Username: "receiver", Email: "receiver@test.com"}
	assert.NoError(t, userRepo.CreateUser(sender))
	assert.NoError(t, userRepo.CreateUser(receiver))

	msg := &models.Message{
		SenderID:   sender.ID,
		ReceiverID: uintPtr(receiver.ID),
		Type:       models.MessageTypeText,
		Content:    "Read this!",
	}
	assert.NoError(t, msgRepo.CreateMessage(msg))

	t.Run("marks a message as seen", func(t *testing.T) {
		err := msgRepo.MarkMessageSeen(msg.ID, receiver.ID)
		assert.NoError(t, err)

		// Verify the seen record exists
		var seen models.Seen
		err = db.Where("message_id = ? AND user_id = ?", msg.ID, receiver.ID).First(&seen).Error
		assert.NoError(t, err)
		assert.NotZero(t, seen.SeenAt)
	})

	t.Run("upserts when marking already seen", func(t *testing.T) {
		// Mark again — should not error (upsert)
		err := msgRepo.MarkMessageSeen(msg.ID, receiver.ID)
		assert.NoError(t, err)
	})

	t.Run("allows multiple users to mark the same message as seen", func(t *testing.T) {
		user3 := &models.User{Username: "third", Email: "third@test.com"}
		assert.NoError(t, userRepo.CreateUser(user3))

		err := msgRepo.MarkMessageSeen(msg.ID, user3.ID)
		assert.NoError(t, err)

		// Verify both seen records exist
		assert.NoError(t, db.Where("message_id = ? AND user_id = ?", msg.ID, receiver.ID).First(&models.Seen{}).Error)
		assert.NoError(t, db.Where("message_id = ? AND user_id = ?", msg.ID, user3.ID).First(&models.Seen{}).Error)
	})
}

func TestMessageRepository_DeleteMessage(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	msgRepo := NewMessageRepository(db)

	user := &models.User{Username: "user", Email: "user@test.com"}
	assert.NoError(t, userRepo.CreateUser(user))

	t.Run("soft deletes a message", func(t *testing.T) {
		msg := &models.Message{
			SenderID:   user.ID,
			ReceiverID: uintPtr(user.ID),
			Type:       models.MessageTypeText,
			Content:    "Delete me",
		}
		assert.NoError(t, msgRepo.CreateMessage(msg))

		err := msgRepo.DeleteMessage(msg.ID)
		assert.NoError(t, err)

		// Should not be findable
		var found models.Message
		err = db.First(&found, msg.ID).Error
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("deleting non-existent message does not error", func(t *testing.T) {
		err := msgRepo.DeleteMessage(9999)
		assert.NoError(t, err)
	})
}

func TestMessageRepository_GetUnseenCountDetailed(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)
	groupRepo := NewGroupRepository(db)
	msgRepo := NewMessageRepository(db)

	alice := &models.User{Username: "alice", Email: "alice@test.com"}
	bob := &models.User{Username: "bob", Email: "bob@test.com"}
	carol := &models.User{Username: "carol", Email: "carol@test.com"}
	assert.NoError(t, userRepo.CreateUser(alice))
	assert.NoError(t, userRepo.CreateUser(bob))
	assert.NoError(t, userRepo.CreateUser(carol))

	group := &models.Group{Name: "Team Group"}
	assert.NoError(t, groupRepo.CreateGroup(group))
	assert.NoError(t, groupRepo.AddMember(group.ID, alice.ID, models.RoleMember))
	assert.NoError(t, groupRepo.AddMember(group.ID, bob.ID, models.RoleMember))
	assert.NoError(t, groupRepo.AddMember(group.ID, carol.ID, models.RoleMember))

	t.Run("returns zero counts when there are no unseen messages", func(t *testing.T) {
		result, err := msgRepo.GetUnseenCountDetailed(alice.ID)
		assert.NoError(t, err)

		parsed := jsonRoundTrip(t, result)

		direct := parsed["direct_messages"].([]interface{})
		groupMsgs := parsed["group_messages"].([]interface{})
		total := parsed["total"].(float64)

		assert.Empty(t, direct)
		assert.Empty(t, groupMsgs)
		assert.Equal(t, float64(0), total)
	})

	t.Run("counts unseen direct messages grouped by sender", func(t *testing.T) {
		// Alice sends 3 messages to Bob
		for i := 0; i < 3; i++ {
			msg := &models.Message{
				SenderID:   alice.ID,
				ReceiverID: uintPtr(bob.ID),
				Type:       models.MessageTypeText,
				Content:    "Hey Bob",
			}
			assert.NoError(t, msgRepo.CreateMessage(msg))
		}
		// Carol sends 2 messages to Bob
		for i := 0; i < 2; i++ {
			msg := &models.Message{
				SenderID:   carol.ID,
				ReceiverID: uintPtr(bob.ID),
				Type:       models.MessageTypeText,
				Content:    "Hey Bob from Carol",
			}
			assert.NoError(t, msgRepo.CreateMessage(msg))
		}

		result, err := msgRepo.GetUnseenCountDetailed(bob.ID)
		assert.NoError(t, err)

		parsed := jsonRoundTrip(t, result)

		direct := parsed["direct_messages"].([]interface{})
		total := parsed["total"].(float64)

		assert.Len(t, direct, 2)
		assert.Equal(t, float64(5), total)

		// Verify grouped counts
		aliceGroup := direct[0].(map[string]interface{})
		carolGroup := direct[1].(map[string]interface{})
		assert.Equal(t, float64(alice.ID), aliceGroup["user_id"])
		assert.Equal(t, float64(3), aliceGroup["unseen_count"])
		assert.Equal(t, float64(carol.ID), carolGroup["user_id"])
		assert.Equal(t, float64(2), carolGroup["unseen_count"])
	})

	t.Run("excludes seen messages from count", func(t *testing.T) {
		// Bob marks Alice's first message as seen
		var firstMsg models.Message
		db.Where("sender_id = ? AND receiver_id = ?", alice.ID, bob.ID).First(&firstMsg)
		assert.NoError(t, msgRepo.MarkMessageSeen(firstMsg.ID, bob.ID))

		result, err := msgRepo.GetUnseenCountDetailed(bob.ID)
		assert.NoError(t, err)

		parsed := jsonRoundTrip(t, result)

		total := parsed["total"].(float64)
		assert.Equal(t, float64(4), total, "one of 5 messages was seen")
	})

	t.Run("counts unseen group messages", func(t *testing.T) {
		// Alice and Carol send messages to group (Bob is also a member)
		for i := 0; i < 2; i++ {
			msg := &models.Message{
				SenderID: alice.ID,
				GroupID:  uintPtr(group.ID),
				Type:     models.MessageTypeText,
				Content:  "Group msg from Alice",
			}
			assert.NoError(t, msgRepo.CreateMessage(msg))
		}
		for i := 0; i < 3; i++ {
			msg := &models.Message{
				SenderID: carol.ID,
				GroupID:  uintPtr(group.ID),
				Type:     models.MessageTypeText,
				Content:  "Group msg from Carol",
			}
			assert.NoError(t, msgRepo.CreateMessage(msg))
		}

		result, err := msgRepo.GetUnseenCountDetailed(bob.ID)
		assert.NoError(t, err)

		parsed := jsonRoundTrip(t, result)

		groupMsgs := parsed["group_messages"].([]interface{})
		assert.Len(t, groupMsgs, 1) // one group
		groupEntry := groupMsgs[0].(map[string]interface{})
		assert.Equal(t, float64(group.ID), groupEntry["group_id"])
		assert.Equal(t, float64(5), groupEntry["unseen_count"])
	})

	t.Run("excludes own messages from group unseen count", func(t *testing.T) {
		// Bob sends a message to the group — should not count as unseen for Bob
		msg := &models.Message{
			SenderID: bob.ID,
			GroupID:  uintPtr(group.ID),
			Type:     models.MessageTypeText,
			Content:  "Bob's own message",
		}
		assert.NoError(t, msgRepo.CreateMessage(msg))

		result, err := msgRepo.GetUnseenCountDetailed(bob.ID)
		assert.NoError(t, err)

		parsed := jsonRoundTrip(t, result)

		groupMsgs := parsed["group_messages"].([]interface{})
		groupEntry := groupMsgs[0].(map[string]interface{})
		// Should still be 5, not 6
		assert.Equal(t, float64(5), groupEntry["unseen_count"])
	})
}
