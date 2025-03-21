package main

import (
	"fmt"

	"github.com/hawk/mcgraph/internal/extensions"
	"github.com/spf13/cobra"
)

func init() {
	// Add enable and disable flags
	extCmd.AddCommand(enableCmd)
	extCmd.AddCommand(disableCmd)
	extCmd.AddCommand(listExtCmd)
	
	rootCmd.AddCommand(extCmd)
}

var extCmd = &cobra.Command{
	Use:     "ext",
	Aliases: []string{"extensions"},
	Short:   "Manage McGraph extensions",
	Long:    `Commands for managing and configuring McGraph extensions.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action is to list all extensions
		listExtensions()
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable the extensions system",
	Long:  `Enable the McGraph extensions system, allowing the use of plugins.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the current config
		config, err := extensions.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		// Check if already enabled
		if config.Enabled {
			fmt.Println("Extensions are already enabled.")
			return nil
		}
		
		// Enable extensions
		config.Enabled = true
		
		// Save the updated config
		if err := extensions.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		
		fmt.Println("Extensions have been enabled. Restart McGraph for the change to take effect.")
		return nil
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable the extensions system",
	Long:  `Disable the McGraph extensions system. This improves security by preventing any plugins from loading.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load the current config
		config, err := extensions.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		
		// Check if already disabled
		if !config.Enabled {
			fmt.Println("Extensions are already disabled.")
			return nil
		}
		
		// Disable extensions
		config.Enabled = false
		
		// Save the updated config
		if err := extensions.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		
		fmt.Println("Extensions have been disabled. Restart McGraph for the change to take effect.")
		return nil
	},
}

var listExtCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed extensions",
	Long:  `List all installed McGraph extensions and their commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		listExtensions()
	},
}

// listExtensions lists all installed extensions
func listExtensions() {
	if !extManager.IsEnabled() {
		fmt.Println("Extensions are currently disabled.")
		fmt.Println("To enable extensions, run: mcg ext enable")
		return
	}
	
	extensions := extManager.ListExtensions()
	if len(extensions) == 0 {
		fmt.Println("No extensions installed.")
		fmt.Println("Extensions should be placed in ~/.mcgraph/extensions/ as .so files.")
		return
	}
	
	fmt.Println("Installed extensions:")
	for _, ext := range extensions {
		fmt.Printf("\n/%s - %s\n", ext.Name(), ext.Description())
		
		// Print commands for this extension
		commands := extManager.GetCommands(ext.Name())
		if len(commands) == 0 {
			fmt.Println("  No commands available")
			continue
		}
		
		fmt.Println("  Commands:")
		for name, cmd := range commands {
			fmt.Printf("    /%s %s - %s\n", ext.Name(), name, cmd.Description())
		}
	}
	
	fmt.Println("\nUse extensions in chat with commands like: /filesystem ls")
	fmt.Println("For a list of available extensions and commands while in chat, type: /help")
}