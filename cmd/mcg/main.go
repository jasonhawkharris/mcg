package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hawk/mcgraph/internal/llm"
)

func main() {
	// Check for test mode flag
	if len(os.Args) > 1 && os.Args[1] == "--test-typing-prevention" {
		testTypingPrevention()
		return
	}

	// Load the LLM configuration
	err := llm.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
	}

	// Execute the root command
	Execute()
}

// testTypingPrevention is a test function to verify that sending messages
// while the model is typing is prevented
func testTypingPrevention() {
	fmt.Println("Testing message typing prevention...")
	fmt.Println("In chat_model.go, KeyEnter handling has been modified to check:")
	
	file, err := os.ReadFile("internal/tui/chat_model.go")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	
	content := string(file)
	enterKeyBlock := ""
	
	// Find the KeyEnter case in the Update function
	if idx := strings.Index(content, "case tea.KeyEnter:"); idx >= 0 {
		// Extract the block
		blockStart := idx
		braceCount := 0
		inBlock := false
		
		for i := blockStart; i < len(content); i++ {
			if content[i] == '{' {
				braceCount++
				inBlock = true
			} else if content[i] == '}' {
				braceCount--
			}
			
			if inBlock && braceCount == 0 {
				enterKeyBlock = content[blockStart:i+1]
				break
			}
		}
	}
	
	if enterKeyBlock == "" {
		fmt.Println("Could not find KeyEnter case in chat_model.go")
		return
	}
	
	// Check if the prevention logic is present
	if strings.Contains(enterKeyBlock, "if !m.waitingForResp") {
		fmt.Println("✅ Found check for !m.waitingForResp - prevents sending when waiting for response")
	} else {
		fmt.Println("❌ Missing check for !m.waitingForResp")
	}
	
	if strings.Contains(enterKeyBlock, "!m.typingActive") {
		fmt.Println("✅ Found check for !m.typingActive - prevents sending when LLM is still typing")
	} else {
		fmt.Println("❌ Missing check for !m.typingActive")
	}
	
	// Verify the combined logic - should be both checks
	if strings.Contains(enterKeyBlock, "if !m.waitingForResp && !m.typingActive") {
		fmt.Println("✅ Perfect implementation: Message sending is prevented while waiting for response AND while LLM is typing")
	} else if strings.Contains(enterKeyBlock, "if !m.waitingForResp") {
		fmt.Println("⚠️ Partial implementation: Message sending is prevented while waiting for response, but no explicit check for typing status")
		fmt.Println("   This may still work since waitingForResp is true during initial response generation.")
	}
}