package extensions

import (
	"fmt"
	"os"
	"strings"
)

// SimpleExtensionLoader directly loads a simple extension from disk
// This is a fallback mechanism for loading extensions when the plugin system fails
func (m *Manager) LoadSimpleExtensions() {
	// Add a basic system extension
	sysExt := &SimpleExtension{}
	m.extensions[sysExt.Name()] = sysExt
	
	// Add commands
	m.commands[sysExt.Name()] = map[string]Command{
		"ls":   &SimpleCommand{name: "ls", description: "List files in a directory", execute: commandLS},
		"pwd":  &SimpleCommand{name: "pwd", description: "Print working directory", execute: commandPWD},
		"read": &SimpleCommand{name: "read", description: "Read file contents", execute: commandRead},
	}
	
	fmt.Printf("Loaded built-in extension: %s - %s\n", sysExt.Name(), sysExt.Description())
}

// SimpleExtension is a basic built-in extension
type SimpleExtension struct{}

// Name returns the extension name
func (e *SimpleExtension) Name() string {
	return "system"
}

// Description returns the extension description
func (e *SimpleExtension) Description() string {
	return "Basic system commands"
}

// Commands returns the extension commands
func (e *SimpleExtension) Commands() []Command {
	return []Command{
		&SimpleCommand{name: "ls", description: "List files in a directory", execute: commandLS},
		&SimpleCommand{name: "pwd", description: "Print working directory", execute: commandPWD},
		&SimpleCommand{name: "read", description: "Read file contents", execute: commandRead},
	}
}

// SimpleCommand is a basic command
type SimpleCommand struct {
	name        string
	description string
	execute     func(args []string) (string, error)
}

// Name returns the command name
func (c *SimpleCommand) Name() string {
	return c.name
}

// Description returns the command description
func (c *SimpleCommand) Description() string {
	return c.description
}

// Execute runs the command
func (c *SimpleCommand) Execute(args []string) (string, error) {
	return c.execute(args)
}

// commandPWD implements the pwd command
func commandPWD(args []string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

// commandLS implements the ls command
func commandLS(args []string) (string, error) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Contents of %s:\n", path))
	
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}

		sb.WriteString(fmt.Sprintf("%10d %s\n", info.Size(), name))
	}

	return sb.String(), nil
}

// commandRead implements the read command
func commandRead(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("please provide a file path")
	}

	path := args[0]
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Only allow small files
	if len(content) > 1024*1024 {
		return "", fmt.Errorf("file too large (>1MB)")
	}

	return string(content), nil
}