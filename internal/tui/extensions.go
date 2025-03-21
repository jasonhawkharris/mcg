package tui

import (
	"github.com/hawk/mcgraph/internal/extensions"
)

// extManager is the package-level reference to the extension manager
var extManager *extensions.Manager

// SetExtensionManager sets the extension manager for the TUI package
func SetExtensionManager(manager *extensions.Manager) {
	extManager = manager
}