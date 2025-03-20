package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"conversations"},
	Short:   "List all conversations",
	Long:    `List all saved conversations.`,
	Run: func(cmd *cobra.Command, args []string) {
		displayConversationList()
	},
}

// displayConversationList displays all saved conversations (separate from history.go's version)
func displayConversationList() {
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
		timeAgo := formatTimeAgo(time.Since(conv.UpdatedAt))
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", shortID, conv.Title, conv.Model, timeAgo)
	}

	w.Flush()
	fmt.Println("\nUse 'mcg history show <id>' to view a conversation")
	fmt.Println("Use 'mcg chat --continue <id>' to continue a conversation")
}

// formatTimeAgo returns a human-readable string representing how long ago a time was
func formatTimeAgo(duration time.Duration) string {
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