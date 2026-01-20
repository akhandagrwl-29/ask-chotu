package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Model       string    `json:"model,omitempty"`
	TopP        int       `json:"top_p,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type ChatHistory struct {
	Messages []Message
}

// AddMessage adds a message to history
func (h *ChatHistory) AddMessage(role, content string) {
	h.Messages = append(h.Messages, Message{
		Role:    role,
		Content: content,
	})
}

// GetChatbotResponse sends a request to GitHub Models API
func getChatbotResponse(input string, context string, history *ChatHistory) (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	curr := time.Now()
	fmt.Printf("Request to chatbot at time: %s\n", curr)

	// Build messages array
	messages := []Message{}

	// Add system context if provided
	if context != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: context,
		})
	}

	// Add conversation history
	if history != nil {
		messages = append(messages, history.Messages...)
	}

	// Add current user input
	messages = append(messages, Message{
		Role:    "user",
		Content: input,
	})

	// Prepare request
	reqBody := ChatRequest{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
		Model:       "openai/gpt-4o-mini",
		TopP:        1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	url := "https://models.github.ai/inference/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+githubToken)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	response := chatResp.Choices[0].Message.Content

	// Add to history
	if history != nil {
		history.AddMessage("user", input)
		history.AddMessage("assistant", response)
	}

	fmt.Printf("Chatbot response time: %v\n", time.Since(curr))
	return response, nil
}
