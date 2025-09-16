package fluxinstaller_test

import (
	"context"
	"testing"
	"time"

	fluxinstaller "github.com/devantler-tech/ksail-go/pkg/installer/flux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewFluxInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := fluxinstaller.NewMockHelmClient(t)
	installer := fluxinstaller.NewFluxInstaller(client, kubeconfig, context, timeout)

	assert.NotNil(t, installer)
}

func TestFluxInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Install(mock.Anything, mock.Anything).Return(nil)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestFluxInstallerInstallError(t *testing.T) {
	t.Parallel()

	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Install(mock.Anything, mock.Anything).Return(assert.AnError)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Flux operator")
}

func TestFluxInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Uninstall("flux-operator").Return(nil)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestFluxInstallerUninstallError(t *testing.T) {
	t.Parallel()

	client := fluxinstaller.NewMockHelmClient(t)
	client.EXPECT().Uninstall("flux-operator").Return(assert.AnError)

	installer := fluxinstaller.NewFluxInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall flux-operator release")
}
