package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// Group represents a chat group or channel.
type Group struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:255;not null;unique" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Members  []GroupMember `json:"members"`
	Messages []Message     `gorm:"foreignKey:GroupID" json:"-"`
}

// GroupMember represents a user's membership within a group.
type GroupMember struct {
	GroupID uint   `gorm:"primaryKey" json:"group_id"`
	UserID  uint   `gorm:"primaryKey" json:"user_id"`
	Role    string `gorm:"size:50;not null;default:'member'" json:"role"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}