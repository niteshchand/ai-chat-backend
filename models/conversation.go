package models

import "time"

type Conversation struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	SystemPrompt string    `json:"system_prompt" gorm:"default:'You are a helpful assistant.'"`
	CreatedAt time.Time `json:"created_at"`
	Messages  []Message `json:"messages,omitempty" gorm:"foreignKey:ConversationID"`
}