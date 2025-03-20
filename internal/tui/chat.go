package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	tea "github.com/charmbracelet/bubbletea"
)

// DBInterface defines the database operations needed by the TUI
type DBInterface interface {
	AddMessage(ctx context.Context, conversationID uuid.UUID, role, content string) (DBMessage, error)
	GenerateTitle(ctx context.Context, conversationID uuid.UUID) (string, error)
}

// DBMessage is an alias for database.Message to avoid import cycle
type DBMessage = interface{}

// StartChat starts the chat TUI
func StartChat(db DBInterface, conversationID uuid.UUID, loadedMessages []Message) error {
	p := tea.NewProgram(
		NewChatModel(db, conversationID, loadedMessages),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	_, err := p.Run()
	if err != nil {
		fmt.Printf("Error running chat: %v\n", err)
		os.Exit(1)
	}

	return nil
}