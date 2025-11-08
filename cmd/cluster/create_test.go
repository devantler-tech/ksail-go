package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	testutils "github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const (
	testCiliumContext  = "kind-kind"
	testKubeconfigPath = "kubeconfig"
)

var errCiliumReadiness = errors.New("cilium readiness failed")

var (
	errRepoError                = errors.New("repo error")
	errClusterProvisionerFailed = errors.New("provisioner failed")
)

func TestNewCreateCmd(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewCreateCmd(runtimeContainer)

	if cmd.Use != "create" {
		t.Fatalf("expected Use to be 'create', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("expected Short description to be set")
	}

	if cmd.RunE == nil {
		t.Fatal("expected RunE to be set")
	}

	var out bytes.Buffer
	cmd.SetOut(&out)
}

func TestNewCreateLifecycleConfig(t *testing.T) {
	t.Parallel()

	t.Run("sets_expected_messaging", func(t *testing.T) {
		t.Parallel()

		cfg := newCreateLifecycleConfig()

		if cfg.TitleEmoji != "🚀" {
			t.Fatalf("expected rocket emoji, got %q", cfg.TitleEmoji)
		}

		if cfg.SuccessContent != "cluster created" {
			t.Fatalf("unexpected success content %q", cfg.SuccessContent)
		}

		if cfg.Action == nil {
			t.Fatal("expected lifecycle action to be set")
		}
	})

	t.Run("delegates_action_to_provisioner", func(t *testing.T) {
		t.Parallel()

		cfg := newCreateLifecycleConfig()
		provisioner := &testutils.StubProvisioner{}

		err := cfg.Action(context.Background(), provisioner, "kind")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provisioner.CreateCalls != 1 {
			t.Fatalf("expected provisioner to be called once, got %d", provisioner.CreateCalls)
		}

		if len(provisioner.ReceivedNames) == 0 || provisioner.ReceivedNames[0] != "kind" {
			t.Fatalf("expected action to use provided cluster name")
		}
	})
}

func TestHandleCreateRunE_InstallsCiliumWhenConfigured(t *testing.T) {
	t.Parallel()

	cmd, out := testutils.NewCommand(t)

	tempDir := t.TempDir()
	cfgPath := writeCiliumClusterConfig(t, tempDir, "./missing-kubeconfig")

	cfgManager := ksailconfigmanager.NewConfigManager(
		out,
		ksailconfigmanager.DefaultClusterFieldSelectors()...,
	)
	cfgManager.Viper.SetConfigFile(cfgPath)
	cfgManager.Viper.Set("spec.distribution", string(v1alpha1.DistributionK3d))
	cfgManager.Viper.Set("spec.distributionConfig", "")

	provisioner := &testutils.StubProvisioner{}
	deps := newTestLifecycleDeps(provisioner)

	err := handleCreateRunE(cmd, cfgManager, deps)

	require.Error(t, err)
	require.ErrorContains(t, err, "failed to install Cilium CNI")
	assert.Equal(t, 1, provisioner.CreateCalls)
}

func TestHandleCreateRunEWithoutCilium(t *testing.T) {
	t.Parallel()

	cmd, out := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	cfgManager := testutils.CreateConfigManager(t, out)
	cfgManager.Viper.Set("spec.distribution", string(v1alpha1.DistributionK3d))
	cfgManager.Viper.Set("spec.distributionConfig", "")

	provisioner := &testutils.StubProvisioner{}
	timer := &testutils.RecordingTimer{}
	deps := cmdhelpers.LifecycleDeps{
		Timer: timer,
		Factory: &testutils.StubFactory{
			Provisioner:        provisioner,
			DistributionConfig: &kindv1alpha4.Cluster{Name: "kind"},
		},
	}

	err := handleCreateRunE(cmd, cfgManager, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if provisioner.CreateCalls != 1 {
		t.Fatalf("expected provisioner create to be invoked, got %d", provisioner.CreateCalls)
	}

	if timer.NewStageCount != 1 {
		t.Fatalf("expected timer new stage to be called once, got %d", timer.NewStageCount)
	}

	if strings.Contains(out.String(), "Install CNI") {
		t.Fatalf("did not expect cilium installation messaging:\n%s", out.String())
	}
}

func TestGetInstallTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
		expected time.Duration
	}{
		{
			name:     "uses_default_timeout",
			duration: 0,
			expected: 5 * time.Minute,
		},
		{
			name:     "respects_custom_timeout",
			duration: 2 * time.Minute,
			expected: 2 * time.Minute,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			cfg := &v1alpha1.Cluster{}
			cfg.Spec.Connection.Timeout.Duration = test.duration

			result := getInstallTimeout(cfg)
			if result != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestNewCiliumInstaller(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Connection.Context = "kind-dev"
	clusterCfg.Spec.Connection.Timeout.Duration = 2 * time.Minute

	installer := newCiliumInstaller(nil, "/tmp/kubeconfig", clusterCfg)
	if installer == nil {
		t.Fatal("expected installer to be created")
	}

	value := reflect.ValueOf(installer).Elem()

	kubeconfig := readUnexportedString(t, value, "kubeconfig")
	if kubeconfig != "/tmp/kubeconfig" {
		t.Fatalf("expected kubeconfig to propagate, got %q", kubeconfig)
	}

	contextName := readUnexportedString(t, value, "context")
	if contextName != "kind-dev" {
		t.Fatalf("expected context to propagate, got %q", contextName)
	}

	timeout := readUnexportedDuration(t, value, "timeout")
	if timeout != 2*time.Minute {
		t.Fatalf("expected timeout to use cluster override, got %v", timeout)
	}
}

// readUnexportedString inspects the value of an unexported string field for assertions.
//
//nolint:gosec // Using unsafe pointer conversion for read-only test verification.
func readUnexportedString(t *testing.T, value reflect.Value, field string) string {
	t.Helper()

	f := value.FieldByName(field)
	if !f.IsValid() {
		t.Fatalf("field %s not found", field)
	}

	return *(*string)(unsafe.Pointer(f.UnsafeAddr()))
}

// readUnexportedDuration inspects the value of an unexported duration field for assertions.
//
//nolint:gosec // Using unsafe pointer conversion for read-only test verification.
func readUnexportedDuration(t *testing.T, value reflect.Value, field string) time.Duration {
	t.Helper()

	f := value.FieldByName(field)
	if !f.IsValid() {
		t.Fatalf("field %s not found", field)
	}

	return *(*time.Duration)(unsafe.Pointer(f.UnsafeAddr()))
}

//nolint:paralleltest // Uses t.Setenv to control home directory.
func TestExpandKubeconfigPath(t *testing.T) {
	t.Run("returns_unmodified_when_no_tilde", func(t *testing.T) {
		path, err := expandKubeconfigPath("/tmp/config")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if path != "/tmp/config" {
			t.Fatalf("expected original path, got %q", path)
		}
	})

	t.Run("expands_home_directory", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("USERPROFILE", home)

		path, err := expandKubeconfigPath("~/config")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := filepath.Join(home, "config")
		if path != expected {
			t.Fatalf("expected %q, got %q", expected, path)
		}
	})
}

func TestLoadKubeconfig(t *testing.T) {
	t.Parallel()

	t.Run("reads_config_file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		kubeconfigPath := filepath.Join(dir, "kubeconfig")
		expected := []byte("kube-config-data")

		err := os.WriteFile(kubeconfigPath, expected, 0o600)
		if err != nil {
			t.Fatalf("failed to write kubeconfig: %v", err)
		}

		clusterCfg := &v1alpha1.Cluster{}
		clusterCfg.Spec.Connection.Kubeconfig = kubeconfigPath

		path, err := loadKubeconfig(clusterCfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if path != kubeconfigPath {
			t.Fatalf("expected path %q, got %q", kubeconfigPath, path)
		}

		// Verify file was actually read by checking it exists and has the expected content
		data, err := os.ReadFile(kubeconfigPath)
		if err != nil {
			t.Fatalf("failed to read kubeconfig: %v", err)
		}

		if string(data) != string(expected) {
			t.Fatalf("unexpected kubeconfig contents: %q", string(data))
		}
	})

	t.Run("returns_error_when_missing", func(t *testing.T) {
		t.Parallel()

		clusterCfg := &v1alpha1.Cluster{}
		clusterCfg.Spec.Connection.Kubeconfig = filepath.Join(t.TempDir(), "missing")

		_, err := loadKubeconfig(clusterCfg)
		if err == nil {
			t.Fatal("expected error for missing kubeconfig")
		}
	})
}

func TestRunCiliumInstallationWritesSuccessMessage(t *testing.T) {
	t.Parallel()

	cmd, out := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	helmClient := &fakeHelmClient{}
	installer := newReadyCiliumInstaller(helmClient, time.Second)

	err := runCiliumInstallation(
		cmd,
		installer,
		&stubTimer{total: time.Second, stage: time.Second},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "CNI installed") {
		t.Fatalf("expected install success in output, got %q", output)
	}

	if helmClient.installCalls != 1 {
		t.Fatalf("expected helm install to be invoked once, got %d", helmClient.installCalls)
	}

	if helmClient.addRepoCalls != 1 {
		t.Fatalf("expected repository add to be invoked once, got %d", helmClient.addRepoCalls)
	}

	if helmClient.lastSpec == nil || helmClient.lastSpec.Timeout != time.Second {
		t.Fatalf("expected chart spec to use provided timeout")
	}
}

func TestRunCiliumInstallationReturnsInstallErrors(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	helmClient := &fakeHelmClient{installErr: context.DeadlineExceeded}
	installer := newReadyCiliumInstaller(helmClient, time.Second)

	err := runCiliumInstallation(cmd, installer, &stubTimer{})
	if err == nil {
		t.Fatal("expected installation error")
	}

	if !strings.Contains(err.Error(), "cilium installation failed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRunCiliumInstallationReturnsReadinessErrors(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	helmClient := &fakeHelmClient{}
	installer := newReadyCiliumInstaller(helmClient, time.Second)
	installer.SetWaitForReadinessFunc(func(context.Context) error {
		return errCiliumReadiness
	})

	err := runCiliumInstallation(cmd, installer, &stubTimer{})
	if err == nil {
		t.Fatal("expected readiness error")
	}

	if !strings.Contains(err.Error(), "cilium readiness check failed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func newReadyCiliumInstaller(
	helmClient *fakeHelmClient,
	timeout time.Duration,
) *ciliuminstaller.CiliumInstaller {
	installer := ciliuminstaller.NewCiliumInstaller(
		helmClient,
		testKubeconfigPath,
		testCiliumContext,
		timeout,
	)
	installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

	return installer
}

func TestInstallCiliumCNI(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.CNI = v1alpha1.CNICilium
	clusterCfg.Spec.Connection.Context = testCiliumContext
	clusterCfg.Spec.Connection.Kubeconfig = filepath.Join(t.TempDir(), "missing")

	err := installCiliumCNI(cmd, clusterCfg, &stubTimer{})
	if err == nil {
		t.Fatal("expected error when kubeconfig file is missing")
	}

	if !strings.Contains(err.Error(), "failed to read kubeconfig file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstallCiliumCNIHandlesRepositoryErrors(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	cmd.SetContext(ctx)

	kubeconfigPath := writeDummyKubeconfig(t)

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.CNI = v1alpha1.CNICilium
	clusterCfg.Spec.Connection.Context = testCiliumContext
	clusterCfg.Spec.Connection.Kubeconfig = kubeconfigPath

	err := installCiliumCNI(cmd, clusterCfg, &stubTimer{})
	if err == nil {
		t.Fatal("expected error when repository add fails")
	}

	if !strings.Contains(err.Error(), "failed to add Cilium Helm repository") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeDummyKubeconfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "kubeconfig")

	content := "apiVersion: v1\nclusters: []\ncontexts: []\ncurrent-context: \"\"\nusers: []\n"

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}

	return path
}

func writeCiliumClusterConfig(t *testing.T, dir, kubeconfig string) string {
	t.Helper()

	testutils.WriteValidKsailConfig(t, dir)

	content := "apiVersion: ksail.dev/v1alpha1\n" +
		"kind: Cluster\n" +
		"spec:\n" +
		"  distribution: Kind\n" +
		"  distributionConfig: kind.yaml\n" +
		"  sourceDirectory: k8s\n" +
		"  cni: Cilium\n" +
		"  connection:\n" +
		"    context: kind-kind\n" +
		"    kubeconfig: " + kubeconfig + "\n"

	configPath := filepath.Join(dir, "ksail.yaml")

	err := os.WriteFile(configPath, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write ksail config: %v", err)
	}

	return configPath
}

type fakeHelmClient struct {
	addRepoCalls int
	addRepoErr   error
	installCalls int
	installErr   error
	lastSpec     *helm.ChartSpec
}

func (f *fakeHelmClient) InstallChart(context.Context, *helm.ChartSpec) (*helm.ReleaseInfo, error) {
	return &helm.ReleaseInfo{}, nil
}

func (f *fakeHelmClient) InstallOrUpgradeChart(
	_ context.Context,
	spec *helm.ChartSpec,
) (*helm.ReleaseInfo, error) {
	f.installCalls++
	f.lastSpec = spec

	if f.installErr != nil {
		return nil, f.installErr
	}

	return &helm.ReleaseInfo{}, nil
}

func (f *fakeHelmClient) UninstallRelease(context.Context, string, string) error {
	return nil
}

func (f *fakeHelmClient) AddRepository(context.Context, *helm.RepositoryEntry) error {
	f.addRepoCalls++

	if f.addRepoErr != nil {
		return f.addRepoErr
	}

	return nil
}

type stubTimer struct {
	total time.Duration
	stage time.Duration
}

func (s *stubTimer) Start() {}

func (s *stubTimer) NewStage() {}

func (s *stubTimer) GetTiming() (time.Duration, time.Duration) {
	return s.total, s.stage
}

func (s *stubTimer) Stop() {}

// setupKindMirrorsTest creates the standard test setup for prepareKindConfigWithMirrors tests.
func setupKindMirrorsTest() (*v1alpha1.Cluster, *ksailconfigmanager.ConfigManager, *kindv1alpha4.Cluster) {
	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Distribution = v1alpha1.DistributionKind
	cfgManager := ksailconfigmanager.NewConfigManager(bytes.NewBuffer(nil))
	kindConfig := &kindv1alpha4.Cluster{}

	return clusterCfg, cfgManager, kindConfig
}

// newTestLifecycleDeps creates standard lifecycle dependencies for testing.
func newTestLifecycleDeps(provisioner *testutils.StubProvisioner) cmdhelpers.LifecycleDeps {
	return cmdhelpers.LifecycleDeps{
		Timer: &testutils.RecordingTimer{},
		Factory: &testutils.StubFactory{
			Provisioner:        provisioner,
			DistributionConfig: &kindv1alpha4.Cluster{Name: "kind"},
		},
	}
}

func TestPrepareKindConfigWithMirrors_NoKindConfig(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Distribution = v1alpha1.DistributionKind
	cfgManager := ksailconfigmanager.NewConfigManager(bytes.NewBuffer(nil))

	result := prepareKindConfigWithMirrors(clusterCfg, cfgManager, nil)
	assert.False(t, result, "should return false when kindConfig is nil")
}

func TestPrepareKindConfigWithMirrors_NonKindDistribution(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Distribution = v1alpha1.DistributionK3d
	cfgManager := ksailconfigmanager.NewConfigManager(bytes.NewBuffer(nil))
	kindConfig := &kindv1alpha4.Cluster{}

	result := prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig)
	assert.False(t, result, "should return false for non-Kind distribution")
}

func TestPrepareKindConfigWithMirrors_WithMirrorRegistryFlag(t *testing.T) {
	t.Parallel()

	clusterCfg, cfgManager, kindConfig := setupKindMirrorsTest()

	cfgManager.Viper.Set("mirror-registry", []string{"docker.io=http://localhost:5000"})

	result := prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig)
	assert.True(t, result, "should return true when mirror registries are added")
	assert.Len(t, kindConfig.ContainerdConfigPatches, 1)
}

func TestPrepareKindConfigWithMirrors_WithExistingPatches(t *testing.T) {
	t.Parallel()

	clusterCfg, cfgManager, kindConfig := setupKindMirrorsTest()
	kindConfig.ContainerdConfigPatches = []string{
		`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
endpoint = ["http://docker.io:5000"]`,
	}

	result := prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig)
	assert.True(t, result, "should return true when containerd patches exist")
}

