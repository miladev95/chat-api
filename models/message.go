package models

import (
	"time"

	"gorm.io/gorm"
)

// MessageType enumerates the supported message formats.
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
)

// Message represents a conversation entry.
type Message struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	SenderID   uint           `gorm:"not null" json:"sender_id"`
	ReceiverID *uint          `json:"receiver_id,omitempty"`
	GroupID    *uint          `json:"group_id,omitempty"`
	Type       MessageType    `gorm:"size:50;not null" json:"type"`
	Content    string         `gorm:"type:text" json:"content"`
	FileID     *uint          `json:"file_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	Sender   *User  `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
	Receiver *User  `gorm:"foreignKey:ReceiverID" json:"receiver,omitempty"`
	Group    *Group `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	File     *File  `json:"file,omitempty"`
}

// Seen represents message read receipts.
type Seen struct {
	MessageID uint      `gorm:"primaryKey" json:"message_id"`
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	SeenAt    time.Time `gorm:"autoCreateTime" json:"seen_at"`
}

// File contains metadata for uploaded assets.
type File struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	URL       string         `gorm:"size:512;not null" json:"url"`
	Size      int64          `json:"size"`
	Type      string         `gorm:"size:100" json:"type"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}