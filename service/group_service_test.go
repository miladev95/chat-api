package service

import (
	"testing"

	"chat/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGroupService_CreateGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockGroupRepo{
			createGroupFn: func(g *models.Group) error {
				g.ID = 1
				return nil
			},
		}
		svc := NewGroupService(repo)

		group, err := svc.CreateGroup("Test Group")
		assert.NoError(t, err)
		assert.Equal(t, uint(1), group.ID)
		assert.Equal(t, "Test Group", group.Name)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockGroupRepo{
			createGroupFn: func(g *models.Group) error {
				return errTest
			},
		}
		svc := NewGroupService(repo)

		group, err := svc.CreateGroup("Test Group")
		assert.Error(t, err)
		assert.Nil(t, group)
	})
}

func TestGroupService_DeleteGroup(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockGroupRepo{
			deleteGroupFn: func(id uint) error {
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.DeleteGroup(1)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockGroupRepo{
			deleteGroupFn: func(id uint) error {
				return errTest
			},
		}
		svc := NewGroupService(repo)

		err := svc.DeleteGroup(1)
		assert.Error(t, err)
	})
}

func TestGroupService_AddMember(t *testing.T) {
	t.Run("success with explicit role", func(t *testing.T) {
		var capturedRole string
		repo := &mockGroupRepo{
			addMemberFn: func(gid, uid uint, role string) error {
				capturedRole = role
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleAdmin)
		assert.NoError(t, err)
		assert.Equal(t, models.RoleAdmin, capturedRole)
	})

	t.Run("defaults to member role when empty", func(t *testing.T) {
		var capturedRole string
		repo := &mockGroupRepo{
			addMemberFn: func(gid, uid uint, role string) error {
				capturedRole = role
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, "")
		assert.NoError(t, err)
		assert.Equal(t, models.RoleMember, capturedRole)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockGroupRepo{
			addMemberFn: func(gid, uid uint, role string) error {
				return errTest
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleMember)
		assert.Error(t, err)
	})
}

func TestGroupService_RemoveMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockGroupRepo{
			removeMemberFn: func(gid, uid uint) error {
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.RemoveMember(1, 2)
		assert.NoError(t, err)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockGroupRepo{
			removeMemberFn: func(gid, uid uint) error {
				return errTest
			},
		}
		svc := NewGroupService(repo)

		err := svc.RemoveMember(1, 2)
		assert.Error(t, err)
	})
}

func TestGroupService_GetMembers(t *testing.T) {
	t.Run("returns members", func(t *testing.T) {
		expected := []models.GroupMember{
			{GroupID: 1, UserID: 10, Role: models.RoleMember},
			{GroupID: 1, UserID: 20, Role: models.RoleAdmin},
		}
		repo := &mockGroupRepo{
			getMembersFn: func(gid uint) ([]models.GroupMember, error) {
				return expected, nil
			},
		}
		svc := NewGroupService(repo)

		members, err := svc.GetMembers(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, members)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMembersFn: func(gid uint) ([]models.GroupMember, error) {
				return nil, errTest
			},
		}
		svc := NewGroupService(repo)

		members, err := svc.GetMembers(1)
		assert.Error(t, err)
		assert.Nil(t, members)
	})
}

func TestGroupService_GetGroupByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		expected := &models.Group{ID: 1, Name: "My Group"}
		repo := &mockGroupRepo{
			getGroupByIDFn: func(id uint) (*models.Group, error) {
				return expected, nil
			},
		}
		svc := NewGroupService(repo)

		group, err := svc.GetGroupByID(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, group)
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockGroupRepo{
			getGroupByIDFn: func(id uint) (*models.Group, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		group, err := svc.GetGroupByID(999)
		assert.Error(t, err)
		assert.Nil(t, group)
	})
}
