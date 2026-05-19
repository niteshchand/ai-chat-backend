package models

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-"` // "-" means never send to frontend
	CreatedAt time.Time `json:"created_at"`
}