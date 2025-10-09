package helminstaller_test

import (
	"context"
	"testing"
	"time"

	helminstaller "github.com/devantler-tech/ksail-go/pkg/installer/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewHelmInstaller(t *testing.T) {
	t.Parallel()

	releaseName := "test-release"
	chartName := "test/chart"
	namespace := "test-namespace"
	version := "1.0.0"
	valuesYaml := "key: value"
	timeout := 5 * time.Minute

	client := helminstaller.NewMockHelmClient(t)
	installer := helminstaller.NewHelmInstaller(
		client,
		releaseName,
		chartName,
		namespace,
		version,
		valuesYaml,
		timeout,
	)

	assert.NotNil(t, installer)
}

func TestHelmInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	client := helminstaller.NewMockHelmClient(t)
	client.EXPECT().InstallChart(mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	installer := helminstaller.NewHelmInstaller(
		client,
		"test-release",
		"test/chart",
		"test-namespace",
		"1.0.0",
		"",
		5*time.Second,
	)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestHelmInstallerInstallError(t *testing.T) {
	t.Parallel()

	client := helminstaller.NewMockHelmClient(t)
	client.EXPECT().InstallChart(mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)

	installer := helminstaller.NewHelmInstaller(
		client,
		"test-release",
		"test/chart",
		"test-namespace",
		"1.0.0",
		"",
		5*time.Second,
	)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Helm chart")
}

func TestHelmInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	client := helminstaller.NewMockHelmClient(t)
	client.EXPECT().UninstallReleaseByName("test-release").Return(nil)

	installer := helminstaller.NewHelmInstaller(
		client,
		"test-release",
		"test/chart",
		"test-namespace",
		"1.0.0",
		"",
		5*time.Second,
	)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestHelmInstallerUninstallError(t *testing.T) {
	t.Parallel()

	client := helminstaller.NewMockHelmClient(t)
	client.EXPECT().UninstallReleaseByName("test-release").Return(assert.AnError)

	installer := helminstaller.NewHelmInstaller(
		client,
		"test-release",
		"test/chart",
		"test-namespace",
		"1.0.0",
		"",
		5*time.Second,
	)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall release test-release")
}
