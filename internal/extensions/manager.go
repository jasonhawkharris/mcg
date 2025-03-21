package extensions

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

// Extension is the interface that all extensions must implement
type Extension interface {
	Name() string
	Description() string
	Commands() []Command
}

// Command represents a command that can be executed by an extension
type Command interface {
	Name() string
	Description() string
	Execute(args []string) (string, error)
}

// Manager manages loaded extensions
type Manager struct {
	extensions map[string]Extension
	commands   map[string]map[string]Command // map[extensionName][commandName]Command
	enabled    bool
}

// NewManager creates a new extension manager
func NewManager(enabled bool) *Manager {
	return &Manager{
		extensions: make(map[string]Extension),
		commands:   make(map[string]map[string]Command),
		enabled:    enabled,
	}
}

// LoadExtensions loads extensions from the extensions directory
func (m *Manager) LoadExtensions() error {
	if !m.enabled {
		fmt.Println("Extensions are disabled")
		return nil
	}

	// First load built-in extensions that don't rely on plugins
	m.LoadSimpleExtensions()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	extDir := filepath.Join(homeDir, ".mcgraph", "extensions")
	if _, err := os.Stat(extDir); os.IsNotExist(err) {
		// Extensions directory doesn't exist, create it
		if err := os.MkdirAll(extDir, 0755); err != nil {
			return fmt.Errorf("failed to create extensions directory: %w", err)
		}
		return nil
	}

	// Read the directory
	entries, err := os.ReadDir(extDir)
	if err != nil {
		return fmt.Errorf("failed to read extensions directory: %w", err)
	}

	// Load each .so file as a plugin
	pluginsLoaded := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".so") {
			path := filepath.Join(extDir, entry.Name())
			if err := m.LoadExtension(path); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load extension %s: %v\n", path, err)
				continue
			}
			pluginsLoaded = true
		}
	}
	
	if !pluginsLoaded {
		fmt.Println("No plugin extensions loaded. Using built-in extensions only.")
	}

	return nil
}

// LoadExtension loads a specific extension from a file
func (m *Manager) LoadExtension(path string) error {
	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look up the Extension symbol
	sym, err := p.Lookup("Extension")
	if err != nil {
		return fmt.Errorf("plugin does not export 'Extension' symbol: %w", err)
	}

	// Get the extension as an interface{}
	extValue, ok := sym.(interface{})
	if !ok || extValue == nil {
		return fmt.Errorf("plugin's Extension symbol is not a valid interface value")
	}
	
	// For debugging
	fmt.Printf("Extension type: %T\n", extValue)
	
	// Try to use reflection to check if it has the required methods
	extName, err := callExtensionMethod(extValue, "Name")
	if err != nil {
		return fmt.Errorf("extension doesn't implement Name(): %w", err)
	}
	name, ok := extName.(string)
	if !ok {
		return fmt.Errorf("extension's Name() doesn't return a string")
	}
	
	// Get description
	extDesc, err := callExtensionMethod(extValue, "Description")
	if err != nil {
		return fmt.Errorf("extension doesn't implement Description(): %w", err)
	}
	description, ok := extDesc.(string)
	if !ok {
		return fmt.Errorf("extension's Description() doesn't return a string")
	}
	
	// Get commands
	extCmds, err := callExtensionMethod(extValue, "Commands")
	if err != nil {
		return fmt.Errorf("extension doesn't implement Commands(): %w", err)
	}
	
	// Create a wrapper extension that can handle any type
	wrapper := &ExtensionWrapper{
		name:        name,
		description: description,
		value:       extValue,
	}
	
	// Also save any commands
	cmdSlice, ok := extCmds.([]interface{})
	if !ok {
		// Try to convert whatever we got to a slice of interfaces
		cmdSlice = convertToInterfaceSlice(extCmds)
	}
	
	// Create a wrapper command for each command
	commands := make(map[string]Command)
	for _, cmdValue := range cmdSlice {
		cmdName, err := callExtensionMethod(cmdValue, "Name")
		if err != nil {
			fmt.Printf("Warning: command doesn't implement Name(): %v\n", err)
			continue
		}
		
		cmdNameStr, ok := cmdName.(string)
		if !ok {
			fmt.Printf("Warning: command's Name() doesn't return a string\n")
			continue
		}
		
		cmdDesc, err := callExtensionMethod(cmdValue, "Description")
		if err != nil {
			fmt.Printf("Warning: command %s doesn't implement Description(): %v\n", cmdNameStr, err)
			continue
		}
		
		cmdDescStr, ok := cmdDesc.(string)
		if !ok {
			fmt.Printf("Warning: command %s's Description() doesn't return a string\n", cmdNameStr)
			continue
		}
		
		// Create a wrapper command
		commands[cmdNameStr] = &CommandWrapper{
			name:        cmdNameStr,
			description: cmdDescStr,
			value:       cmdValue,
		}
	}
	
	// Register the extension wrapper
	m.extensions[name] = wrapper
	m.commands[name] = commands

	fmt.Printf("Loaded extension: %s - %s\n", name, description)
	return nil
}

// ExecuteCommand executes a command from an extension
func (m *Manager) ExecuteCommand(extName, cmdName string, args []string) (string, error) {
	if !m.enabled {
		return "", fmt.Errorf("extensions are disabled")
	}

	// Check if the extension exists
	_, ok := m.extensions[extName]
	if !ok {
		return "", fmt.Errorf("extension '%s' not found", extName)
	}

	// Check if the command exists
	cmd, ok := m.commands[extName][cmdName]
	if !ok {
		return "", fmt.Errorf("command '%s' not found in extension '%s'", cmdName, extName)
	}

	// Execute the command
	return cmd.Execute(args)
}

// GetExtension returns an extension by name
func (m *Manager) GetExtension(name string) (Extension, bool) {
	ext, ok := m.extensions[name]
	return ext, ok
}

// GetCommands returns all commands for an extension
func (m *Manager) GetCommands(extName string) map[string]Command {
	return m.commands[extName]
}

// ListExtensions returns a list of loaded extensions
func (m *Manager) ListExtensions() []Extension {
	extensions := make([]Extension, 0, len(m.extensions))
	for _, ext := range m.extensions {
		extensions = append(extensions, ext)
	}
	return extensions
}

// IsEnabled returns whether extensions are enabled
func (m *Manager) IsEnabled() bool {
	return m.enabled
}