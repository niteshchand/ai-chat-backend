package db

import (
	"ai-chat-backend/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	database.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
		&models.Document{},
		&models.DocumentChunk{},
	)

	DB = database
	log.Println("Database connected!")
}