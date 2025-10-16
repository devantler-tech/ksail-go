package ciliuminstaller

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCiliumInstaller(t *testing.T) {
	t.Parallel()

	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	client := NewMockHelmClient(t)
	installer := NewCiliumInstaller(client, kubeconfig, context, timeout)

	assert.NotNil(t, installer)
}

func TestCiliumInstallerInstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newDefaultInstaller(t)
	expectCiliumAddRepository(t, client, nil)
	expectCiliumInstallChart(t, client, nil)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestCiliumInstallerInstallError(t *testing.T) {
	t.Parallel()

	installer, client := newDefaultInstaller(t)
	expectCiliumAddRepository(t, client, nil)
	expectCiliumInstallChart(t, client, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Cilium")
}

func TestCiliumInstallerInstallAddRepositoryError(t *testing.T) {
	t.Parallel()

	installer, client := newDefaultInstaller(t)
	expectCiliumAddRepository(t, client, assert.AnError)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add cilium repository")
}

func TestCiliumInstallerUninstallSuccess(t *testing.T) {
	t.Parallel()

	installer, client := newDefaultInstaller(t)
	expectCiliumUninstall(t, client, nil)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestCiliumInstallerUninstallError(t *testing.T) {
	t.Parallel()

	installer, client := newDefaultInstaller(t)
	expectCiliumUninstall(t, client, assert.AnError)

	err := installer.Uninstall(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall cilium release")
}

func TestApplyDefaultValuesSetsOperatorReplicaWhenMissing(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{}

	applyDefaultValues(spec)

	assert.Equal(t, "1", spec.SetJSONVals["operator.replicas"])
}

func TestApplyDefaultValuesDoesNotOverrideExistingValues(t *testing.T) {
	t.Parallel()

	spec := &helm.ChartSpec{
		SetJSONVals: map[string]string{
			"operator.replicas": "3",
		},
	}

	applyDefaultValues(spec)

	assert.Equal(t, "3", spec.SetJSONVals["operator.replicas"])
}

func TestCiliumInstallerWaitForReadinessUsesInjectedFunc(t *testing.T) {
	t.Parallel()

	client := NewMockHelmClient(t)
	installer := NewCiliumInstaller(client, "kubeconfig", "", time.Second)

	called := false

	installer.SetWaitForReadinessFunc(func(_ context.Context) error {
		called = true

		return nil
	})

	err := installer.WaitForReadiness(context.Background())

	require.NoError(t, err)
	assert.True(t, called, "custom wait function should be invoked")
}

func TestCiliumInstallerWaitForReadinessRestoresDefaultWhenNil(t *testing.T) {
	t.Parallel()

	client := NewMockHelmClient(t)
	installer := NewCiliumInstaller(client, "kubeconfig", "", time.Second)

	defaultFn := installer.waitFn
	require.NotNil(t, defaultFn)
	defaultPtr := reflect.ValueOf(defaultFn).Pointer()

	installer.SetWaitForReadinessFunc(func(_ context.Context) error { return nil })
	customPtr := reflect.ValueOf(installer.waitFn).Pointer()
	require.NotEqual(t, defaultPtr, customPtr)

	installer.SetWaitForReadinessFunc(nil)
	restoredPtr := reflect.ValueOf(installer.waitFn).Pointer()
	require.Equal(t, defaultPtr, restoredPtr)
}

func newDefaultInstaller(t *testing.T) (*CiliumInstaller, *MockHelmClient) {
	t.Helper()
	client := NewMockHelmClient(t)
	installer := NewCiliumInstaller(
		client,
		"~/.kube/config",
		"test-context",
		5*time.Second,
	)

	return installer, client
}

func expectCiliumAddRepository(t *testing.T, client *MockHelmClient, err error) {
	t.Helper()
	client.EXPECT().
		AddRepository(
			mock.Anything,
			mock.MatchedBy(func(entry *helm.RepositoryEntry) bool {
				assert.Equal(t, "cilium", entry.Name)
				assert.Equal(t, "https://helm.cilium.io", entry.URL)

				return true
			}),
		).
		Return(err)
}

func expectCiliumInstallChart(t *testing.T, client *MockHelmClient, installErr error) {
	t.Helper()
	client.EXPECT().
		InstallOrUpgradeChart(
			mock.Anything,
			mock.MatchedBy(func(spec *helm.ChartSpec) bool {
				assert.Equal(t, "cilium", spec.ReleaseName)
				assert.Equal(t, "cilium/cilium", spec.ChartName)
				assert.Equal(t, "kube-system", spec.Namespace)
				assert.Equal(t, "https://helm.cilium.io", spec.RepoURL)
				assert.True(t, spec.Wait)
				assert.True(t, spec.WaitForJobs)
				assert.Equal(t, "1", spec.SetJSONVals["operator.replicas"])

				return true
			}),
		).
		Return(nil, installErr)
}

func expectCiliumUninstall(t *testing.T, client *MockHelmClient, err error) {
	t.Helper()
	client.EXPECT().
		UninstallRelease(mock.Anything, "cilium", "kube-system").
		Return(err)
}
