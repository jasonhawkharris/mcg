package main

import (
	"fmt"
	"os"

	"github.com/hawk/mcgraph/internal/llm"
)

func main() {
	// Load the LLM configuration
	err := llm.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
	}

	Execute()
}