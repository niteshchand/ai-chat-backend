package handlers

import (
	"ai-chat-backend/db"
	"ai-chat-backend/models"
	"ai-chat-backend/services"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Message string `json:"message" binding:"required,min=1"`
}

func ChatHandler(c *gin.Context) {
	var req ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message field is required",
		})
		return
	}

	// Save user message to DB
	userMsg := models.Message{
		Role:    "user",
		Content: req.Message,
	}
	db.DB.Create(&userMsg)

	// Build messages for AI
	messages := []services.Message{
		{Role: "system", Content: "You are a helpful Go programming assistant."},
		{Role: "user", Content: req.Message},
	}

	// Get AI response
	response, err := services.AskGroq(os.Getenv("GROQ_API_KEY"), messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Save AI response to DB
	aiMsg := models.Message{
		Role:    "assistant",
		Content: response,
	}
	db.DB.Create(&aiMsg)

	c.JSON(http.StatusOK, gin.H{
		"reply": response,
	})
}