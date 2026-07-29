package service

import (
	"errors"

	"chat/models"
)

// --- UserRepository mock ---

type mockUserRepo struct {
	createUserFn  func(user *models.User) error
	getUserByIDFn func(userID uint) (*models.User, error)
	getAllUsersFn func() ([]models.User, error)
}

func (m *mockUserRepo) CreateUser(user *models.User) error {
	return m.createUserFn(user)
}

func (m *mockUserRepo) GetUserByID(userID uint) (*models.User, error) {
	return m.getUserByIDFn(userID)
}

func (m *mockUserRepo) GetAllUsers() ([]models.User, error) {
	return m.getAllUsersFn()
}

// --- GroupRepository mock ---

type mockGroupRepo struct {
	createGroupFn  func(group *models.Group) error
	getGroupByIDFn func(groupID uint) (*models.Group, error)
	addMemberFn    func(groupID, userID uint, role string) error
	removeMemberFn func(groupID, userID uint) error
	getMembersFn   func(groupID uint) ([]models.GroupMember, error)
	deleteGroupFn  func(groupID uint) error
}

func (m *mockGroupRepo) CreateGroup(group *models.Group) error {
	return m.createGroupFn(group)
}

func (m *mockGroupRepo) GetGroupByID(groupID uint) (*models.Group, error) {
	return m.getGroupByIDFn(groupID)
}

func (m *mockGroupRepo) AddMember(groupID, userID uint, role string) error {
	return m.addMemberFn(groupID, userID, role)
}

func (m *mockGroupRepo) RemoveMember(groupID, userID uint) error {
	return m.removeMemberFn(groupID, userID)
}

func (m *mockGroupRepo) GetMembers(groupID uint) ([]models.GroupMember, error) {
	return m.getMembersFn(groupID)
}

func (m *mockGroupRepo) DeleteGroup(groupID uint) error {
	return m.deleteGroupFn(groupID)
}

// --- MessageRepository mock ---

type mockMessageRepo struct {
	createMessageFn              func(msg *models.Message) error
	getMessagesByConversationFn  func(receiverID, groupID *uint, limit, offset int) ([]models.Message, error)
	markMessageSeenFn            func(messageID, userID uint) error
	deleteMessageFn              func(messageID uint) error
	getUnseenCountDetailedFn     func(userID uint) (map[string]interface{}, error)
}

func (m *mockMessageRepo) CreateMessage(msg *models.Message) error {
	return m.createMessageFn(msg)
}

func (m *mockMessageRepo) GetMessagesByConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error) {
	return m.getMessagesByConversationFn(receiverID, groupID, limit, offset)
}

func (m *mockMessageRepo) MarkMessageSeen(messageID, userID uint) error {
	return m.markMessageSeenFn(messageID, userID)
}

func (m *mockMessageRepo) DeleteMessage(messageID uint) error {
	return m.deleteMessageFn(messageID)
}

func (m *mockMessageRepo) GetUnseenCountDetailed(userID uint) (map[string]interface{}, error) {
	return m.getUnseenCountDetailedFn(userID)
}

// --- Shared helpers ---

var errTest = errors.New("test error")
