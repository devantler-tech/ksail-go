package fluxinstaller_test

import (
	"context"
	"testing"
	"time"

	fluxinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/flux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFluxInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	installer := fluxinstaller.NewFluxInstaller(kubeconfig, context, timeout)

	assert.NotNil(t, installer)
}

func TestFluxInstallerInstall(t *testing.T) {
	t.Skip("Skipping test that requires flux CLI to be installed")
	t.Parallel()

	installer := newFluxInstallerWithDefaults(t)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestFluxInstallerUninstall(t *testing.T) {
	t.Skip("Skipping test that requires flux CLI to be installed")
	t.Parallel()

	installer := newFluxInstallerWithDefaults(t)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func newFluxInstallerWithDefaults(
	t *testing.T,
) *fluxinstaller.FluxInstaller {
	t.Helper()
	installer := fluxinstaller.NewFluxInstaller(
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer
}
