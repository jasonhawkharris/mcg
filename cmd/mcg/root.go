package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hawk/mcgraph/internal/db"
	"github.com/spf13/cobra"
)

// Global database connection
var dbConn *db.DB

var rootCmd = &cobra.Command{
	Use:   "mcg",
	Short: "mcgraph - A multi-LLM coding assistant",
	Long:  `mcgraph is a CLI tool that can use multiple LLMs for coding assistance.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip database initialization for commands that don't need it
		if cmd.Name() == "version" || cmd.Name() == "help" || cmd.Name() == "pick" || cmd.Name() == "llms" {
			return nil
		}

		// Initialize database connection
		ctx := context.Background()
		config := db.ConfigFromEnv()
		
		// Validate configuration and show setup instructions if using defaults
		if setupInstructions := db.ValidateConfig(config); setupInstructions != "" {
			fmt.Println(setupInstructions)
			return fmt.Errorf("database configuration required")
		}
		
		var err error
		dbConn, err = db.New(ctx, config)
		if err != nil {
			// Provide a more user-friendly error message
			return fmt.Errorf("failed to connect to database: %w\n\nPlease ensure PostgreSQL is running and properly configured.", err)
		}

		// Initialize database schema
		if err := dbConn.Init(ctx); err != nil {
			return fmt.Errorf("failed to initialize database schema: %w", err)
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Close database connection
		if dbConn != nil {
			dbConn.Close()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Default behavior when no subcommand is specified
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Configure the root command
}