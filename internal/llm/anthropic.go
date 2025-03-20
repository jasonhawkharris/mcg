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

const anthropicAPI = "https://api.anthropic.com/v1/messages"

// AnthropicRequest represents the request structure for Anthropic API
type AnthropicRequest struct {
	Model     string     `json:"model"`
	MaxTokens int        `json:"max_tokens"`
	System    string     `json:"system"`
	Messages  []Message  `json:"messages"`
}

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicResponse represents the response structure from Anthropic API
type AnthropicResponse struct {
	Content []ContentBlock `json:"content"`
	Error   struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// ContentBlock represents a block of content in the response
type ContentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
}

// GetClaudeResponse sends a question to Anthropic's Claude and returns the response
func GetClaudeResponse(question string) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", errors.New("ANTHROPIC_API_KEY environment variable not set")
	}

	requestBody := AnthropicRequest{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: 800,
		System:    "You are McGraph, a helpful coding assistant AI. Provide concise and technical answers to coding questions.",
		Messages: []Message{
			{
				Role:    "user",
				Content: question,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", anthropicAPI, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", err
	}

	if len(anthropicResp.Content) == 0 {
		return "", errors.New("no response from Claude")
	}

	var textParts []string
	for _, block := range anthropicResp.Content {
		if block.Type == "text" {
			textParts = append(textParts, block.Text)
		}
	}

	answer := strings.Join(textParts, "\n")
	answer = strings.TrimSpace(answer)

	return answer, nil
}