package argocdinstaller_test

import (
	"context"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	argocdinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/argocd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewArgoCDInstaller(t *testing.T) {
	t.Parallel()

	timeout := 5 * time.Second

	client := helm.NewMockInterface(t)
	installer := argocdinstaller.NewArgoCDInstaller(client, timeout)

	assert.NotNil(t, installer)
}

func TestArgoCDInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newArgoCDInstallerWithDefaults(t)
	expectArgoCDInstall(t, client, nil)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestArgoCDInstallerInstallError(t *testing.T) {
	t.Parallel()

	installer, client := newArgoCDInstallerWithDefaults(t)
	expectArgoCDInstall(t, client, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install ArgoCD")
}

func TestArgoCDInstallerInstallAddRepositoryError(t *testing.T) {
	t.Parallel()

	installer, client := newArgoCDInstallerWithDefaults(t)
	expectArgoCDAddRepository(t, client, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add argo repository")
}

func TestArgoCDInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newArgoCDInstallerWithDefaults(t)
	expectArgoCDUninstall(t, client, nil)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestArgoCDInstallerUninstallError(t *testing.T) {
	t.Parallel()

	installer, client := newArgoCDInstallerWithDefaults(t)
	expectArgoCDUninstall(t, client, assert.AnError)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall argocd release")
}

func newArgoCDInstallerWithDefaults(
	t *testing.T,
) (*argocdinstaller.ArgoCDInstaller, *helm.MockInterface) {
	t.Helper()
	client := helm.NewMockInterface(t)
	installer := argocdinstaller.NewArgoCDInstaller(
		client,
		5*time.Second,
	)

	return installer, client
}

func expectArgoCDAddRepository(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				assert.Equal(t, "argo", entry.Name)
				assert.Equal(t, "https://argoproj.github.io/argo-helm", entry.URL)

				return true
			}),
		).
		Return(err)
}

func expectArgoCDInstall(t *testing.T, client *helm.MockInterface, installErr error) {
	t.Helper()
	expectArgoCDAddRepository(t, client, nil)
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				assert.Equal(t, "argocd", spec.ReleaseName)
				assert.Equal(t, "argo/argo-cd", spec.ChartName)
				assert.Equal(t, "argocd", spec.Namespace)
				assert.Equal(t, "https://argoproj.github.io/argo-helm", spec.RepoURL)
				assert.True(t, spec.CreateNamespace)
				assert.True(t, spec.Atomic)
				assert.True(t, spec.UpgradeCRDs)

				return true
			}),
		).
		Return(nil, installErr)
}

func expectArgoCDUninstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "argocd", "argocd").
		Return(err)
}
