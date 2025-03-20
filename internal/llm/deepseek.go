package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const deepseekAPI = "https://api.deepseek.com/v1/chat/completions"

// DeepSeekRequest represents the request structure for DeepSeek API
type DeepSeekRequest struct {
	Model       string              `json:"model"`
	Messages    []DeepSeekMessage   `json:"messages"`
	Temperature float64             `json:"temperature,omitempty"`
	MaxTokens   int                 `json:"max_tokens,omitempty"`
}

// DeepSeekMessage represents a message in the conversation
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekResponse represents the response structure from DeepSeek API
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      DeepSeekMessage `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// GetDeepSeekResponse sends a question to DeepSeek and returns the response
func GetDeepSeekResponse(question string) (string, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "", errors.New("DEEPSEEK_API_KEY environment variable not set")
	}

	requestBody := DeepSeekRequest{
		Model: "deepseek-coder",
		Messages: []DeepSeekMessage{
			{
				Role:    "system",
				Content: "You are McGraph, a helpful coding assistant AI. Provide concise and technical answers to coding questions.",
			},
			{
				Role:    "user",
				Content: question,
			},
		},
		Temperature: 0.7,
		MaxTokens:   800,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", deepseekAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var deepseekResp DeepSeekResponse
	if err := json.Unmarshal(body, &deepseekResp); err != nil {
		return "", err
	}

	if len(deepseekResp.Choices) == 0 {
		return "", errors.New("no response from DeepSeek")
	}

	answer := deepseekResp.Choices[0].Message.Content
	answer = strings.TrimSpace(answer)

	return answer, nil
}