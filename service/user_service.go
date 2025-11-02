package service

import (
	"chat/models"
	"chat/repository"
)

// UserService defines business logic for user management.
type UserService interface {
	CreateUser(user *models.User) error
	GetUserByID(userID uint) (*models.User, error)
	GetAllUsers() ([]models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService constructs a UserService implementation.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(user *models.User) error {
	return s.userRepo.CreateUser(user)
}

func (s *userService) GetUserByID(userID uint) (*models.User, error) {
	return s.userRepo.GetUserByID(userID)
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAllUsers()
}