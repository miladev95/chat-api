package repository

import (
	"testing"

	"chat/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGroupRepository_CreateGroup(t *testing.T) {
	db := setupTestDB()
	repo := NewGroupRepository(db)

	t.Run("successfully creates a group", func(t *testing.T) {
		group := &models.Group{Name: "Project Team"}

		err := repo.CreateGroup(group)
		assert.NoError(t, err)
		assert.NotZero(t, group.ID)
		assert.Equal(t, "Project Team", group.Name)
	})

	t.Run("fails with duplicate group name", func(t *testing.T) {
		repo.CreateGroup(&models.Group{Name: "Unique Group"})

		duplicate := &models.Group{Name: "Unique Group"}
		err := repo.CreateGroup(duplicate)
		assert.Error(t, err)
	})
}

func TestGroupRepository_GetGroupByID(t *testing.T) {
	db := setupTestDB()
	repo := NewGroupRepository(db)

	t.Run("returns group when exists", func(t *testing.T) {
		group := &models.Group{Name: "Design Team"}
		assert.NoError(t, repo.CreateGroup(group))

		found, err := repo.GetGroupByID(group.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, group.ID, found.ID)
		assert.Equal(t, "Design Team", found.Name)
	})

	t.Run("returns error when group does not exist", func(t *testing.T) {
		found, err := repo.GetGroupByID(9999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Nil(t, found)
	})
}

func TestGroupRepository_AddMember(t *testing.T) {
	db := setupTestDB()
	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	// Shared fixtures
	group := &models.Group{Name: "Test Group"}
	user := &models.User{Username: "member_user", Email: "member@example.com"}
	assert.NoError(t, groupRepo.CreateGroup(group))
	assert.NoError(t, userRepo.CreateUser(user))

	t.Run("adds a member with default role", func(t *testing.T) {
		err := groupRepo.AddMember(group.ID, user.ID, models.RoleMember)
		assert.NoError(t, err)
	})

	t.Run("adds a member with owner role", func(t *testing.T) {
		owner := &models.User{Username: "owner_user", Email: "owner@example.com"}
		assert.NoError(t, userRepo.CreateUser(owner))

		err := groupRepo.AddMember(group.ID, owner.ID, models.RoleOwner)
		assert.NoError(t, err)
	})

	t.Run("upserts member on duplicate (role update)", func(t *testing.T) {
		// Add as member first
		memberUser := &models.User{Username: "promoted_user", Email: "promoted@example.com"}
		assert.NoError(t, userRepo.CreateUser(memberUser))
		assert.NoError(t, groupRepo.AddMember(group.ID, memberUser.ID, models.RoleMember))

		// Now add again with owner role — should upsert (not error)
		err := groupRepo.AddMember(group.ID, memberUser.ID, models.RoleOwner)
		assert.NoError(t, err)

		// Verify role was updated
		members, err := groupRepo.GetMembers(group.ID)
		assert.NoError(t, err)

		var found bool
		for _, m := range members {
			if m.UserID == memberUser.ID {
				assert.Equal(t, models.RoleOwner, m.Role, "role should be updated to owner")
				found = true
				break
			}
		}
		assert.True(t, found, "user should still be a member after upsert")
	})

	t.Run("adds member to non-existent group succeeds silently", func(t *testing.T) {
		// GORM with SQLite does not enforce foreign keys by default,
		// so AddMember on a deleted/non-existent group will succeed.
		newUser := &models.User{Username: "orphan_user", Email: "orphan@example.com"}
		assert.NoError(t, userRepo.CreateUser(newUser))

		err := groupRepo.AddMember(9999, newUser.ID, models.RoleMember)
		assert.NoError(t, err, "GORM does not enforce FK constraints by default in SQLite")
	})
}

func TestGroupRepository_GetMembers(t *testing.T) {
	db := setupTestDB()
	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	group := &models.Group{Name: "Team Alpha"}
	assert.NoError(t, groupRepo.CreateGroup(group))

	t.Run("returns empty slice when no members", func(t *testing.T) {
		members, err := groupRepo.GetMembers(group.ID)
		assert.NoError(t, err)
		assert.Empty(t, members)
	})

	t.Run("returns all members after adding them", func(t *testing.T) {
		users := []*models.User{
			{Username: "alpha", Email: "alpha@example.com"},
			{Username: "beta", Email: "beta@example.com"},
			{Username: "gamma", Email: "gamma@example.com"},
		}
		for _, u := range users {
			assert.NoError(t, userRepo.CreateUser(u))
			assert.NoError(t, groupRepo.AddMember(group.ID, u.ID, models.RoleMember))
		}

		members, err := groupRepo.GetMembers(group.ID)
		assert.NoError(t, err)
		assert.Len(t, members, 3)
	})
}

func TestGroupRepository_RemoveMember(t *testing.T) {
	db := setupTestDB()
	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	group := &models.Group{Name: "Team Beta"}
	user := &models.User{Username: "removable", Email: "removable@example.com"}
	assert.NoError(t, groupRepo.CreateGroup(group))
	assert.NoError(t, userRepo.CreateUser(user))
	assert.NoError(t, groupRepo.AddMember(group.ID, user.ID, models.RoleMember))

	t.Run("removes an existing member", func(t *testing.T) {
		err := groupRepo.RemoveMember(group.ID, user.ID)
		assert.NoError(t, err)

		members, err := groupRepo.GetMembers(group.ID)
		assert.NoError(t, err)
		assert.Empty(t, members)
	})

	t.Run("removing non-existent member does not error", func(t *testing.T) {
		err := groupRepo.RemoveMember(group.ID, 9999)
		assert.NoError(t, err)
	})
}

func TestGroupRepository_GetAllGroups(t *testing.T) {
	db := setupTestDB()
	repo := NewGroupRepository(db)

	t.Run("returns empty slice when no groups exist", func(t *testing.T) {
		groups, err := repo.GetAllGroups()
		assert.NoError(t, err)
		assert.Empty(t, groups)
	})

	t.Run("returns all created groups", func(t *testing.T) {
		assert.NoError(t, repo.CreateGroup(&models.Group{Name: "Alpha"}))
		assert.NoError(t, repo.CreateGroup(&models.Group{Name: "Beta"}))

		groups, err := repo.GetAllGroups()
		assert.NoError(t, err)
		assert.Len(t, groups, 2)
	})
}

func TestGroupRepository_UpdateGroup(t *testing.T) {
	db := setupTestDB()
	repo := NewGroupRepository(db)

	group := &models.Group{Name: "Original Name"}
	assert.NoError(t, repo.CreateGroup(group))

	t.Run("updates group name", func(t *testing.T) {
		err := repo.UpdateGroup(group.ID, "Updated Name")
		assert.NoError(t, err)

		updated, err := repo.GetGroupByID(group.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
	})

	t.Run("updating non-existent group does not error", func(t *testing.T) {
		err := repo.UpdateGroup(9999, "Ghost")
		assert.NoError(t, err)
	})
}

func TestGroupRepository_UpdateMemberRole(t *testing.T) {
	db := setupTestDB()
	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	group := &models.Group{Name: "Role Test Group"}
	user := &models.User{Username: "role_user", Email: "role@example.com"}
	assert.NoError(t, groupRepo.CreateGroup(group))
	assert.NoError(t, userRepo.CreateUser(user))
	assert.NoError(t, groupRepo.AddMember(group.ID, user.ID, models.RoleMember))

	t.Run("updates member role", func(t *testing.T) {
		err := groupRepo.UpdateMemberRole(group.ID, user.ID, models.RoleAdmin)
		assert.NoError(t, err)

		role, err := groupRepo.GetMemberRole(group.ID, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, models.RoleAdmin, role)
	})

	t.Run("updating non-existent member does not error", func(t *testing.T) {
		err := groupRepo.UpdateMemberRole(group.ID, 9999, models.RoleAdmin)
		assert.NoError(t, err)
	})
}

func TestGroupRepository_DeleteGroup(t *testing.T) {
	db := setupTestDB()
	groupRepo := NewGroupRepository(db)
	userRepo := NewUserRepository(db)

	t.Run("soft deletes a group and its members", func(t *testing.T) {
		group := &models.Group{Name: "Deletable Group"}
		assert.NoError(t, groupRepo.CreateGroup(group))

		user := &models.User{Username: "temp_member", Email: "temp@example.com"}
		assert.NoError(t, userRepo.CreateUser(user))
		assert.NoError(t, groupRepo.AddMember(group.ID, user.ID, models.RoleMember))

		// Delete the group
		err := groupRepo.DeleteGroup(group.ID)
		assert.NoError(t, err)

		// Verify group is soft-deleted
		_, err = groupRepo.GetGroupByID(group.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		// Verify members are also deleted
		members, err := groupRepo.GetMembers(group.ID)
		assert.NoError(t, err)
		assert.Empty(t, members)
	})

	t.Run("deleting non-existent group does not error", func(t *testing.T) {
		err := groupRepo.DeleteGroup(9999)
		assert.NoError(t, err) // GORM Delete on 0 rows is not an error
	})
}