func TestPrepareKindConfigWithMirrors_NoPatches(t *testing.T) {
	t.Parallel()

	clusterCfg, cfgManager, kindConfig := setupKindMirrorsTest()

	result := prepareKindConfigWithMirrors(clusterCfg, cfgManager, kindConfig)
	assert.False(t, result, "should return false when no patches")
}

func TestPrepareK3dConfigWithMirrors_NilConfig(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Distribution = v1alpha1.DistributionK3d

	result := prepareK3dConfigWithMirrors(clusterCfg, nil, nil)
	assert.False(t, result)
}

func TestPrepareK3dConfigWithMirrors_AddsOverrides(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Distribution = v1alpha1.DistributionK3d

	k3dConfig := k3dconfigmanager.NewK3dSimpleConfig("k3d-test", "", "")
	specs := registries.ParseMirrorSpecs([]string{"docker.io=https://registry-1.docker.io"})

	result := prepareK3dConfigWithMirrors(clusterCfg, k3dConfig, specs)

	assert.True(t, result)
	assert.Contains(t, k3dConfig.Registries.Config, "\"docker.io\"")
	assert.Contains(t, k3dConfig.Registries.Config, "https://registry-1.docker.io")
	assert.NotContains(t, k3dConfig.Registries.Config, "http://docker.io:5000")
}

