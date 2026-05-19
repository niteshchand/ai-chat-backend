package services

import (
	"strings"
	"unicode"
)

const (
	ChunkSize    = 500  // characters per chunk
	ChunkOverlap = 100  // overlap between chunks
)

// Split text into overlapping chunks
func ChunkText(text string) []string {
	// Clean the text first
	text = cleanText(text)

	if len(text) == 0 {
		return []string{}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + ChunkSize

		// Don't go past end of text
		if end > len(text) {
			end = len(text)
		} else {
			// Find nearest sentence boundary
			end = findSentenceBoundary(text, end)
		}

		chunk := strings.TrimSpace(text[start:end])
		if len(chunk) > 0 {
			chunks = append(chunks, chunk)
		}

		// Move forward with overlap
		start = end - ChunkOverlap
		if start < 0 {
			start = 0
		}

		// Prevent infinite loop
		if end == len(text) {
			break
		}
	}

	return chunks
}

// Find nearest sentence end (. ! ?) near position
func findSentenceBoundary(text string, pos int) int {
	// Look forward up to 100 chars for sentence end
	searchEnd := pos + 100
	if searchEnd > len(text) {
		searchEnd = len(text)
	}

	for i := pos; i < searchEnd; i++ {
		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			return i + 1
		}
	}

	// Look backward for sentence end
	for i := pos; i > pos-100 && i >= 0; i-- {
		if text[i] == '.' || text[i] == '!' || text[i] == '?' {
			return i + 1
		}
	}

	return pos
}

// Clean extracted PDF text
func cleanText(text string) string {
	// Remove excessive whitespace
	var builder strings.Builder
	prevSpace := false

	for _, r := range text {
		if unicode.IsSpace(r) {
			if !prevSpace {
				builder.WriteRune(' ')
			}
			prevSpace = true
		} else {
			builder.WriteRune(r)
			prevSpace = false
		}
	}

	return strings.TrimSpace(builder.String())
}