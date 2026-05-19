package handlers

import (
	"ai-chat-backend/db"
	"ai-chat-backend/models"
	"ai-chat-backend/services"
	"fmt"
	"net/http"
	"os"
	"strconv"

	pdf "github.com/dslipak/pdf"
	"github.com/gin-gonic/gin"
)

const MaxFileSize = 10 << 20

func UploadHandler(c *gin.Context) {
	userID := c.GetUint("userID")

	convIDStr := c.PostForm("conversation_id")
	convID, err := strconv.Atoi(convIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id required"})
		return
	}

	var conversation models.Conversation
	if err := db.DB.Where("id = ? AND user_id = ?", convID, userID).
		First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxFileSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	if header.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only PDF files are supported"})
		return
	}

	reader, err := pdf.NewReader(file, header.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read PDF"})
		return
	}

	var fullText string
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		fullText += text + " "
	}

	if len(fullText) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not extract text from PDF"})
		return
	}

	chunks := services.ChunkText(fullText)
	if len(chunks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no content found in PDF"})
		return
	}

	// ── Generate embeddings ──
	fmt.Println("Generating embeddings for", len(chunks), "chunks...")
	embeddings, err := services.GenerateEmbeddings(
		os.Getenv("GEMINI_API_KEY"),
		chunks,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate embeddings: " + err.Error(),
		})
		return
	}
	fmt.Println("Embeddings generated successfully!")

	// Save document
	doc := models.Document{
		UserID:         userID,
		ConversationID: uint(convID),
		Filename:       header.Filename,
		ChunkCount:     len(chunks),
	}
	db.DB.Create(&doc)

	// Save chunks WITH embeddings
	for i, chunk := range chunks {
		embJSON, err := services.EmbeddingToJSON(embeddings[i])
		if err != nil {
			fmt.Println("Failed to convert embedding:", err)
			continue
		}
		docChunk := models.DocumentChunk{
			DocumentID: doc.ID,
			UserID:     userID,
			Content:    chunk,
			ChunkIndex: i,
			Embedding:  embJSON, // ← saving embedding!
		}
		db.DB.Create(&docChunk)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "PDF processed with embeddings",
		"document_id": doc.ID,
		"filename":    header.Filename,
		"chunks":      len(chunks),
		"preview":     fmt.Sprintf("%.200s...", fullText),
	})
}