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

const geminiAPI = "https://generativelanguage.googleapis.com/v1/models/gemini-1.5-pro:generateContent"

// GeminiRequest represents the request structure for Google's Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
	GenerationConfig GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents content in the request
type GeminiContent struct {
	Role  string           `json:"role,omitempty"`
	Parts []GeminiContentPart `json:"parts"`
}

// GeminiContentPart represents a part of the content
type GeminiContentPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation configuration for Gemini
type GeminiGenerationConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

// GeminiResponse represents the response structure from Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// GetGeminiResponse sends a question to Google's Gemini and returns the response
func GetGeminiResponse(question string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", errors.New("GEMINI_API_KEY environment variable not set")
	}

	systemPrompt := "You are McGraph, a helpful coding assistant AI. Provide concise and technical answers to coding questions."
	
	// Gemini doesn't have a dedicated system message, so we include it in the user message
	fullQuestion := fmt.Sprintf("%s\n\nUser question: %s", systemPrompt, question)
	
	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiContentPart{
					{
						Text: fullQuestion,
					},
				},
			},
		},
		GenerationConfig: GeminiGenerationConfig{
			MaxOutputTokens: 800,
			Temperature:     0.7,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// Add API key as a query parameter
	url := fmt.Sprintf("%s?key=%s", geminiAPI, apiKey)
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

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

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", err
	}

	if geminiResp.Error.Message != "" {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no response from Gemini")
	}

	answer := geminiResp.Candidates[0].Content.Parts[0].Text
	answer = strings.TrimSpace(answer)

	return answer, nil
}