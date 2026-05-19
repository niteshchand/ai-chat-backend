package handlers

import (
	"ai-chat-backend/services"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type SearchRequest struct {
	Query          string `json:"query" binding:"required"`
	ConversationID uint   `json:"conversation_id" binding:"required"`
}

func SearchHandler(c *gin.Context) {
	var req SearchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query and conversation_id required",
		})
		return
	}

	chunks, err := services.RetrieveRelevantChunks(
		os.Getenv("GEMINI_API_KEY"),
		req.Query,
		req.ConversationID,
		2,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var results []map[string]interface{}
	for _, chunk := range chunks {
		results = append(results, map[string]interface{}{
			"chunk_index": chunk.ChunkIndex,
			"content":     chunk.Content,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   req.Query,
		"results": results,
	})
}