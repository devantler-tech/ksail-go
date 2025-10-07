package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	internaltestutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { internaltestutils.RunTestMainWithSnapshotCleanup(m) }

type recordingListFactory struct {
	provisioner clusterprovisioner.ClusterProvisioner
	err         error
	callCount   int
	captured    []*v1alpha1.Cluster
}

//nolint:ireturn // Test doubles satisfy interface contract.
func (f *recordingListFactory) Create(
	_ context.Context,
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, any, error) {
	f.callCount++
	f.captured = append(f.captured, cluster)
	if f.err != nil {
		return nil, nil, f.err
	}

	return f.provisioner, nil, nil
}

type recordingListProvisioner struct {
	listResult []string
	listErr    error
	listCalls  int
}

func (p *recordingListProvisioner) Create(context.Context, string) error { return nil }
func (p *recordingListProvisioner) Delete(context.Context, string) error { return nil }
func (p *recordingListProvisioner) Start(context.Context, string) error  { return nil }
func (p *recordingListProvisioner) Stop(context.Context, string) error   { return nil }

func (p *recordingListProvisioner) List(context.Context) ([]string, error) {
	p.listCalls++
	if p.listErr != nil {
		return nil, p.listErr
	}

	clone := append([]string(nil), p.listResult...)

	return clone, nil
}

func (p *recordingListProvisioner) Exists(context.Context, string) (bool, error) {
	return false, nil
}

func createConfigManagerWithFile(t *testing.T, writer io.Writer) *configmanager.ConfigManager {
	t.Helper()

	selectors := configmanager.DefaultClusterFieldSelectors()
	cfgManager := configmanager.NewConfigManager(writer, selectors...)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	cfgManager.Viper.SetConfigFile(filepath.Join(tempDir, "ksail.yaml"))

	return cfgManager
}

const ignoredConfigValue = "ignored"

func TestHandleListRunE_ReturnsErrorWhenConfigLoadFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer(t)

	tmpDir := t.TempDir()

	badConfigPath := filepath.Join(tmpDir, "ksail.yaml")

	err := os.WriteFile(badConfigPath, []byte(": invalid yaml"), 0o600)
	if err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := configmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badConfigPath)

	deps := ListDeps{Factory: fakeFactory{}}

	err = HandleListRunE(cmd, cfgManager, deps)
	if err == nil {
		t.Fatal("expected configuration load error, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "failed to load configuration") {
		t.Fatalf("expected error to mention configuration load failure, got %q", message)
	}

	if !strings.Contains(message, "failed to read config file") {
		t.Fatalf("expected config read failure to be reported, got %q", message)
	}
}

func TestHandleListRunE(t *testing.T) {
	t.Parallel()

	t.Run("success displays clusters", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer(t)
		cfgManager := createConfigManagerWithFile(t, out)

		provisioner := &recordingListProvisioner{listResult: []string{"alpha"}}
		factory := &recordingListFactory{provisioner: provisioner}

		err := HandleListRunE(cmd, cfgManager, ListDeps{Factory: factory})
		require.NoError(t, err)

		require.Equal(t, 1, factory.callCount)
		require.Equal(t, 1, provisioner.listCalls)

		snaps.MatchSnapshot(t, out.String())
	})

	t.Run("list failure wraps error", func(t *testing.T) {
		t.Parallel()

		cmd, _ := newCommandWithBuffer(t)
		cfgManager := createConfigManagerWithFile(t, io.Discard)

		provisioner := &recordingListProvisioner{listErr: context.DeadlineExceeded}
		factory := &recordingListFactory{provisioner: provisioner}

		err := HandleListRunE(cmd, cfgManager, ListDeps{Factory: factory})
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to list clusters")
	})
}

func TestListClusters_ReturnsErrorWhenFactoryFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer(t)

	cfgManager := configmanager.NewConfigManager(io.Discard)
	cfgManager.Config.Spec.Distribution = v1alpha1.Distribution("Unsupported")
	cfgManager.Config.Spec.DistributionConfig = ignoredConfigValue
	cfgManager.Config.Spec.Connection.Kubeconfig = ignoredConfigValue
	cfgManager.Config.Spec.SourceDirectory = ignoredConfigValue

	deps := ListDeps{Factory: fakeFactory{err: clusterprovisioner.ErrUnsupportedDistribution}}

	err := listClusters(cfgManager, deps, cmd)
	if err == nil {
		t.Fatal("expected resolver failure, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "failed to resolve cluster provisioner") {
		t.Fatalf("expected factory error to be wrapped, got %q", message)
	}

	if !strings.Contains(message, "unsupported distribution") {
		t.Fatalf("expected unsupported distribution to be reported, got %q", message)
	}
}

func TestListClusters_ListFailure(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer(t)
	cfgManager := configmanager.NewConfigManager(io.Discard)
	cfgManager.Config.Spec.Distribution = v1alpha1.DistributionKind

	provisioner := &recordingListProvisioner{listErr: context.DeadlineExceeded}
	factory := &recordingListFactory{provisioner: provisioner}

	err := listClusters(cfgManager, ListDeps{Factory: factory}, cmd)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list clusters")
	require.Equal(t, 1, provisioner.listCalls)
}

