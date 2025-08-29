package fluxinstaller_test

import (
	"testing"
	"time"

	fluxinstaller "github.com/devantler-tech/ksail-go/pkg/installer/flux"
	"github.com/stretchr/testify/assert"
)

func TestNewFluxInstaller(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	// Act
	installer := fluxinstaller.NewFluxInstaller(kubeconfig, context, timeout)

	// Assert
	assert.NotNil(t, installer)
}

func TestFluxInstaller_Install_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := fluxinstaller.NewFluxInstaller(
		"/nonexistent/kubeconfig",
		"test-context",
		5*time.Minute,
	)

	// Act
	err := installer.Install()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestFluxInstaller_Uninstall_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := fluxinstaller.NewFluxInstaller(
		"/nonexistent/kubeconfig",
		"test-context",
		5*time.Minute,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}