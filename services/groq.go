package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"fmt"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type GroqResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// Existing non-streaming function — keep this
func AskGroq(apiKey string, messages []Message) (string, error) {
	body := GroqRequest{
		Model:       "llama-3.3-70b-versatile",
		Temperature: 0.7,
		Messages:    messages,
		Stream:      false,
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("failed to reach Groq API")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result GroqResponse
	json.Unmarshal(respBody, &result)

	if len(result.Choices) == 0 {
		return "", errors.New("no response from Groq")
	}
	return result.Choices[0].Message.Content, nil
}

// NEW — returns full response string after streaming
func StreamGroqWithSave(apiKey string, messages []Message, w http.ResponseWriter) (string, error) {
	body := GroqRequest{
		Model:       "llama-3.3-70b-versatile",
		Temperature: 0.7,
		Messages:    messages,
		Stream:      true,
	}

	jsonBody, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("failed to reach Groq API")
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return "", errors.New("streaming not supported")
	}

	var fullResponse string // collect full response here

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 1024*1024)
scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			w.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		content := chunk.Choices[0].Delta.Content
		if content != "" {
			fullResponse += content // build full response
			 fmt.Print(content) 
			w.Write([]byte("data: " + content + "\n\n"))
			flusher.Flush()
		}
	}

	return fullResponse, nil // return it to handler for saving


	if err := scanner.Err(); err != nil {
    fmt.Println("Scanner error:", err)
}

return fullResponse, nil
}