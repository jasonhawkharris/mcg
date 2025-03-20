package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hawk/mcgraph/internal/db"
	"github.com/hawk/mcgraph/internal/llm"
	"github.com/hawk/mcgraph/internal/tui"
	"github.com/spf13/cobra"
)

var continueID string

func init() {
	chatCmd.Flags().StringVarP(&continueID, "continue", "c", "", "Continue a previous conversation by ID")
	rootCmd.AddCommand(chatCmd)
}

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with McGraph",
	Long:  `Start an interactive TUI chat session with McGraph.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		
		// If continueID is provided, load the conversation
		var loadedMessages []tui.Message
		var conversationID uuid.UUID
		var err error
		
		if continueID != "" {
			// Check if this is a full UUID or a shortened one
			if len(continueID) == 36 {
				// This is a full UUID
				conversationID, err = uuid.Parse(continueID)
				if err != nil {
					return fmt.Errorf("invalid conversation ID: %w", err)
				}
			} else {
				// Try to find conversation with the partial ID
				conversations, err := dbConn.ListConversations(ctx)
				if err != nil {
					return fmt.Errorf("error listing conversations: %w", err)
				}
				
				found := false
				for _, conv := range conversations {
					if strings.HasPrefix(conv.ID.String(), continueID) {
						conversationID = conv.ID
						found = true
						break
					}
				}
				
				if !found {
					return fmt.Errorf("no conversation found with ID starting with: %s", continueID)
				}
			}
			
			// Load the conversation
			conversation, err := dbConn.GetConversation(ctx, conversationID)
			if err != nil {
				return fmt.Errorf("error loading conversation: %w", err)
			}
			
			// Set the current LLM to the one used in the conversation
			err = llm.SetCurrentLLM(string(conversation.Model))
			if err != nil {
				// If the LLM is not available, continue with the current one but warn the user
				fmt.Fprintf(os.Stderr, "Warning: This conversation used %s but it's not available. Using %s instead.\n",
					conversation.Model, llm.GetCurrentLLM())
			}
			
			// Add welcome message first
			currentLLM := llm.GetCurrentLLM()
			welcomeMsg := fmt.Sprintf("Welcome back to McGraph Chat! Current LLM: %s\nContinuing conversation: %s\nType your questions and press Enter to submit. Type Ctrl+C to quit.", 
				currentLLM, conversation.Title)
			
			loadedMessages = append(loadedMessages, tui.Message{
				Content:       welcomeMsg,
				VisibleContent: welcomeMsg,
				IsUser:        false,
				IsComplete:    true,
				Time:          time.Now(),
			})
			
			// Then add all the conversation messages
			for _, msg := range conversation.Messages {
				loadedMessages = append(loadedMessages, tui.Message{
					Content:       msg.Content,
					VisibleContent: msg.Content,
					IsUser:        msg.Role == "user",
					IsComplete:    true,
					Time:          msg.CreatedAt, // Use the original timestamp
				})
			}
			
			fmt.Printf("Continuing conversation: %s\n", conversation.Title)
		} else {
			// Create a new conversation
			model := string(llm.GetCurrentLLM())
			conversation, err := dbConn.CreateConversation(ctx, "New Conversation", model)
			if err != nil {
				return fmt.Errorf("error creating conversation: %w", err)
			}
			conversationID = conversation.ID
		}
		
		// Create a DB adapter and start the interactive TUI chat
		dbAdapter := db.NewAdapter(dbConn)
		return tui.StartChat(dbAdapter, conversationID, loadedMessages)
	},
}