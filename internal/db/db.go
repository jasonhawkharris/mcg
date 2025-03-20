package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Message represents a single message in a conversation
type Message struct {
	ID            uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Role          string    `json:"role"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
}

// Conversation represents a chat conversation with an LLM
type Conversation struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Messages  []Message `json:"messages,omitempty"`
}

// DB handles database operations
type DB struct {
	pool *pgxpool.Pool
}

// Config represents database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// DefaultConfig returns a default configuration for development
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "mcgraph",
	}
}

// ValidateConfig validates the database configuration and provides setup instructions
func ValidateConfig(config Config) string {
	if config.Host == "localhost" && config.Port == 5432 && 
	   config.User == "postgres" && config.Password == "postgres" && 
	   config.DBName == "mcgraph" {
		// Using default config, provide setup instructions
		return `
PostgreSQL Database Setup Required:

McGraph uses PostgreSQL to store conversation history. Please set up the database:

1. Install PostgreSQL if not already installed
2. Create the database: createdb mcgraph
3. Configure access by setting these environment variables:

export MCGRAPH_DB_HOST=localhost
export MCGRAPH_DB_PORT=5432
export MCGRAPH_DB_USER=postgres
export MCGRAPH_DB_PASSWORD=your_password
export MCGRAPH_DB_NAME=mcgraph

Or customize these values as needed for your PostgreSQL setup.
`
	}
	return ""
}

// ConfigFromEnv loads database configuration from environment variables
func ConfigFromEnv() Config {
	config := DefaultConfig()

	if host := os.Getenv("MCGRAPH_DB_HOST"); host != "" {
		config.Host = host
	}
	if port := os.Getenv("MCGRAPH_DB_PORT"); port != "" {
		var portInt int
		fmt.Sscanf(port, "%d", &portInt)
		if portInt > 0 {
			config.Port = portInt
		}
	}
	if user := os.Getenv("MCGRAPH_DB_USER"); user != "" {
		config.User = user
	}
	if password := os.Getenv("MCGRAPH_DB_PASSWORD"); password != "" {
		config.Password = password
	}
	if dbName := os.Getenv("MCGRAPH_DB_NAME"); dbName != "" {
		config.DBName = dbName
	}

	return config
}

// ConnectionString returns a PostgreSQL connection string
func (c Config) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", 
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

// New creates a new database connection
func New(ctx context.Context, config Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(config.ConnectionString())
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Init initializes the database schema
func (db *DB) Init(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id UUID PRIMARY KEY,
		title TEXT NOT NULL,
		model TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY,
		conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
	`

	_, err := db.pool.Exec(ctx, schema)
	return err
}

// CreateConversation creates a new conversation
func (db *DB) CreateConversation(ctx context.Context, title, model string) (Conversation, error) {
	id := uuid.New()
	now := time.Now().UTC()

	conversation := Conversation{
		ID:        id,
		Title:     title,
		Model:     model,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := db.pool.Exec(ctx,
		"INSERT INTO conversations (id, title, model, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		conversation.ID, conversation.Title, conversation.Model, conversation.CreatedAt, conversation.UpdatedAt,
	)

	return conversation, err
}

// GetConversation retrieves a conversation by ID
func (db *DB) GetConversation(ctx context.Context, id uuid.UUID) (Conversation, error) {
	var conversation Conversation

	err := db.pool.QueryRow(ctx,
		"SELECT id, title, model, created_at, updated_at FROM conversations WHERE id = $1",
		id,
	).Scan(&conversation.ID, &conversation.Title, &conversation.Model, &conversation.CreatedAt, &conversation.UpdatedAt)
	if err != nil {
		return Conversation{}, err
	}

	// Get messages for the conversation
	messages, err := db.GetMessages(ctx, id)
	if err != nil {
		return Conversation{}, err
	}

	conversation.Messages = messages
	return conversation, nil
}

// UpdateConversationTitle updates the title of a conversation
func (db *DB) UpdateConversationTitle(ctx context.Context, id uuid.UUID, title string) error {
	now := time.Now().UTC()
	_, err := db.pool.Exec(ctx,
		"UPDATE conversations SET title = $1, updated_at = $2 WHERE id = $3",
		title, now, id,
	)
	return err
}

// DeleteConversation deletes a conversation by ID
func (db *DB) DeleteConversation(ctx context.Context, id uuid.UUID) error {
	_, err := db.pool.Exec(ctx, "DELETE FROM conversations WHERE id = $1", id)
	return err
}

// ListConversations retrieves a list of all conversations
func (db *DB) ListConversations(ctx context.Context) ([]Conversation, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, title, model, created_at, updated_at FROM conversations ORDER BY updated_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conversation Conversation
		err := rows.Scan(&conversation.ID, &conversation.Title, &conversation.Model, &conversation.CreatedAt, &conversation.UpdatedAt)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, rows.Err()
}

// AddMessage adds a new message to a conversation
func (db *DB) AddMessage(ctx context.Context, conversationID uuid.UUID, role, content string) (Message, error) {
	id := uuid.New()
	now := time.Now().UTC()

	message := Message{
		ID:            id,
		ConversationID: conversationID,
		Role:          role,
		Content:       content,
		CreatedAt:     now,
	}

	_, err := db.pool.Exec(ctx,
		"INSERT INTO messages (id, conversation_id, role, content, created_at) VALUES ($1, $2, $3, $4, $5)",
		message.ID, message.ConversationID, message.Role, message.Content, message.CreatedAt,
	)
	if err != nil {
		return Message{}, err
	}

	// Update the conversation's updated_at timestamp
	_, err = db.pool.Exec(ctx,
		"UPDATE conversations SET updated_at = $1 WHERE id = $2",
		now, conversationID,
	)
	if err != nil {
		return Message{}, err
	}

	return message, nil
}

// GetMessages retrieves all messages for a conversation
func (db *DB) GetMessages(ctx context.Context, conversationID uuid.UUID) ([]Message, error) {
	rows, err := db.pool.Query(ctx,
		"SELECT id, conversation_id, role, content, created_at FROM messages WHERE conversation_id = $1 ORDER BY created_at ASC",
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var message Message
		err := rows.Scan(&message.ID, &message.ConversationID, &message.Role, &message.Content, &message.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

// GenerateTitle uses the first user message to generate a title for the conversation
func (db *DB) GenerateTitle(ctx context.Context, conversationID uuid.UUID) (string, error) {
	// Get the first user message
	var content string
	err := db.pool.QueryRow(ctx,
		"SELECT content FROM messages WHERE conversation_id = $1 AND role = 'user' ORDER BY created_at ASC LIMIT 1",
		conversationID,
	).Scan(&content)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "New Conversation", nil
		}
		return "", err
	}

	// Create a title from the message
	// For now, just truncate the first message; later we could use an LLM to generate a better title
	title := content
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	// Update the conversation title
	err = db.UpdateConversationTitle(ctx, conversationID, title)
	if err != nil {
		return "", err
	}

	return title, nil
}

// EnsureDBDir ensures that the database directory exists
func EnsureDBDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dbDir := filepath.Join(home, ".mcgraph", "db")
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return "", err
	}

	return dbDir, nil
}