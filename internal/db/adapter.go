package db

import (
	"context"

	"github.com/google/uuid"
)

// DBAdapter adapts the DB type to implement any interface requiring the database methods
type DBAdapter struct {
	DB *DB
}

// NewAdapter creates a new adapter for the DB
func NewAdapter(db *DB) *DBAdapter {
	return &DBAdapter{DB: db}
}

// AddMessage adds a message and returns an interface{} compatible with tui.DBInterface
func (a *DBAdapter) AddMessage(ctx context.Context, conversationID uuid.UUID, role, content string) (interface{}, error) {
	return a.DB.AddMessage(ctx, conversationID, role, content)
}

// GenerateTitle generates a title from the first user message
func (a *DBAdapter) GenerateTitle(ctx context.Context, conversationID uuid.UUID) (string, error) {
	return a.DB.GenerateTitle(ctx, conversationID)
}