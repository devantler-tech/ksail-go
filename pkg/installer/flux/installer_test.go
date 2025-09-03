package fluxinstaller_test

import (
	"testing"
	"time"

	fluxinstaller "github.com/devantler-tech/ksail-go/pkg/installer/flux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewFluxInstaller(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	// Act
	client := fluxinstaller.NewMockHelmClient(t)
	installer := fluxinstaller.NewFluxInstaller(client, kubeconfig, context, timeout)

	// Assert
	assert.NotNil(t, installer)
}

func TestFluxInstaller_Install_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Install(mock.Anything, mock.Anything).Return(nil)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	// Act
	err := installer.Install()

	// Assert
	require.NoError(t, err)
}

func TestFluxInstaller_Install_Error(t *testing.T) {
	t.Parallel()

	// Arrange
	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Install(mock.Anything, mock.Anything).Return(assert.AnError)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Flux operator")
}

func TestFluxInstaller_Uninstall_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Uninstall("flux-operator").Return(nil)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.NoError(t, err)
}

func TestFluxInstaller_Uninstall_Error(t *testing.T) {
	t.Parallel()

	// Arrange
	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Uninstall("flux-operator").Return(assert.AnError)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall flux-operator release")
}
