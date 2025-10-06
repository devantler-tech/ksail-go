package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/spf13/cobra"
)

func TestHandleListRunE_ReturnsErrorWhenConfigLoadFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer(t)

	tmpDir := t.TempDir()
	badConfigPath := filepath.Join(tmpDir, "ksail.yaml")
	if err := os.WriteFile(badConfigPath, []byte(": invalid yaml"), 0o600); err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}

	cfgManager := configmanager.NewConfigManager(io.Discard)
	cfgManager.Viper.SetConfigFile(badConfigPath)

	deps := ListDeps{Factory: fakeFactory{}}

	err := HandleListRunE(cmd, cfgManager, deps)
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

func TestListClusters_ReturnsErrorWhenFactoryFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer(t)

	cfgManager := configmanager.NewConfigManager(io.Discard)
	cfgManager.Config.Spec.Distribution = v1alpha1.Distribution("Unsupported")
	cfgManager.Config.Spec.DistributionConfig = "ignored"
	cfgManager.Config.Spec.Connection.Kubeconfig = "ignored"
	cfgManager.Config.Spec.SourceDirectory = "ignored"

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

func TestDisplayClusterList(t *testing.T) {
	t.Parallel()

	t.Run("no clusters writes activity message", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer(t)

		displayClusterList(v1alpha1.DistributionKind, nil, cmd)

		got := out.String()
		want := "► no clusters found\n"

		if got != want {
			t.Fatalf("expected activity notification for empty list. want %q, got %q", want, got)
		}
	})

	t.Run("clusters are formatted per distribution", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer(t)

		displayClusterList(v1alpha1.DistributionK3d, []string{"alpha", "beta"}, cmd)

		got := out.String()
		want := "K3d: alpha, beta\n"

		if got != want {
			t.Fatalf("expected formatted cluster list. want %q, got %q", want, got)
		}
	})
}

func TestBindAllFlagBindsViperState(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "list"}
	cfgManager := configmanager.NewConfigManager(io.Discard)
	bindAllFlag(cmd, cfgManager)

	if err := cmd.Flags().Set(allFlag, "true"); err != nil {
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
