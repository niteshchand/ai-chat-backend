package services

import (
	"ai-chat-backend/db"
	"ai-chat-backend/models"
	"sort"
	"fmt"
)

type ScoredChunk struct {
	Chunk      models.DocumentChunk
	Similarity float32
}

// Find top K most relevant chunks for a query
func RetrieveRelevantChunks(
	apiKey string,
	query string,
	conversationID uint,
	topK int,
) ([]models.DocumentChunk, error) {

	queryEmbedding, err := GenerateEmbedding(apiKey, query)
	if err != nil {
		return nil, err
	}
	fmt.Println("Query embedding length:", len(queryEmbedding))

	var chunks []models.DocumentChunk
	db.DB.Where("document_id IN (?)",
		db.DB.Model(&models.Document{}).
			Select("id").
			Where("conversation_id = ?", conversationID),
	).Find(&chunks)

	fmt.Println("Chunks found:", len(chunks))

	if len(chunks) == 0 {
		return []models.DocumentChunk{}, nil
	}

	var scored []ScoredChunk
	for _, chunk := range chunks {
		fmt.Println("Chunk embedding length:", len(chunk.Embedding))
		if chunk.Embedding == "" {
			continue
		}

		chunkEmbedding, err := JSONToEmbedding(chunk.Embedding)
		if err != nil {
			continue
		}

		similarity := CosineSimilarity(queryEmbedding, chunkEmbedding)
		fmt.Println("Similarity score:", similarity)

		if similarity > 0.5 {
			scored = append(scored, ScoredChunk{
				Chunk:      chunk,
				Similarity: similarity,
			})
		}
	}
    // ... rest of code

	// Sort by similarity score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Similarity > scored[j].Similarity
	})

	// Return top K chunks
	if topK > len(scored) {
		topK = len(scored)
	}

	result := make([]models.DocumentChunk, topK)
	for i := 0; i < topK; i++ {
		result[i] = scored[i].Chunk
	}

	return result, nil
}