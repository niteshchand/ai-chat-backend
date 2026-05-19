package handlers

import (
	"ai-chat-backend/db"
	"ai-chat-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHistoryHandler(c *gin.Context) {
	var messages []models.Message

	// Get last 50 messages, oldest first
	db.DB.Order("created_at asc").Limit(50).Find(&messages)

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}