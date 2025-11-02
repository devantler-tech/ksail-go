package traefikinstaller_test

import (
	"context"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	traefikinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/traefik"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTraefikInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := traefikinstaller.NewMockHelmClient(t)
	installer := traefikinstaller.NewTraefikInstaller(client, kubeconfig, context, timeout)

	assert.NotNil(t, installer)
}

func TestTraefikInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newTraefikInstallerWithDefaults(t)
	expectTraefikInstall(t, client, nil)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestTraefikInstallerInstallRepositoryError(t *testing.T) {
	t.Parallel()

	installer, client := newTraefikInstallerWithDefaults(t)

	client.EXPECT().
		AddRepository(mock.Anything, mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
			return entry.Name == "traefik" && entry.URL == "https://traefik.github.io/charts"
		})).
		Return(assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add traefik repository")
}

func TestTraefikInstallerInstallChartError(t *testing.T) {
	t.Parallel()

	installer, client := newTraefikInstallerWithDefaults(t)

	client.EXPECT().
		AddRepository(mock.Anything, mock.Anything).
		Return(nil)

	client.EXPECT().
		InstallOrUpgradeChart(mock.Anything, mock.Anything).
		Return(nil, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install traefik chart")
}

func TestTraefikInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newTraefikInstallerWithDefaults(t)
	expectTraefikUninstall(t, client, nil)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestTraefikInstallerUninstallError(t *testing.T) {
	t.Parallel()

	installer, client := newTraefikInstallerWithDefaults(t)
	expectTraefikUninstall(t, client, assert.AnError)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall traefik release")
}

func newTraefikInstallerWithDefaults(
	t *testing.T,
) (*traefikinstaller.TraefikInstaller, *traefikinstaller.MockHelmClient) {
	t.Helper()
	client := traefikinstaller.NewMockHelmClient(t)
	installer := traefikinstaller.NewTraefikInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer, client
}

func expectTraefikInstall(t *testing.T, client *traefikinstaller.MockHelmClient, installErr error) {
	t.Helper()

	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				assert.Equal(t, "traefik", entry.Name)
				assert.Equal(t, "https://traefik.github.io/charts", entry.URL)

				return true
			}),
		).
		Return(nil)

	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				assert.Equal(t, "traefik", spec.ReleaseName)
				assert.Equal(t, "traefik/traefik", spec.ChartName)
				assert.Equal(t, "traefik", spec.Namespace)
				assert.True(t, spec.CreateNamespace)
				assert.True(t, spec.Atomic)
				assert.True(t, spec.Wait)
				assert.True(t, spec.WaitForJobs)

				return true
			}),
		).
		Return(&helm.ReleaseInfo{}, installErr)
}

func expectTraefikUninstall(
	t *testing.T,
	client *traefikinstaller.MockHelmClient,
	uninstallErr error,
) {
	t.Helper()

	client.EXPECT().
		UninstallRelease(mock.Anything, "traefik", "traefik").
		Return(uninstallErr)
}
