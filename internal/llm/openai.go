package llm

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// GetOpenAIResponse sends a question to OpenAI and returns the response
func GetOpenAIResponse(question string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", errors.New("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are McGraph, a helpful coding assistant AI. Provide concise and technical answers to coding questions.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: question,
				},
			},
			MaxTokens: 800,
		},
	)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	// Clean up the response a bit
	answer := resp.Choices[0].Message.Content
	answer = strings.TrimSpace(answer)

	return answer, nil
}