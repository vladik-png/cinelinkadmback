package models

import "time"

type ChatType string

const (
	ChatTypeDirect    ChatType = "direct"
	ChatTypeGroup     ChatType = "group"
	ChatTypeChannel   ChatType = "channel"
	ChatTypeCommunity ChatType = "community"
	ChatTypeCorporate ChatType = "corporate"
)

type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeMedia MessageType = "media"
)

type CorporateChat struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    
	ChatType  ChatType  `gorm:"type:chat_type;default:'direct'"`
	CreatorID uint      
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type CorporateChatMember struct {
	ID                 uint      `gorm:"primaryKey"`
	EmployeeID         uint      
	ChatID             uint      
	MemberIndex        int       
	Role               string    `gorm:"default:'user'"`
	AddedAt            time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	LastSeenMessageID  uint      `gorm:"default:0"`
	CreatedAt          time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type CorporateChatMessage struct {
	ID            uint        `gorm:"primaryKey"`
	MessageID     uint        
	EmployeeID    uint        
	ChatID        uint        
	MessageType   MessageType `gorm:"type:message_type;default:'text'"`
	MessageContent string     `gorm:"column:message"`
	AddedAt       time.Time   `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}
