package repository

import (
	"testing"

	"chat/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("successfully creates a user", func(t *testing.T) {
		user := &models.User{
			Username: "john_doe",
			Email:    "john@example.com",
		}

		err := repo.CreateUser(user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "john_doe", user.Username)
		assert.Equal(t, "john@example.com", user.Email)
	})

	t.Run("fails with duplicate username", func(t *testing.T) {
		repo.CreateUser(&models.User{Username: "duplicate_user", Email: "first@example.com"})

		duplicate := &models.User{Username: "duplicate_user", Email: "second@example.com"}
		err := repo.CreateUser(duplicate)
		assert.Error(t, err)
	})

	t.Run("fails with duplicate email", func(t *testing.T) {
		repo.CreateUser(&models.User{Username: "user_a", Email: "same@example.com"})

		duplicate := &models.User{Username: "user_b", Email: "same@example.com"}
		err := repo.CreateUser(duplicate)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("returns user when exists", func(t *testing.T) {
		user := &models.User{Username: "jane_doe", Email: "jane@example.com"}
		err := repo.CreateUser(user)
		assert.NoError(t, err)

		found, err := repo.GetUserByID(user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, "jane_doe", found.Username)
		assert.Equal(t, "jane@example.com", found.Email)
	})

	t.Run("returns error when user does not exist", func(t *testing.T) {
		found, err := repo.GetUserByID(9999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Nil(t, found)
	})

	t.Run("returns error for zero ID", func(t *testing.T) {
		found, err := repo.GetUserByID(0)
		assert.Error(t, err)
		assert.Nil(t, found)
	})
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	db := setupTestDB()
	repo := NewUserRepository(db)

	t.Run("returns empty slice when no users exist", func(t *testing.T) {
		users, err := repo.GetAllUsers()
		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("returns all created users", func(t *testing.T) {
		user1 := &models.User{Username: "alice", Email: "alice@example.com"}
		user2 := &models.User{Username: "bob", Email: "bob@example.com"}
		user3 := &models.User{Username: "carol", Email: "carol@example.com"}

		assert.NoError(t, repo.CreateUser(user1))
		assert.NoError(t, repo.CreateUser(user2))
		assert.NoError(t, repo.CreateUser(user3))

		users, err := repo.GetAllUsers()
		assert.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("does not return soft-deleted users", func(t *testing.T) {
		user := &models.User{Username: "deletable", Email: "deletable@example.com"}
		assert.NoError(t, repo.CreateUser(user))

		// Soft delete via GORM
		db.Delete(&models.User{}, user.ID)

		users, err := repo.GetAllUsers()
		assert.NoError(t, err)

		for _, u := range users {
			assert.NotEqual(t, user.ID, u.ID, "soft-deleted user should not appear")
		}
	})
}
