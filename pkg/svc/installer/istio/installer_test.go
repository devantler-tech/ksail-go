package istioinstaller_test

import (
	"context"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	istioinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/istio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testTimeout = 5 * time.Second

func TestNewIstioInstaller(t *testing.T) {
	t.Parallel()

	timeout := 5 * time.Minute

	client := helm.NewMockInterface(t)
	installer := istioinstaller.NewIstioInstaller(client, timeout)

	assert.NotNil(t, installer)
}

func TestIstioInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newIstioInstallerWithDefaults(t)
	expectIstioInstall(t, client, nil, nil)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestIstioInstallerInstallErrorOnBase(t *testing.T) {
	t.Parallel()

	installer, client := newIstioInstallerWithDefaults(t)
	expectIstioBaseInstall(t, client, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Istio base")
}

func TestIstioInstallerInstallErrorOnIstiod(t *testing.T) {
	t.Parallel()

	installer, client := newIstioInstallerWithDefaults(t)
	expectIstioBaseInstall(t, client, nil)
	expectIstiodInstall(t, client, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Istiod")
}

func TestIstioInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newIstioInstallerWithDefaults(t)
	expectIstioUninstall(t, client, nil, nil)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestIstioInstallerUninstallErrorOnIstiod(t *testing.T) {
	t.Parallel()

	installer, client := newIstioInstallerWithDefaults(t)
	client.EXPECT().
		UninstallRelease(mock.Anything, "istiod", "istio-system").
		Return(assert.AnError)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall istiod release")
}

func TestIstioInstallerUninstallErrorOnBase(t *testing.T) {
	t.Parallel()

	installer, client := newIstioInstallerWithDefaults(t)
	client.EXPECT().
		UninstallRelease(mock.Anything, "istiod", "istio-system").
		Return(nil)
	client.EXPECT().
		UninstallRelease(mock.Anything, "istio-base", "istio-system").
		Return(assert.AnError)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall istio-base release")
}

func newIstioInstallerWithDefaults(
	t *testing.T,
) (*istioinstaller.IstioInstaller, *helm.MockInterface) {
	t.Helper()
	client := helm.NewMockInterface(t)
	installer := istioinstaller.NewIstioInstaller(
		client,
		testTimeout,
	)

	return installer, client
}

func expectIstioInstall(
	t *testing.T,
	client *helm.MockInterface,
	baseErr error,
	istiodErr error,
) {
	t.Helper()
	expectIstioBaseInstall(t, client, baseErr)

	if baseErr == nil {
		expectIstiodInstall(t, client, istiodErr)
	}
}

func expectIstioBaseInstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	expectChartInstall(t, client, "istio-base", "istio/base", err)
}

func expectIstiodInstall(t *testing.T, client *helm.MockInterface, err error) {
	t.Helper()
	expectChartInstall(t, client, "istiod", "istio/istiod", err)
}

// expectChartInstall is a helper function to set up mock expectations for chart installation.
func expectChartInstall(
	t *testing.T,
	client *helm.MockInterface,
	releaseName, chartName string,
	err error,
) {
	t.Helper()

	client.EXPECT().
		AddRepository(mock.Anything, mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
			return entry.Name == "istio" &&
				entry.URL == "https://istio-release.storage.googleapis.com/charts"
		})).
		Return(nil).
		Once()

	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				return spec.ReleaseName == releaseName &&
					spec.ChartName == chartName &&
					spec.Namespace == "istio-system" &&
					spec.CreateNamespace == true &&
					spec.Atomic == true &&
					spec.UpgradeCRDs == true &&
					spec.Wait == true &&
					spec.WaitForJobs == true &&
					spec.Timeout == testTimeout
			}),
		).
		Return(nil, err).
		Once()
}

func expectIstioUninstall(
	t *testing.T,
	client *helm.MockInterface,
	istiodErr error,
	baseErr error,
) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "istiod", "istio-system").
		Return(istiodErr)

	if istiodErr == nil {
		client.EXPECT().
			UninstallRelease(mock.Anything, "istio-base", "istio-system").
			Return(baseErr)
	}
}