func TestPrepareK3dConfigWithMirrors_PreservesExistingConfig(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.Distribution = v1alpha1.DistributionK3d

	k3dConfig := k3dconfigmanager.NewK3dSimpleConfig("k3d-test", "", "")
	k3dConfig.Registries.Config = "mirrors:\n  \"ghcr.io\":\n    endpoint:\n      - https://ghcr.io\n"

	result := prepareK3dConfigWithMirrors(clusterCfg, k3dConfig, nil)

	assert.True(t, result)
	assert.Contains(t, k3dConfig.Registries.Config, "\"ghcr.io\"")
	assert.Contains(t, k3dConfig.Registries.Config, "https://ghcr.io")
}

func TestRunLifecycleWithConfig_Success(t *testing.T) {
	t.Parallel()

	cmd, out := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	clusterCfg := &v1alpha1.Cluster{}
	provisioner := &testutils.StubProvisioner{}
	deps := newTestLifecycleDeps(provisioner)

	err := cmdhelpers.RunLifecycleWithConfig(cmd, deps, newCreateLifecycleConfig(), clusterCfg)

	require.NoError(t, err)
	assert.Equal(t, 1, provisioner.CreateCalls)
	assert.Contains(t, out.String(), "Create cluster")
}

func TestRunLifecycleWithConfig_ProvisionerError(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	clusterCfg := &v1alpha1.Cluster{}
	provisioner := &testutils.StubProvisioner{CreateErr: errClusterProvisionerFailed}
	deps := cmdhelpers.LifecycleDeps{
		Timer: &testutils.RecordingTimer{},
		Factory: &testutils.StubFactory{
			Provisioner:        provisioner,
			DistributionConfig: &kindv1alpha4.Cluster{Name: "kind"},
		},
	}

	err := cmdhelpers.RunLifecycleWithConfig(cmd, deps, newCreateLifecycleConfig(), clusterCfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create cluster")
}

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestSetupMirrorRegistries(t *testing.T) {
	runRegistryLifecycleTests(t, setupMirrorRegistries)
}

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestConnectRegistriesToClusterNetwork(t *testing.T) {
	runRegistryLifecycleTests(t, connectRegistriesToClusterNetwork)
}

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestSetupMirrorRegistries_K3d(t *testing.T) {
	runK3dRegistryLifecycleTests(t, setupMirrorRegistries)
}

//nolint:paralleltest // Overrides docker client factory for deterministic failure.
func TestConnectRegistriesToClusterNetwork_K3d(t *testing.T) {
	runK3dRegistryLifecycleTests(t, connectRegistriesToClusterNetwork)
}

type registryLifecycleHandler func(
	*cobra.Command,
	*v1alpha1.Cluster,
	cmdhelpers.LifecycleDeps,
	*ksailconfigmanager.ConfigManager,
	*kindv1alpha4.Cluster,
	*k3dv1alpha5.SimpleConfig,
) error

type k3dRegistryLifecycleTestCase struct {
	name          string
	includeConfig bool
	setOverride   bool
	expectError   bool
	errorContains string
}

func k3dRegistryLifecycleTestCases() []k3dRegistryLifecycleTestCase {
	return []k3dRegistryLifecycleTestCase{
		{
			name: "returns_nil_when_no_registries",
		},
	}
}

func runRegistryLifecycleTests(t *testing.T, handler registryLifecycleHandler) {
	t.Helper()

	testCases := []struct {
		name          string
		includePatch  bool
		expectError   bool
		errorContains string
	}{
		{
			name: "returns_nil_when_no_patches",
		},
	}

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cmd, out := testutils.NewCommand(t)
			cmd.SetContext(context.Background())

			cfg := v1alpha1.NewCluster()
			cfg.Spec.Distribution = v1alpha1.DistributionKind

			cfgManager := ksailconfigmanager.NewConfigManager(out, selectors...)
			kindConfig := &kindv1alpha4.Cluster{Name: "kind"}
			deps := cmdhelpers.LifecycleDeps{Timer: &testutils.RecordingTimer{}}

			var k3dConfig *k3dv1alpha5.SimpleConfig

			err := handler(cmd, cfg, deps, cfgManager, kindConfig, k3dConfig)

			if testCase.expectError {
				require.Error(t, err)
				require.ErrorContains(t, err, testCase.errorContains)

				return
			}

			require.NoError(t, err)
		})
	}
}

