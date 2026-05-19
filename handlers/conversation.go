package handlers

import (
	"ai-chat-backend/db"
	"ai-chat-backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type CreateConversationRequest struct {
	Title        string `json:"title"`
	SystemPrompt string `json:"system_prompt"`
}

// Create new conversation
func CreateConversation(c *gin.Context) {
	userID := c.GetUint("userID")

	var req CreateConversationRequest
	c.ShouldBindJSON(&req)

	// Set defaults if not provided
	if req.Title == "" {
		req.Title = "New Chat"
	}
	if req.SystemPrompt == "" {
		req.SystemPrompt = "You are a helpful assistant."
	}

	conversation := models.Conversation{
		UserID:       userID,
		Title:        req.Title,
		SystemPrompt: req.SystemPrompt,
	}

	db.DB.Create(&conversation)

	c.JSON(http.StatusCreated, gin.H{
		"conversation": conversation,
	})
}

// Update conversation system prompt
func UpdateConversation(c *gin.Context) {
	userID := c.GetUint("userID")
	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Verify ownership + update
	result := db.DB.Model(&models.Conversation{}).
		Where("id = ? AND user_id = ?", convID, userID).
		Updates(map[string]interface{}{
			"title":         req.Title,
			"system_prompt": req.SystemPrompt,
		})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "conversation updated"})
}

// Get all conversations for user
func GetConversations(c *gin.Context) {
	userID := c.GetUint("userID")

	var conversations []models.Conversation
	db.DB.Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&conversations)

	c.JSON(http.StatusOK, gin.H{
		"conversations": conversations,
	})
}

// Get single conversation with messages
func GetConversationMessages(c *gin.Context) {
	userID := c.GetUint("userID")
	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var conversation models.Conversation

	result := db.DB.Where("id = ? AND user_id = ?", convID, userID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at asc")
		}).
		First(&conversation)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation": conversation,
	})
}

// Delete conversation
func DeleteConversation(c *gin.Context) {
	userID := c.GetUint("userID")
	convID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	db.DB.Where("conversation_id = ?", convID).Delete(&models.Message{})
	db.DB.Where("id = ? AND user_id = ?", convID, userID).Delete(&models.Conversation{})

	c.JSON(http.StatusOK, gin.H{"message": "conversation deleted"})
}
func GetUsageHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	// Count AI messages in last hour
	var count int64
	oneHourAgo := time.Now().Add(-time.Hour)

	db.DB.Model(&models.Message{}).
		Where("user_id = ? AND role = ? AND created_at > ?",
			userID, "user", oneHourAgo).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"requests_used":      count,
		"requests_remaining": 20 - count,
		"limit":              20,
		"window":             "1 hour",
	})
}