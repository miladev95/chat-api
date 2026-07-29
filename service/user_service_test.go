package service

import (
	"errors"
	"testing"

	"chat/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestUserService_CreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockUserRepo{
			createUserFn: func(u *models.User) error {
				u.ID = 1
				return nil
			},
		}
		svc := NewUserService(repo)

		user := &models.User{Username: "test", Email: "test@example.com"}
		err := svc.CreateUser(user)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), user.ID)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockUserRepo{
			createUserFn: func(u *models.User) error {
				return errTest
			},
		}
		svc := NewUserService(repo)

		err := svc.CreateUser(&models.User{Username: "test", Email: "test@example.com"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, errTest)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		expected := &models.User{ID: 1, Username: "jane", Email: "jane@example.com"}
		repo := &mockUserRepo{
			getUserByIDFn: func(id uint) (*models.User, error) {
				return expected, nil
			},
		}
		svc := NewUserService(repo)

		user, err := svc.GetUserByID(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, user)
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getUserByIDFn: func(id uint) (*models.User, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}
		svc := NewUserService(repo)

		user, err := svc.GetUserByID(999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		assert.Nil(t, user)
	})
}

func TestUserService_GetAllUsers(t *testing.T) {
	t.Run("returns users", func(t *testing.T) {
		expected := []models.User{
			{ID: 1, Username: "alice", Email: "alice@example.com"},
			{ID: 2, Username: "bob", Email: "bob@example.com"},
		}
		repo := &mockUserRepo{
			getAllUsersFn: func() ([]models.User, error) {
				return expected, nil
			},
		}
		svc := NewUserService(repo)

		users, err := svc.GetAllUsers()
		assert.NoError(t, err)
		assert.Equal(t, expected, users)
	})

	t.Run("returns empty slice", func(t *testing.T) {
		repo := &mockUserRepo{
			getAllUsersFn: func() ([]models.User, error) {
				return []models.User{}, nil
			},
		}
		svc := NewUserService(repo)

		users, err := svc.GetAllUsers()
		assert.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mockUserRepo{
			getAllUsersFn: func() ([]models.User, error) {
				return nil, errors.New("db error")
			},
		}
		svc := NewUserService(repo)

		users, err := svc.GetAllUsers()
		assert.Error(t, err)
		assert.Nil(t, users)
	})
}
