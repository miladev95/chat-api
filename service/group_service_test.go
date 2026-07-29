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
	t.Run("owner can delete", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleOwner, nil
			},
			deleteGroupFn: func(id uint) error {
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.DeleteGroup(1, 1)
		assert.NoError(t, err)
	})

	t.Run("non-owner cannot delete", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleMember, nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.DeleteGroup(1, 2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("non-member cannot delete", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return "", gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		err := svc.DeleteGroup(1, 99)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotMember)
	})
}

func TestGroupService_AddMember(t *testing.T) {
	t.Run("owner can add member", func(t *testing.T) {
		var capturedRole string
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleOwner, nil
			},
			addMemberFn: func(gid, uid uint, role string) error {
				capturedRole = role
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleMember, 1)
		assert.NoError(t, err)
		assert.Equal(t, models.RoleMember, capturedRole)
	})

	t.Run("admin can add member", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleAdmin, nil
			},
			addMemberFn: func(gid, uid uint, role string) error {
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleMember, 3)
		assert.NoError(t, err)
	})

	t.Run("admin cannot assign owner role", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleAdmin, nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleOwner, 3)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("defaults to member role when empty", func(t *testing.T) {
		var capturedRole string
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleOwner, nil
			},
			addMemberFn: func(gid, uid uint, role string) error {
				capturedRole = role
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, "", 1)
		assert.NoError(t, err)
		assert.Equal(t, models.RoleMember, capturedRole)
	})

	t.Run("non-member cannot add", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return "", gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleMember, 99)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotMember)
	})

	t.Run("member cannot add", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleMember, nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.AddMember(1, 2, models.RoleMember, 3)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInsufficientRole)
	})
}

func TestGroupService_RemoveMember(t *testing.T) {
	t.Run("owner can remove member", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				if uid == 1 { // requester
					return models.RoleOwner, nil
				}
				return models.RoleMember, nil // target
			},
			removeMemberFn: func(gid, uid uint) error {
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.RemoveMember(1, 2, 1)
		assert.NoError(t, err)
	})

	t.Run("cannot remove owner", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				if uid == 1 { // requester
					return models.RoleOwner, nil
				}
				return models.RoleOwner, nil // target is also owner
			},
		}
		svc := NewGroupService(repo)

		err := svc.RemoveMember(1, 2, 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("non-member cannot remove", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return "", gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		err := svc.RemoveMember(1, 2, 99)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotMember)
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

func TestGroupService_GetAllGroups(t *testing.T) {
	t.Run("returns all groups", func(t *testing.T) {
		expected := []models.Group{
			{ID: 1, Name: "Alpha"},
			{ID: 2, Name: "Beta"},
		}
		repo := &mockGroupRepo{
			getAllGroupsFn: func() ([]models.Group, error) {
				return expected, nil
			},
		}
		svc := NewGroupService(repo)

		groups, err := svc.GetAllGroups()
		assert.NoError(t, err)
		assert.Equal(t, expected, groups)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockGroupRepo{
			getAllGroupsFn: func() ([]models.Group, error) {
				return nil, errTest
			},
		}
		svc := NewGroupService(repo)

		groups, err := svc.GetAllGroups()
		assert.Error(t, err)
		assert.Nil(t, groups)
	})
}

func TestGroupService_UpdateGroup(t *testing.T) {
	t.Run("owner can update", func(t *testing.T) {
		var capturedName string
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleOwner, nil
			},
			updateGroupFn: func(gid uint, name string) error {
				capturedName = name
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateGroup(1, 1, "New Name")
		assert.NoError(t, err)
		assert.Equal(t, "New Name", capturedName)
	})

	t.Run("non-owner cannot update", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleMember, nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateGroup(1, 2, "New Name")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("non-member cannot update", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return "", gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateGroup(1, 99, "New Name")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotMember)
	})
}

func TestGroupService_UpdateMemberRole(t *testing.T) {
	t.Run("owner can change to admin", func(t *testing.T) {
		var capturedRole string
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				if uid == 1 {
					return models.RoleOwner, nil
				}
				return models.RoleMember, nil
			},
			updateMemberRoleFn: func(gid, uid uint, role string) error {
				capturedRole = role
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateMemberRole(1, 2, 1, models.RoleAdmin)
		assert.NoError(t, err)
		assert.Equal(t, models.RoleAdmin, capturedRole)
	})

	t.Run("admin can change to member", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				if uid == 1 {
					return models.RoleAdmin, nil
				}
				return models.RoleMember, nil
			},
			updateMemberRoleFn: func(gid, uid uint, role string) error {
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateMemberRole(1, 2, 1, models.RoleMember)
		assert.NoError(t, err)
	})

	t.Run("admin cannot assign owner role", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleAdmin, nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateMemberRole(1, 2, 1, models.RoleOwner)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("cannot change owner's role", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				if uid == 1 {
					return models.RoleOwner, nil
				}
				return models.RoleOwner, nil // target is also owner
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateMemberRole(1, 2, 1, models.RoleMember)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("non-member cannot change roles", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return "", gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		err := svc.UpdateMemberRole(1, 2, 99, models.RoleMember)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotMember)
	})
}

func TestGroupService_LeaveGroup(t *testing.T) {
	t.Run("member can leave", func(t *testing.T) {
		var removedUserID uint
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleMember, nil
			},
			removeMemberFn: func(gid, uid uint) error {
				removedUserID = uid
				return nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.LeaveGroup(1, 5)
		assert.NoError(t, err)
		assert.Equal(t, uint(5), removedUserID)
	})

	t.Run("owner cannot leave", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return models.RoleOwner, nil
			},
		}
		svc := NewGroupService(repo)

		err := svc.LeaveGroup(1, 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotOwner)
	})

	t.Run("non-member cannot leave", func(t *testing.T) {
		repo := &mockGroupRepo{
			getMemberRoleFn: func(gid, uid uint) (string, error) {
				return "", gorm.ErrRecordNotFound
			},
		}
		svc := NewGroupService(repo)

		err := svc.LeaveGroup(1, 99)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrNotMember)
	})
}

func TestGroupService_IsMember(t *testing.T) {
	t.Run("returns true when member", func(t *testing.T) {
		repo := &mockGroupRepo{
			isMemberFn: func(gid, uid uint) (bool, error) {
				return true, nil
			},
		}
		svc := NewGroupService(repo)

		isMember, err := svc.IsMember(1, 2)
		assert.NoError(t, err)
		assert.True(t, isMember)
	})

	t.Run("returns false when not member", func(t *testing.T) {
		repo := &mockGroupRepo{
			isMemberFn: func(gid, uid uint) (bool, error) {
				return false, nil
			},
		}
		svc := NewGroupService(repo)

		isMember, err := svc.IsMember(1, 99)
		assert.NoError(t, err)
		assert.False(t, isMember)
	})
}
