package service

import (
	"errors"

	"chat/models"
	"chat/repository"
)

var (
	ErrNotMember       = errors.New("user is not a member of this group")
	ErrInsufficientRole = errors.New("insufficient role: owner or admin access required")
	ErrNotOwner         = errors.New("only the group owner can perform this action")
)

type GroupService interface {
	CreateGroup(name string) (*models.Group, error)
	DeleteGroup(groupID, requesterID uint) error
	AddMember(groupID, userID uint, role string, requesterID uint) error
	RemoveMember(groupID, userID uint, requesterID uint) error
	GetMembers(groupID uint) ([]models.GroupMember, error)
	GetGroupByID(groupID uint) (*models.Group, error)
	IsMember(groupID, userID uint) (bool, error)
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

func (s *groupService) DeleteGroup(groupID, requesterID uint) error {
	role, err := s.groupRepo.GetMemberRole(groupID, requesterID)
	if err != nil {
		return ErrNotMember
	}
	if role != models.RoleOwner {
		return ErrNotOwner
	}
	return s.groupRepo.DeleteGroup(groupID)
}

func (s *groupService) AddMember(groupID, userID uint, role string, requesterID uint) error {
	requesterRole, err := s.groupRepo.GetMemberRole(groupID, requesterID)
	if err != nil {
		return ErrNotMember
	}
	if requesterRole != models.RoleOwner && requesterRole != models.RoleAdmin {
		return ErrInsufficientRole
	}
	if role == "" {
		role = models.RoleMember
	}
	// Only owners can assign owner or admin roles
	if (role == models.RoleOwner || role == models.RoleAdmin) && requesterRole != models.RoleOwner {
		return ErrNotOwner
	}
	return s.groupRepo.AddMember(groupID, userID, role)
}

func (s *groupService) RemoveMember(groupID, userID uint, requesterID uint) error {
	requesterRole, err := s.groupRepo.GetMemberRole(groupID, requesterID)
	if err != nil {
		return ErrNotMember
	}
	if requesterRole != models.RoleOwner && requesterRole != models.RoleAdmin {
		return ErrInsufficientRole
	}
	// Cannot remove the owner
	targetRole, err := s.groupRepo.GetMemberRole(groupID, userID)
	if err == nil && targetRole == models.RoleOwner {
		return ErrNotOwner
	}
	return s.groupRepo.RemoveMember(groupID, userID)
}

func (s *groupService) GetMembers(groupID uint) ([]models.GroupMember, error) {
	return s.groupRepo.GetMembers(groupID)
}

func (s *groupService) GetGroupByID(groupID uint) (*models.Group, error) {
	return s.groupRepo.GetGroupByID(groupID)
}

func (s *groupService) IsMember(groupID, userID uint) (bool, error) {
	return s.groupRepo.IsMember(groupID, userID)
}
