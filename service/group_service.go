package service

import (
	"chat/models"
	"chat/repository"
)

type GroupService interface {
	CreateGroup(name string) (*models.Group, error)
	DeleteGroup(groupID uint) error
	AddMember(groupID, userID uint, role string) error
	RemoveMember(groupID, userID uint) error
	GetMembers(groupID uint) ([]models.GroupMember, error)
	GetGroupByID(groupID uint) (*models.Group, error)
}

type groupService struct {
	groupRepo repository.GroupRepository
}

func NewGroupService(groupRepo repository.GroupRepository) GroupService {
	return &groupService{groupRepo: groupRepo}
}

func (s *groupService) CreateGroup(name string) (*models.Group, error) {
	group := &models.Group{Name: name}
	if err := s.groupRepo.CreateGroup(group); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *groupService) DeleteGroup(groupID uint) error {
	return s.groupRepo.DeleteGroup(groupID)
}

func (s *groupService) AddMember(groupID, userID uint, role string) error {
	if role == "" {
		role = models.RoleMember
	}
	return s.groupRepo.AddMember(groupID, userID, role)
}

func (s *groupService) RemoveMember(groupID, userID uint) error {
	return s.groupRepo.RemoveMember(groupID, userID)
}

func (s *groupService) GetMembers(groupID uint) ([]models.GroupMember, error) {
	return s.groupRepo.GetMembers(groupID)
}

func (s *groupService) GetGroupByID(groupID uint) (*models.Group, error) {
	return s.groupRepo.GetGroupByID(groupID)
}