func TestListClusters_AllFlagTriggersAdditionalDistribution(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer(t)
	cfgManager := createConfigManagerWithFile(t, io.Discard)
	require.NoError(t, cfgManager.LoadConfigSilent())
	cfgManager.Viper.Set(allFlag, true)

	provisioner := &recordingListProvisioner{listResult: []string{"kind-primary"}}
	factory := &recordingListFactory{provisioner: provisioner}

	err := listClusters(cfgManager, ListDeps{Factory: factory}, cmd)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list clusters for distribution K3d")
	require.Equal(t, 1, provisioner.listCalls)
}

func TestDisplayClusterList(t *testing.T) {
	t.Parallel()

	t.Run("no clusters writes activity message", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer(t)

		displayClusterList(v1alpha1.DistributionKind, nil, cmd, false)

		got := out.String()
		want := "► no clusters found\n"

		if got != want {
			t.Fatalf("expected activity notification for empty list. want %q, got %q", want, got)
		}
	})

	t.Run("clusters are formatted per distribution", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer(t)

		displayClusterList(v1alpha1.DistributionK3d, []string{"alpha", "beta"}, cmd, true)

		got := out.String()
		want := "k3d: alpha, beta\n"

		if got != want {
			t.Fatalf("expected formatted cluster list. want %q, got %q", want, got)
		}
	})
}

func TestCloneClusterForDistribution(t *testing.T) {
	t.Parallel()

	t.Run("nil original returns nil", func(t *testing.T) {
		t.Parallel()

		clone := cloneClusterForDistribution(nil, v1alpha1.DistributionKind)
		require.Nil(t, clone)
	})

	t.Run("distribution and config path updated", func(t *testing.T) {
		t.Parallel()

		original := &v1alpha1.Cluster{}
		original.Spec.Distribution = v1alpha1.DistributionK3d
		original.Spec.DistributionConfig = "custom.yaml"

		clone := cloneClusterForDistribution(original, v1alpha1.DistributionKind)

		require.NotNil(t, clone)
		require.Equal(t, v1alpha1.DistributionKind, clone.Spec.Distribution)
		require.Equal(t, "kind.yaml", clone.Spec.DistributionConfig)
		require.Equal(t, v1alpha1.DistributionK3d, original.Spec.Distribution)
		require.Equal(t, "custom.yaml", original.Spec.DistributionConfig)
	})
}

func TestDefaultDistributionConfigPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		distribution v1alpha1.Distribution
		expected     string
	}{
		{name: "kind", distribution: v1alpha1.DistributionKind, expected: "kind.yaml"},
		{name: "k3d", distribution: v1alpha1.DistributionK3d, expected: "k3d.yaml"},
		{name: "unknown", distribution: "other", expected: "kind.yaml"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := defaultDistributionConfigPath(tc.distribution)
			require.Equal(t, tc.expected, actual)
		})
	}
}

//nolint:paralleltest // Uses t.Chdir for snapshot setup.
func TestNewListCmd_RunESuccess(t *testing.T) {
	factory := &recordingListFactory{}
	provisioner := &recordingListProvisioner{listResult: []string{"kind-mgmt"}}
	factory.provisioner = provisioner

	runtimeContainer := runtime.New(func(injector do.Injector) error {
		do.Provide(injector, func(do.Injector) (clusterprovisioner.Factory, error) {
			return factory, nil
		})

		return nil
	})

	cmd := NewListCmd(runtimeContainer)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)
	t.Chdir(tempDir)

	err := cmd.Execute()
	require.NoError(t, err)
	require.Equal(t, 1, factory.callCount)
	require.Equal(t, 1, provisioner.listCalls)

	snaps.MatchSnapshot(t, out.String())
}

func TestBindAllFlagBindsViperState(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "list"}
	cfgManager := configmanager.NewConfigManager(io.Discard)
	bindAllFlag(cmd, cfgManager)

	err := cmd.Flags().Set(allFlag, "true")
	if err != nil {
		t.Fatalf("failed to set all flag: %v", err)
	}

	if !cfgManager.Viper.GetBool(allFlag) {
		t.Fatal("expected Viper binding to reflect updated flag state")
	}
}

type fakeFactory struct {
	provisioner clusterprovisioner.ClusterProvisioner
	err         error
}

//nolint:ireturn // Tests rely on returning the interface to satisfy ListDeps contract.
func (f fakeFactory) Create(
	_ context.Context,
	_ *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, any, error) {
	return f.provisioner, nil, f.err
}

func newCommandWithBuffer(t *testing.T) (*cobra.Command, *bytes.Buffer) {
	t.Helper()

	tcmd := &cobra.Command{}

	var out bytes.Buffer
	tcmd.SetOut(&out)
	tcmd.SetErr(&out)

	return tcmd, &out
}