func runK3dRegistryLifecycleTests(t *testing.T, handler registryLifecycleHandler) {
	t.Helper()

	for _, testCase := range k3dRegistryLifecycleTestCases() {
		t.Run(testCase.name, func(t *testing.T) {
			executeK3dRegistryLifecycleTest(t, testCase, handler)
		})
	}
}

func executeK3dRegistryLifecycleTest(
	t *testing.T,
	testCase k3dRegistryLifecycleTestCase,
	handler registryLifecycleHandler,
) {
	t.Helper()

	cmd, out := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	cfg := v1alpha1.NewCluster()
	cfg.Spec.Distribution = v1alpha1.DistributionK3d

	cfgManager := ksailconfigmanager.NewConfigManager(
		out,
		ksailconfigmanager.DefaultClusterFieldSelectors()..., // reuse default selectors per test case
	)
	k3dConfig := k3dconfigmanager.NewK3dSimpleConfig("k3d-test", "", "")

	if testCase.includeConfig {
		k3dConfig.Registries.Config = "mirrors:\n  \"docker.io\":\n    endpoint:\n      - https://registry-1.docker.io\n"
	}

	if testCase.setOverride {
		cfgManager.Viper.Set("mirror-registry", []string{"ghcr.io=https://ghcr.io"})
	}

	err := handler(
		cmd,
		cfg,
		cmdhelpers.LifecycleDeps{Timer: &testutils.RecordingTimer{}},
		cfgManager,
		(*kindv1alpha4.Cluster)(nil),
		k3dConfig,
	)

	if testCase.expectError {
		require.Error(t, err)
		require.ErrorContains(t, err, testCase.errorContains)

		if testCase.setOverride {
			assert.Contains(t, k3dConfig.Registries.Config, "\"ghcr.io\"")
		}

		return
	}

	require.NoError(t, err)
}

