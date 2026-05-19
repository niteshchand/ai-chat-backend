package models

import "time"

type Document struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id"`
	ConversationID uint      `json:"conversation_id"`
	Filename       string    `json:"filename"`
	ChunkCount     int       `json:"chunk_count"`
	CreatedAt      time.Time `json:"created_at"`
}

type DocumentChunk struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DocumentID uint      `json:"document_id"`
	UserID     uint      `json:"user_id"`
	Content    string    `json:"content"`
	ChunkIndex int       `json:"chunk_index"`
	Embedding  string    `json:"-" gorm:"type:text"`
	CreatedAt  time.Time `json:"created_at"`
}