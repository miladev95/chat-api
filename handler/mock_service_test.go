package handler

import (
	"mime/multipart"

	"chat/models"
	"chat/service"
)

// compile-time check that mock services implement the interfaces
var _ service.UserService = (*mockUserSvc)(nil)
var _ service.GroupService = (*mockGroupSvc)(nil)
var _ service.MessageService = (*mockMsgSvc)(nil)
var _ service.FileService = (*mockFileSvc)(nil)

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
	createGroupFn     func(name string) (*models.Group, error)
	deleteGroupFn     func(groupID, requesterID uint) error
	addMemberFn       func(groupID, userID uint, role string, requesterID uint) error
	removeMemberFn    func(groupID, userID uint, requesterID uint) error
	updateMemberRoleFn func(groupID, userID, requesterID uint, role string) error
	getMembersFn      func(groupID uint) ([]models.GroupMember, error)
	getGroupByIDFn    func(groupID uint) (*models.Group, error)
	getAllGroupsFn    func() ([]models.Group, error)
	updateGroupFn     func(groupID, requesterID uint, name string) error
	isMemberFn        func(groupID, userID uint) (bool, error)
	leaveGroupFn      func(groupID, userID uint) error
}

func (m *mockGroupSvc) CreateGroup(name string) (*models.Group, error)                  { return m.createGroupFn(name) }
func (m *mockGroupSvc) DeleteGroup(groupID, requesterID uint) error                     { return m.deleteGroupFn(groupID, requesterID) }
func (m *mockGroupSvc) AddMember(groupID, userID uint, role string, requesterID uint) error { return m.addMemberFn(groupID, userID, role, requesterID) }
func (m *mockGroupSvc) RemoveMember(groupID, userID uint, requesterID uint) error       { return m.removeMemberFn(groupID, userID, requesterID) }
func (m *mockGroupSvc) UpdateMemberRole(groupID, userID, requesterID uint, role string) error { return m.updateMemberRoleFn(groupID, userID, requesterID, role) }
func (m *mockGroupSvc) GetMembers(groupID uint) ([]models.GroupMember, error)           { return m.getMembersFn(groupID) }
func (m *mockGroupSvc) GetGroupByID(groupID uint) (*models.Group, error)                { return m.getGroupByIDFn(groupID) }
func (m *mockGroupSvc) GetAllGroups() ([]models.Group, error)                           { return m.getAllGroupsFn() }
func (m *mockGroupSvc) UpdateGroup(groupID, requesterID uint, name string) error        { return m.updateGroupFn(groupID, requesterID, name) }
func (m *mockGroupSvc) IsMember(groupID, userID uint) (bool, error)                     { return m.isMemberFn(groupID, userID) }
func (m *mockGroupSvc) LeaveGroup(groupID, userID uint) error                           { return m.leaveGroupFn(groupID, userID) }

// --- MessageService mock ---

type mockMsgSvc struct {
	sendMessageFn     func(msg *models.Message) error
	getConversationFn func(receiverID, groupID *uint, limit, offset int) ([]models.Message, error)
	markSeenFn        func(messageID, userID uint) error
	deleteMessageFn   func(messageID uint) error
	getUnseenCountFn  func(userID uint) (map[string]interface{}, error)
}

// --- FileService mock ---

type mockFileSvc struct {
	uploadFileFn func(file multipart.File, header *multipart.FileHeader) (*models.File, error)
}

func (m *mockFileSvc) UploadFile(file multipart.File, header *multipart.FileHeader) (*models.File, error) {
	return m.uploadFileFn(file, header)
}

func (m *mockMsgSvc) SendMessage(msg *models.Message) error                       { return m.sendMessageFn(msg) }
func (m *mockMsgSvc) GetConversation(receiverID, groupID *uint, limit, offset int) ([]models.Message, error) {
	return m.getConversationFn(receiverID, groupID, limit, offset)
}
func (m *mockMsgSvc) MarkSeen(messageID, userID uint) error                       { return m.markSeenFn(messageID, userID) }
func (m *mockMsgSvc) DeleteMessage(messageID uint) error                         { return m.deleteMessageFn(messageID) }
func (m *mockMsgSvc) GetUnseenCount(userID uint) (map[string]interface{}, error) { return m.getUnseenCountFn(userID) }
