package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"chat/models"
)

type GroupRepository interface {
	CreateGroup(group *models.Group) error
	GetGroupByID(groupID uint) (*models.Group, error)
	AddMember(groupID, userID uint, role string) error
	RemoveMember(groupID, userID uint) error
	GetMembers(groupID uint) ([]models.GroupMember, error)
	DeleteGroup(groupID uint) error
	IsMember(groupID, userID uint) (bool, error)
	GetMemberRole(groupID, userID uint) (string, error)
}

type groupRepo struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepo{db: db}
}

func (r *groupRepo) CreateGroup(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepo) GetGroupByID(groupID uint) (*models.Group, error) {
	var group models.Group
	if err := r.db.First(&group, groupID).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *groupRepo) AddMember(groupID, userID uint, role string) error {
	member := models.GroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    role,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"role", "updated_at"}),
	}).Create(&member).Error
}

func (r *groupRepo) GetMembers(groupID uint) ([]models.GroupMember, error) {
	var members []models.GroupMember
	if err := r.db.Where("group_id = ?", groupID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// Soft delete a group by ID
func (r *groupRepo) DeleteGroup(groupID uint) error {
	if err := r.db.Delete(&models.Group{}, groupID).Error; err != nil {
		return err
	}
	return r.db.Where("group_id = ?", groupID).Delete(&models.GroupMember{}).Error
}

func (r *groupRepo) RemoveMember(groupID, userID uint) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func (r *groupRepo) IsMember(groupID, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *groupRepo) GetMemberRole(groupID, userID uint) (string, error) {
	var member models.GroupMember
	if err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&member).Error; err != nil {
		return "", err
	}
	return member.Role, nil
}