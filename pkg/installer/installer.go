// Package installer provides functionality for installing and uninstalling components.
package installer

// Installer defines methods for installing and uninstalling components.
type Installer interface {
	// Install installs the component.
	Install() error

	// Uninstall uninstalls the component.
	Uninstall() error
}