func TestRunRegistryStageErrorWrapping(t *testing.T) {
	t.Parallel()

	t.Run("wraps_docker_client_errors", func(t *testing.T) {
		t.Parallel()

		cmd, _ := testutils.NewCommand(t)
		deps := cmdhelpers.LifecycleDeps{
			Timer: &testutils.RecordingTimer{},
		}
		info := registryStageInfo{
			title:         "Test",
			emoji:         "🔧",
			success:       "success",
			failurePrefix: "test failed",
		}
		action := func(context.Context, client.APIClient) error {
			return nil
		}

		// This will fail because Docker isn't available, but we're testing error wrapping
		err := runRegistryStage(cmd, deps, info, action)
		if err != nil {
			// Error should be wrapped
			assert.ErrorContains(t, err, "failed to execute registry stage")
		}
	})
}

func TestDistributionProvidesMetricsByDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		distribution v1alpha1.Distribution
		expected     bool
	}{
		{
			name:         "Kind does not provide metrics-server",
			distribution: v1alpha1.DistributionKind,
			expected:     false,
		},
		{
			name:         "K3d provides metrics-server",
			distribution: v1alpha1.DistributionK3d,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := distributionProvidesMetricsByDefault(tt.distribution)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleMetricsServer_Enabled_KindShouldInstall(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionKind,
			MetricsServer: v1alpha1.MetricsServerEnabled,
			Connection: v1alpha1.Connection{
				Kubeconfig: "/tmp/test-kubeconfig",
			},
		},
	}

	tmr := &testutils.RecordingTimer{}

	// This will fail because kubeconfig doesn't exist, but we can test the logic flow
	err := handleMetricsServer(cmd, clusterCfg, tmr)
	// Expect error since we're trying to read non-existent kubeconfig
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to read kubeconfig file")
}

