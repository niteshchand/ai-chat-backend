package handlers

import (
	"ai-chat-backend/db"
	"ai-chat-backend/models"
	"ai-chat-backend/services"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type StreamRequest struct {
	Message        string `json:"message" binding:"required,min=1"`
	ConversationID uint   `json:"conversation_id" binding:"required"`
}

func StreamHandler(c *gin.Context) {
	var req StreamRequest
	userID := c.GetUint("userID")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "message and conversation_id required",
		})
		return
	}

	// Get conversation + verify ownership
	var conversation models.Conversation
	if err := db.DB.Where("id = ? AND user_id = ?", req.ConversationID, userID).
		First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	// Auto update title
	if conversation.Title == "New Chat" {
		title := req.Message
		if len(title) > 40 {
			title = title[:40] + "..."
		}
		db.DB.Model(&conversation).Update("title", title)
	}

	// Save user message
	userMsg := models.Message{
		UserID:         userID,
		ConversationID: req.ConversationID,
		Role:           "user",
		Content:        req.Message,
	}
	db.DB.Create(&userMsg)

	// Load chat history
	var history []models.Message
	db.DB.Where("conversation_id = ?", req.ConversationID).
		Order("created_at asc").
		Limit(20).
		Find(&history)

	// ── RAG: Retrieve relevant chunks ──
	ragContext := ""
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if geminiKey != "" {
		chunks, err := services.RetrieveRelevantChunks(
			geminiKey,
			req.Message,
			req.ConversationID,
			2, // top 3 chunks
		)

if err == nil && len(chunks) > 0 {
    var contextParts []string
    for _, chunk := range chunks {
        // Truncate each chunk to 300 chars
        content := chunk.Content
        if len(content) > 300 {
            content = content[:300]
        }
        contextParts = append(contextParts, content)
    }
    ragContext = strings.Join(contextParts, "\n\n---\n\n")
}
	} 

	// Build system prompt with RAG context
	systemPrompt := conversation.SystemPrompt
	if ragContext != "" {
		systemPrompt = fmt.Sprintf(`%s

You have access to the following document context. Use it to answer the user's question accurately. If the answer is in the context, use it. If not, answer from your general knowledge but mention it.

DOCUMENT CONTEXT:
%s`, systemPrompt, ragContext)
	}

	// Build messages array
	messages := []services.Message{
		{Role: "system", Content: systemPrompt},
	}
	for _, msg := range history {
		messages = append(messages, services.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Stream response
	fullResponse, err := services.StreamGroqWithSave(
		os.Getenv("GROQ_API_KEY"),
		messages,
		c.Writer,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
fmt.Println("Full response length:", len(fullResponse))
fmt.Println("Full response preview:", fullResponse[:min(100, len(fullResponse))])
	// Save AI response
	if fullResponse != "" {
		aiMsg := models.Message{
			UserID:         userID,
			ConversationID: req.ConversationID,
			Role:           "assistant",
			Content:        fullResponse,
		}
		db.DB.Create(&aiMsg)
	}
}