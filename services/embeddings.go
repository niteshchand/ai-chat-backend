package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"fmt"
)

// const EmbeddingModel = "text-embedding-004"
// const EmbeddingDimensions = 768
const EmbeddingModel = "gemini-embedding-001"
const EmbeddingDimensions = 3072

type EmbeddingRequest struct {
	Model   string         `json:"model"`
	Content EmbeddingInput `json:"content"`
}

type EmbeddingInput struct {
	Parts []EmbeddingPart `json:"parts"`
}

type EmbeddingPart struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding struct {
		Values []float32 `json:"values"`
	} `json:"embedding"`
}

// Generate embedding for a single text
func GenerateEmbedding(apiKey string, text string) ([]float32, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models/" +
		EmbeddingModel + ":embedContent?key=" + apiKey

	reqBody := map[string]interface{}{
		"model": "models/" + EmbeddingModel,
		"content": map[string]interface{}{
			"parts": []map[string]string{
				{"text": text},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, errors.New("failed to reach embedding API")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("Gemini response:", string(body))

	var result EmbeddingResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Embedding.Values) == 0 {
		return nil, errors.New("empty embedding returned")
	}

	return result.Embedding.Values, nil
}
// Generate embeddings for multiple chunks
func GenerateEmbeddings(apiKey string, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := GenerateEmbedding(apiKey, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// Convert embedding to JSON string for storage
func EmbeddingToJSON(embedding []float32) (string, error) {
	data, err := json.Marshal(embedding)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Convert JSON string back to embedding
func JSONToEmbedding(jsonStr string) ([]float32, error) {
	var embedding []float32
	err := json.Unmarshal([]byte(jsonStr), &embedding)
	return embedding, err
}

// Cosine similarity in Go — no pgvector needed!
func CosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / float32(math.Sqrt(float64(normA))*math.Sqrt(float64(normB)))
}