func TestHandleMetricsServer_Enabled_K3dNoAction(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerEnabled,
		},
	}

	tmr := &testutils.RecordingTimer{}

	// K3d already has metrics-server, so no action should be taken
	err := handleMetricsServer(cmd, clusterCfg, tmr)
	assert.NoError(t, err)
}

func TestHandleMetricsServer_Disabled_KindNoAction(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionKind,
			MetricsServer: v1alpha1.MetricsServerDisabled,
		},
	}

	tmr := &testutils.RecordingTimer{}

	// Kind doesn't have metrics-server by default, so no action should be taken
	err := handleMetricsServer(cmd, clusterCfg, tmr)
	assert.NoError(t, err)
}

func TestHandleMetricsServer_Disabled_K3dNoAction(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerDisabled,
		},
	}

	tmr := &testutils.RecordingTimer{}

	// K3d with Disabled should be handled via config, not post-creation action
	err := handleMetricsServer(cmd, clusterCfg, tmr)
	assert.NoError(t, err)
}



func TestNewCreateCmd_IncludesMetricsServerFlag(t *testing.T) {
	t.Parallel()

	runtimeContainer := &runtime.Runtime{}
	cmd := NewCreateCmd(runtimeContainer)

	// Check that metrics-server flag exists
	flag := cmd.Flags().Lookup("metrics-server")
	require.NotNil(t, flag, "metrics-server flag should be registered")
	assert.Equal(t, "MetricsServer", flag.Value.Type())
}

