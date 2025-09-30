package stubs

import (
	"context"
	"errors"
)

// InstallerStub is a stub implementation of installer.Installer interface.
// It provides configurable behavior for testing without external dependencies.
type InstallerStub struct {
	InstallError   error
	UninstallError error

	// Track calls for verification
	InstallCalls   int
	UninstallCalls int
}

// NewInstallerStub creates a new InstallerStub with default success behavior.
func NewInstallerStub() *InstallerStub {
	return &InstallerStub{}
}

// Install simulates component installation.
func (i *InstallerStub) Install(ctx context.Context) error {
	i.InstallCalls++
	return i.InstallError
}

// Uninstall simulates component uninstallation.
func (i *InstallerStub) Uninstall(ctx context.Context) error {
	i.UninstallCalls++
	return i.UninstallError
}

// WithInstallError configures the stub to return an error on Install.
func (i *InstallerStub) WithInstallError(message string) *InstallerStub {
	i.InstallError = errors.New(message)
	return i
}

// WithUninstallError configures the stub to return an error on Uninstall.
func (i *InstallerStub) WithUninstallError(message string) *InstallerStub {
	i.UninstallError = errors.New(message)
	return i
}

// Reset clears all call tracking and errors.
func (i *InstallerStub) Reset() {
	i.InstallCalls = 0
	i.UninstallCalls = 0
	i.InstallError = nil
	i.UninstallError = nil
}
