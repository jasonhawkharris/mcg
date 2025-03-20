package main

import (
	"fmt"
	"strings"

	"github.com/hawk/mcgraph/internal/llm"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pickCmd)
	rootCmd.AddCommand(listLLMsCmd)
}

var pickCmd = &cobra.Command{
	Use:   "pick [llm]",
	Short: "Pick which LLM to use",
	Long:  `Pick which LLM McGraph should use for answering questions.
Available options:
- openai: OpenAI GPT models
- claude: Anthropic Claude models
- deepseek: DeepSeek Coder models
- gemini: Google's Gemini models`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		llmName := strings.ToLower(args[0])
		
		err := llm.SetCurrentLLM(llmName)
		if err != nil {
			fmt.Printf("Error setting LLM: %v\n", err)
			fmt.Println("Available options: openai, claude, deepseek, gemini")
			return
		}
		
		fmt.Printf("Now using %s as the active LLM.\n", llmName)
		
		// Get the appropriate API key environment variable name
		llmType := llm.LLMType(llmName)
		envVar := llm.GetAPIKeyEnvVar(llmType)
		if envVar != "" {
			fmt.Printf("Make sure you have set the %s environment variable.\n", envVar)
		}
	},
}

var listLLMsCmd = &cobra.Command{
	Use:   "llms",
	Short: "List available LLMs",
	Long:  `List all available LLMs that McGraph can use for answering questions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available LLMs:")
		
		for _, llmType := range llm.GetAvailableLLMs() {
			envVar := llm.GetAPIKeyEnvVar(llmType)
			var description string
			
			switch llmType {
			case llm.OpenAI:
				description = "OpenAI GPT models"
			case llm.Claude:
				description = "Anthropic Claude models"
			case llm.DeepSeek:
				description = "DeepSeek Coder models"
			case llm.Gemini:
				description = "Google's Gemini models"
			}
			
			fmt.Printf("- %s: %s (requires %s)\n", llmType, description, envVar)
		}
		
		current := llm.GetCurrentLLM()
		fmt.Printf("\nCurrently using: %s\n", current)
	},
}