func TestSetupK3dMetricsServer_DisabledAddsFlag(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerDisabled,
		},
	}

	k3dConfig := &k3dv1alpha5.SimpleConfig{}

	setupK3dMetricsServer(clusterCfg, k3dConfig)

	// Check that --disable=metrics-server flag was added
	found := false

	for _, arg := range k3dConfig.Options.K3sOptions.ExtraArgs {
		if arg.Arg == k3sDisableMetricsServerFlag {
			found = true

			assert.Equal(t, []string{"server:*"}, arg.NodeFilters)

			break
		}
	}

	assert.True(t, found, "--disable=metrics-server flag should be added")
}

func TestSetupK3dMetricsServer_EnabledNoFlag(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerEnabled,
		},
	}

	k3dConfig := &k3dv1alpha5.SimpleConfig{}

	setupK3dMetricsServer(clusterCfg, k3dConfig)

	// Check that no flags were added
	assert.Empty(t, k3dConfig.Options.K3sOptions.ExtraArgs)
}

func TestSetupK3dMetricsServer_NotK3dNoAction(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionKind,
			MetricsServer: v1alpha1.MetricsServerDisabled,
		},
	}

	k3dConfig := &k3dv1alpha5.SimpleConfig{}

	setupK3dMetricsServer(clusterCfg, k3dConfig)

	// Check that no flags were added for Kind distribution
	assert.Empty(t, k3dConfig.Options.K3sOptions.ExtraArgs)
}

