package models

import "time"

type Message struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	ConversationID uint `json:"conversation_id"`
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}