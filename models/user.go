package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents an application user in the system.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"size:255;not null;unique" json:"username"`
	Email     string         `gorm:"size:255;not null;unique" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	SentMessages     []Message     `gorm:"foreignKey:SenderID" json:"-"`
	ReceivedMessages []Message     `gorm:"foreignKey:ReceiverID" json:"-"`
	GroupMemberships []GroupMember `gorm:"foreignKey:UserID" json:"-"`
}