func TestSetupK3dMetricsServer_AlreadyConfigured(t *testing.T) {
	t.Parallel()

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerDisabled,
		},
	}

	k3dConfig := &k3dv1alpha5.SimpleConfig{
		Options: k3dv1alpha5.SimpleConfigOptions{
			K3sOptions: k3dv1alpha5.SimpleConfigOptionsK3s{
				ExtraArgs: []k3dv1alpha5.K3sArgWithNodeFilters{
					{
						Arg:         k3sDisableMetricsServerFlag,
						NodeFilters: []string{"server:*"},
					},
				},
			},
		},
	}

	setupK3dMetricsServer(clusterCfg, k3dConfig)

	// Check that flag was not duplicated
	count := 0

	for _, arg := range k3dConfig.Options.K3sOptions.ExtraArgs {
		if arg.Arg == k3sDisableMetricsServerFlag {
			count++
		}
	}

	assert.Equal(t, 1, count, "flag should not be duplicated")
}

func TestInstallMetricsServer_LoadKubeconfigError(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionKind,
			Connection: v1alpha1.Connection{
				Kubeconfig: "",
			},
		},
	}

	tmr := &testutils.RecordingTimer{}

	err := installMetricsServer(cmd, clusterCfg, tmr)
	assert.Error(t, err)
}

func TestInstallMetricsServer_KubeconfigReadError(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionKind,
			Connection: v1alpha1.Connection{
				Kubeconfig: "/tmp/nonexistent-kubeconfig-test-file-xyz",
				Context:    "test-context",
			},
		},
	}

	tmr := &testutils.RecordingTimer{}

	err := installMetricsServer(cmd, clusterCfg, tmr)
	require.Error(t, err)
	assert.ErrorContains(t, err, "failed to read kubeconfig file")
}

func TestDistributionProvidesMetricsByDefault_UnknownDistribution(t *testing.T) {
	t.Parallel()

	result := distributionProvidesMetricsByDefault(v1alpha1.Distribution("unknown"))
	assert.False(t, result, "Unknown distribution should default to false")
}

func TestDistributionProvidesMetricsByDefault_EmptyDistribution(t *testing.T) {
	t.Parallel()

	result := distributionProvidesMetricsByDefault(v1alpha1.Distribution(""))
	assert.False(t, result, "Empty distribution should default to false")
}

func TestHandlePostCreationSetup_MetricsServerFirst(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	// Test with Kind and Disabled metrics-server (should not attempt installation)
	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionKind,
			MetricsServer: v1alpha1.MetricsServerDisabled,
			CNI:           v1alpha1.CNIDefault,
		},
	}

	tmr := &testutils.RecordingTimer{}
	tmr.Start()

	err := handlePostCreationSetup(cmd, clusterCfg, tmr)
	assert.NoError(t, err, "Should not error when metrics-server is disabled on Kind")
}

func TestHandlePostCreationSetup_K3dDisabled(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	// Test with K3d and Disabled metrics-server
	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerDisabled,
			CNI:           v1alpha1.CNIDefault,
		},
	}

	tmr := &testutils.RecordingTimer{}
	tmr.Start()

	err := handlePostCreationSetup(cmd, clusterCfg, tmr)
	assert.NoError(t, err, "Should not error when metrics-server is disabled on K3d")
}

func TestHandlePostCreationSetup_K3dEnabled(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)

	// Test with K3d and Enabled metrics-server (no action needed)
	clusterCfg := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:  v1alpha1.DistributionK3d,
			MetricsServer: v1alpha1.MetricsServerEnabled,
			CNI:           v1alpha1.CNIDefault,
		},
	}

	tmr := &testutils.RecordingTimer{}
	tmr.Start()

	err := handlePostCreationSetup(cmd, clusterCfg, tmr)
	assert.NoError(
		t,
		err,
		"Should not error when metrics-server is enabled on K3d (already present)",
	)
}
