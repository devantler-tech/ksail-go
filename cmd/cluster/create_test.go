package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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

func TestHandleCreateRunE(t *testing.T) {
	t.Parallel()

	t.Run("installs_cilium_when_configured", func(t *testing.T) {
		t.Parallel()

		cmd, out := testutils.NewCommand(t)

		tempDir := t.TempDir()
		cfgPath := writeCiliumClusterConfig(t, tempDir, "./missing-kubeconfig")

		cfgManager := ksailconfigmanager.NewConfigManager(
			out,
			ksailconfigmanager.DefaultClusterFieldSelectors()...,
		)
		cfgManager.Viper.SetConfigFile(cfgPath)

		provisioner := &testutils.StubProvisioner{}
		deps := shared.LifecycleDeps{
			Timer: &testutils.RecordingTimer{},
			Factory: &testutils.StubFactory{
				Provisioner:        provisioner,
				DistributionConfig: &kindv1alpha4.Cluster{Name: "kind"},
			},
		}

		err := handleCreateRunE(cmd, cfgManager, deps)
		if err == nil {
			t.Fatal("expected error when kubeconfig is missing")
		}

		if !strings.Contains(err.Error(), "failed to install Cilium CNI") {
			t.Fatalf("unexpected error message: %v", err)
		}

		if provisioner.CreateCalls != 1 {
			t.Fatalf("expected provisioner create to be invoked, got %d", provisioner.CreateCalls)
		}
	})
}

func TestGetCiliumInstallTimeout(t *testing.T) {
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

			result := getCiliumInstallTimeout(cfg)
			if result != test.expected {
				t.Fatalf("expected %v, got %v", test.expected, result)
			}
		})
	}
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

		path, data, err := loadKubeconfig(clusterCfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if path != kubeconfigPath {
			t.Fatalf("expected path %q, got %q", kubeconfigPath, path)
		}

		if string(data) != string(expected) {
			t.Fatalf("unexpected kubeconfig contents: %q", string(data))
		}
	})

	t.Run("returns_error_when_missing", func(t *testing.T) {
		t.Parallel()

		clusterCfg := &v1alpha1.Cluster{}
		clusterCfg.Spec.Connection.Kubeconfig = filepath.Join(t.TempDir(), "missing")

		_, _, err := loadKubeconfig(clusterCfg)
		if err == nil {
			t.Fatal("expected error for missing kubeconfig")
		}
	})
}

func TestRunCiliumInstallation(t *testing.T) {
	t.Parallel()

	t.Run("writes_success_message", func(t *testing.T) {
		t.Parallel()

		cmd, out := testutils.NewCommand(t)
		cmd.SetContext(context.Background())

		helmClient := &fakeHelmClient{}
		installer := ciliuminstaller.NewCiliumInstaller(
			helmClient,
			"kubeconfig",
			"kind-kind",
			time.Second,
		)
		installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

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
			t.Fatalf("expected success message in output, got %q", output)
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
	})

	t.Run("returns_error_when_install_fails", func(t *testing.T) {
		t.Parallel()

		cmd, _ := testutils.NewCommand(t)
		cmd.SetContext(context.Background())

		helmClient := &fakeHelmClient{installErr: context.DeadlineExceeded}
		installer := ciliuminstaller.NewCiliumInstaller(
			helmClient,
			"kubeconfig",
			"kind-kind",
			time.Second,
		)
		installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

		err := runCiliumInstallation(cmd, installer, &stubTimer{})
		if err == nil {
			t.Fatal("expected installation error")
		}

		if !strings.Contains(err.Error(), "cilium installation failed") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}

func TestInstallCiliumCNI(t *testing.T) {
	t.Parallel()

	cmd, _ := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	clusterCfg := &v1alpha1.Cluster{}
	clusterCfg.Spec.CNI = v1alpha1.CNICilium
	clusterCfg.Spec.Connection.Context = "kind-kind"
	clusterCfg.Spec.Connection.Kubeconfig = filepath.Join(t.TempDir(), "missing")

	err := installCiliumCNI(cmd, clusterCfg, &stubTimer{})
	if err == nil {
		t.Fatal("expected error when kubeconfig file is missing")
	}

	if !strings.Contains(err.Error(), "failed to read kubeconfig file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeCiliumClusterConfig(t *testing.T, dir, kubeconfig string) string {
	t.Helper()

	cmdtestutils.WriteValidKsailConfig(t, dir)

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
