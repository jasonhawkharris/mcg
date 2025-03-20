package llm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LLMType represents the type of LLM
type LLMType string

const (
	// OpenAI LLM type
	OpenAI LLMType = "openai"
	// Claude LLM type  
	Claude LLMType = "claude"
	// DeepSeek LLM type
	DeepSeek LLMType = "deepseek"
	// Gemini LLM type
	Gemini LLMType = "gemini"
)

var (
	// Current selected LLM
	currentLLM LLMType = OpenAI

	// ErrInvalidLLM is returned when an invalid LLM type is provided
	ErrInvalidLLM = errors.New("invalid LLM type")
)

// SetCurrentLLM sets the current LLM to use
func SetCurrentLLM(llmType string) error {
	llmType = strings.ToLower(llmType)
	
	switch LLMType(llmType) {
	case OpenAI:
		currentLLM = OpenAI
	case Claude:
		currentLLM = Claude
	case DeepSeek:
		currentLLM = DeepSeek
	case Gemini:
		currentLLM = Gemini
	default:
		return fmt.Errorf("%w: %s", ErrInvalidLLM, llmType)
	}

	// Save the selection to config file
	err := saveConfig()
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetCurrentLLM returns the current LLM type
func GetCurrentLLM() LLMType {
	return currentLLM
}

// GetResponse gets a response from the current LLM
func GetResponse(question string) (string, error) {
	switch currentLLM {
	case OpenAI:
		return GetOpenAIResponse(question)
	case Claude:
		return GetClaudeResponse(question)
	case DeepSeek:
		return GetDeepSeekResponse(question)
	case Gemini:
		return GetGeminiResponse(question)
	default:
		return "", fmt.Errorf("%w: %s", ErrInvalidLLM, currentLLM)
	}
}

// GetAvailableLLMs returns a list of available LLM types
func GetAvailableLLMs() []LLMType {
	return []LLMType{OpenAI, Claude, DeepSeek, Gemini}
}

// GetAPIKeyEnvVar returns the environment variable name for the API key for the given LLM
func GetAPIKeyEnvVar(llmType LLMType) string {
	switch llmType {
	case OpenAI:
		return "OPENAI_API_KEY"
	case Claude:
		return "ANTHROPIC_API_KEY"
	case DeepSeek:
		return "DEEPSEEK_API_KEY"
	case Gemini:
		return "GEMINI_API_KEY"
	default:
		return ""
	}
}

// GetConfigDir returns the config directory
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".mcgraph")
	
	// Create the directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err = os.MkdirAll(configDir, 0755)
		if err != nil {
			return "", err
		}
	}

	return configDir, nil
}

// saveConfig saves the current configuration
func saveConfig() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config")
	return os.WriteFile(configFile, []byte(string(currentLLM)), 0644)
}

// LoadConfig loads the configuration
func LoadConfig() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config")
	
	// If the file doesn't exist, use the default (OpenAI)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	llmType := LLMType(strings.TrimSpace(string(data)))
	
	switch llmType {
	case OpenAI, Claude, DeepSeek, Gemini:
		currentLLM = llmType
	default:
		// If the saved value is invalid, fall back to default
		currentLLM = OpenAI
	}

	return nil
}