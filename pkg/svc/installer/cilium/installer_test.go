package ciliuminstaller_test

import (
	"context"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCiliumInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := ciliuminstaller.NewMockHelmClient(t)
	installer := ciliuminstaller.NewCiliumInstaller(client, kubeconfig, context, timeout)

	assert.NotNil(t, installer)
}

func TestCiliumInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	client := ciliuminstaller.NewMockHelmClient(t)
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				assert.Equal(t, "cilium", spec.ReleaseName)
				assert.Equal(t, "cilium/cilium", spec.ChartName)
				assert.Equal(t, "kube-system", spec.Namespace)

				return true
			}),
		).
		Return(nil, nil)

	installer := ciliuminstaller.NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestCiliumInstallerInstallError(t *testing.T) {
	t.Parallel()

	client := ciliuminstaller.NewMockHelmClient(t)
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				assert.Equal(t, "cilium", spec.ReleaseName)
				assert.Equal(t, "cilium/cilium", spec.ChartName)
				assert.Equal(t, "kube-system", spec.Namespace)

				return true
			}),
		).
		Return(nil, assert.AnError)

	installer := ciliuminstaller.NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Cilium")
}

func TestCiliumInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	client := ciliuminstaller.NewMockHelmClient(t)
	client.EXPECT().
		UninstallRelease(mock.Anything, "cilium", "kube-system").
		Return(nil)

	installer := ciliuminstaller.NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestCiliumInstallerUninstallError(t *testing.T) {
	t.Parallel()

	client := ciliuminstaller.NewMockHelmClient(t)
	client.EXPECT().
		UninstallRelease(mock.Anything, "cilium", "kube-system").
		Return(assert.AnError)

	installer := ciliuminstaller.NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall cilium release")
}
