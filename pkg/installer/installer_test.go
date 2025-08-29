package installer_test

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/installer"
)

// MockInstaller is a simple mock implementation of the Installer interface for testing.
type MockInstaller struct {
	installError   error
	uninstallError error
}

// Install implements installer.Installer.
func (m *MockInstaller) Install() error {
	return m.installError
}

// Uninstall implements installer.Installer.
func (m *MockInstaller) Uninstall() error {
	return m.uninstallError
}

// Verify that MockInstaller implements the Installer interface.
var _ installer.Installer = (*MockInstaller)(nil)

func TestInstallerInterface_CanBeImplemented(t *testing.T) {
	t.Parallel()

	// Arrange
	mock := &MockInstaller{
		installError:   nil,
		uninstallError: nil,
	}

	// Act & Assert - verify interface can be used
	var installer installer.Installer = mock

	// Verify methods can be called
	err := installer.Install()
	if err != nil {
		t.Errorf("Install() returned unexpected error: %v", err)
	}

	err = installer.Uninstall()
	if err != nil {
		t.Errorf("Uninstall() returned unexpected error: %v", err)
	}
}

func TestInstallerInterface_HandlesErrors(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedErr := &MockError{message: "test error"}
	mock := &MockInstaller{
		installError:   expectedErr,
		uninstallError: expectedErr,
	}

	// Act
	installErr := mock.Install()
	uninstallErr := mock.Uninstall()

	// Assert
	if !errors.Is(installErr, expectedErr) {
		t.Errorf("Install() error = %v, want %v", installErr, expectedErr)
	}

	if !errors.Is(uninstallErr, expectedErr) {
		t.Errorf("Uninstall() error = %v, want %v", uninstallErr, expectedErr)
	}
}

// MockError is a simple error implementation for testing.
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}