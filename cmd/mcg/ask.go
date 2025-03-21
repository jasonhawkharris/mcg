package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hawk/mcgraph/internal/db"
	"github.com/hawk/mcgraph/internal/llm"
	"github.com/hawk/mcgraph/internal/tui"
	"github.com/spf13/cobra"
)

var noSave bool

func init() {
	rootCmd.AddCommand(askCmd)
	
	// Add a flag to run in TUI mode
	askCmd.Flags().BoolP("interactive", "i", false, "Run in interactive chat mode with TUI")
	askCmd.Flags().BoolVarP(&noSave, "no-save", "n", false, "Don't save the conversation")
}

var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Ask McGraph a question",
	Long:  `Ask McGraph a question and he will provide an answer using his LLM capabilities.
If used with the -i/--interactive flag, it will start an interactive chat session.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if interactive mode is requested
		interactive, _ := cmd.Flags().GetBool("interactive")
		
		if interactive {
			// Start the interactive TUI
			ctx := context.Background()
			
			// Create a new conversation
			model := string(llm.GetCurrentLLM())
			conversation, err := dbConn.CreateConversation(ctx, "New Conversation", model)
			if err != nil {
				return fmt.Errorf("error creating conversation: %w", err)
			}
			
			// Set extension manager for the TUI to use
			tui.SetExtensionManager(extManager)
			
			// Create a DB adapter and start the interactive TUI chat
			dbAdapter := db.NewAdapter(dbConn)
			return tui.StartChat(dbAdapter, conversation.ID, nil)
		}
		
		// If no arguments provided in non-interactive mode, show help
		if len(args) == 0 {
			fmt.Println("Please provide a question or use the -i flag for interactive mode")
			cmd.Help()
			return nil
		}
		
		// Standard CLI mode
		question := strings.Join(args, " ")
		
		currentLLM := llm.GetCurrentLLM()
		fmt.Printf("Using %s to answer your question...\n", currentLLM)
		
		answer, err := llm.GetResponse(question)
		if err != nil {
			fmt.Printf("Sorry, I encountered an error: %v\n", err)
			
			// Display the relevant API key that needs to be set
			envVar := llm.GetAPIKeyEnvVar(currentLLM)
			if envVar != "" {
				fmt.Printf("Make sure you have set the %s environment variable.\n", envVar)
			} else {
				fmt.Println("Make sure you have set the required API key environment variable.")
			}
			return nil
		}
		
		// Save the conversation if not disabled
		if !noSave {
			ctx := context.Background()
			
			// Create a new conversation
			model := string(currentLLM)
			conversation, err := dbConn.CreateConversation(ctx, question, model)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to save conversation: %v\n", err)
			} else {
				// Add the messages
				_, err = dbConn.AddMessage(ctx, conversation.ID, "user", question)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to save user message: %v\n", err)
				}
				
				_, err = dbConn.AddMessage(ctx, conversation.ID, "assistant", answer)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to save assistant message: %v\n", err)
				}
				
				// Generate a title
				_, err = dbConn.GenerateTitle(ctx, conversation.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to generate title: %v\n", err)
				}
				
				fmt.Printf("\nConversation saved with ID: %s\n", conversation.ID.String()[:8])
				fmt.Println("Use 'mcg history show " + conversation.ID.String()[:8] + "' to view it later")
			}
		}
		
		fmt.Println(answer)
		return nil
	},
}