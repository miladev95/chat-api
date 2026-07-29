package handler

import (
	"chat/models"
	"chat/service"
)

// compile-time check that mock services implement the interfaces
var _ service.UserService = (*mockUserSvc)(nil)
var _ service.GroupService = (*mockGroupSvc)(nil)
var _ service.MessageService = (*mockMsgSvc)(nil)

// --- UserService mock ---

type mockUserSvc struct {
	createUserFn  func(user *models.User) error
	getUserByIDFn func(userID uint) (*models.User, error)
	getAllUsersFn func() ([]models.User, error)
}

func (m *mockUserSvc) CreateUser(user *models.User) error   { return m.createUserFn(user) }
func (m *mockUserSvc) GetUserByID(userID uint) (*models.User, error) { return m.getUserByIDFn(userID) }
func (m *mockUserSvc) GetAllUsers() ([]models.User, error) { return m.getAllUsersFn() }

// --- GroupService mock ---

type mockGroupSvc struct {
	createGroupFn  func(name string) (*models.Group, error)
	deleteGroupFn  func(groupID uint) error
	addMemberFn    func(groupID, userID uint, role string) error
	removeMemberFn func(groupID, userID uint) error
	getMembersFn   func(groupID uint) ([]models.GroupMember, error)
	getGroupByIDFn func(groupID uint) (*models.Group, error)
}

func (m *mockGroupSvc) CreateGroup(name string) (*models.Group, error)    { return m.createGroupFn(name) }
func (m *mockGroupSvc) DeleteGroup(groupID uint) error                   { return m.deleteGroupFn(groupID) }
func (m *mockGroupSvc) AddMember(groupID, userID uint, role string) error { return m.addMemberFn(groupID, userID, role) }
func (m *mockGroupSvc) RemoveMember(groupID, userID uint) error           { return m.removeMemberFn(groupID, userID) }
func (m *mockGroupSvc) GetMembers(groupID uint) ([]models.GroupMember, error) { return m.getMembersFn(groupID) }
func (m *mockGroupSvc) GetGroupByID(groupID uint) (*models.Group, error) { return m.getGroupByIDFn(groupID) }

// --- MessageService mock ---

type mockMsgSvc struct {
	sendMessageFn     func(msg *models.Message) error
	getConversationFn func(receiverID, groupID *uint, limit, offset int) ([]models.Message, error)
	markSeenFn        func(messageID, userID uint) error
	deleteMessageFn   func(messageID uint) error
	getUnseenCountFn  func(userID uint) (map[string]interface{}, error)
}

func (m *mockMsgSvc) SendMessage(msg *models.Message) error                       { return m.sendMessageFn(msg) }
func (m *mockMsgSvc) GetConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error) {
	return m.getConversationFn(receiverID, groupID, limit, offset)
}
func (m *mockMsgSvc) MarkSeen(messageID, userID uint) error                       { return m.markSeenFn(messageID, userID) }
func (m *mockMsgSvc) DeleteMessage(messageID uint) error                         { return m.deleteMessageFn(messageID) }
func (m *mockMsgSvc) GetUnseenCount(userID uint) (map[string]interface{}, error) { return m.getUnseenCountFn(userID) }
