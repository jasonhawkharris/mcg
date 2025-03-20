package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Aliases: []string{"hist"},
	Short:   "Manage conversation history",
	Long:    `View, continue, or delete conversation history.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior is to list conversations
		listConversations()
	},
}

var historyShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show a specific conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		partialID := args[0]
		
		// Check if this is a full UUID or a shortened one
		if len(partialID) == 36 {
			id, err := uuid.Parse(partialID)
			if err != nil {
				return fmt.Errorf("invalid conversation ID: %w", err)
			}
			return showConversation(id)
		} else {
			// Try to find conversation with the partial ID
			ctx := context.Background()
			conversations, err := dbConn.ListConversations(ctx)
			if err != nil {
				return fmt.Errorf("error listing conversations: %w", err)
			}
			
			for _, conv := range conversations {
				if strings.HasPrefix(conv.ID.String(), partialID) {
					return showConversation(conv.ID)
				}
			}
			
			return fmt.Errorf("no conversation found with ID starting with: %s", partialID)
		}
	},
}

var historyDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a conversation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		partialID := args[0]
		
		// Check if this is a full UUID or a shortened one
		if len(partialID) == 36 {
			id, err := uuid.Parse(partialID)
			if err != nil {
				return fmt.Errorf("invalid conversation ID: %w", err)
			}
			return deleteConversation(id)
		} else {
			// Try to find conversation with the partial ID
			ctx := context.Background()
			conversations, err := dbConn.ListConversations(ctx)
			if err != nil {
				return fmt.Errorf("error listing conversations: %w", err)
			}
			
			for _, conv := range conversations {
				if strings.HasPrefix(conv.ID.String(), partialID) {
					return deleteConversation(conv.ID)
				}
			}
			
			return fmt.Errorf("no conversation found with ID starting with: %s", partialID)
		}
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyShowCmd)
	historyCmd.AddCommand(historyDeleteCmd)
}

// listConversations displays all saved conversations
func listConversations() {
	ctx := context.Background()
	conversations, err := dbConn.ListConversations(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing conversations: %v\n", err)
		return
	}

	if len(conversations) == 0 {
		fmt.Println("No saved conversations found.")
		return
	}

	// Create a tabwriter for nicely formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTitle\tModel\tLast Updated")
	fmt.Fprintln(w, "--\t-----\t-----\t------------")

	for _, conv := range conversations {
		// Format the ID to be shorter
		shortID := conv.ID.String()[:8]
		// Format the time
		timeAgo := timeAgo(time.Since(conv.UpdatedAt))
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", shortID, conv.Title, conv.Model, timeAgo)
	}

	w.Flush()
	fmt.Println("\nUse 'mcg history show <id>' to view a conversation")
	fmt.Println("Use 'mcg chat --continue <id>' to continue a conversation")
}

// showConversation displays a specific conversation
func showConversation(id uuid.UUID) error {
	ctx := context.Background()
	conversation, err := dbConn.GetConversation(ctx, id)
	if err != nil {
		return fmt.Errorf("error retrieving conversation: %w", err)
	}

	fmt.Printf("Title: %s\n", conversation.Title)
	fmt.Printf("Model: %s\n", conversation.Model)
	fmt.Printf("Created: %s\n", conversation.CreatedAt.Format(time.RFC1123))
	fmt.Printf("Updated: %s\n\n", conversation.UpdatedAt.Format(time.RFC1123))

	for _, msg := range conversation.Messages {
		// Format role in uppercase for clarity
		role := msg.Role
		if role == "user" {
			fmt.Printf("USER: %s\n\n", msg.Content)
		} else if role == "assistant" {
			fmt.Printf("ASSISTANT: %s\n\n", msg.Content)
		} else {
			fmt.Printf("%s: %s\n\n", role, msg.Content)
		}
	}

	return nil
}

// deleteConversation deletes a specific conversation
func deleteConversation(id uuid.UUID) error {
	ctx := context.Background()
	
	// First get the conversation to show what we're deleting
	conversation, err := dbConn.GetConversation(ctx, id)
	if err != nil {
		return fmt.Errorf("error retrieving conversation: %w", err)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete conversation: \"%s\" [y/N]? ", conversation.Title)
	var confirm string
	fmt.Scanln(&confirm)
	
	if confirm != "y" && confirm != "Y" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	// Delete the conversation
	err = dbConn.DeleteConversation(ctx, id)
	if err != nil {
		return fmt.Errorf("error deleting conversation: %w", err)
	}

	fmt.Println("Conversation deleted successfully.")
	return nil
}

// timeAgo returns a human-readable string representing how long ago a time was
func timeAgo(duration time.Duration) string {
	seconds := int(duration.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24

	if days > 0 {
		return fmt.Sprintf("%dd ago", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh ago", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm ago", minutes)
	} else {
		return fmt.Sprintf("%ds ago", seconds)
	}
}