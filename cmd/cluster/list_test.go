package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"io"
	"strings"
	"testing"

	commandutils "github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

func TestHandleListRunE_ReturnsErrorWhenConfigLoadFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer()
	utils := &commandutils.CommandUtils{
		ConfigManager: configmanager.NewConfigManager(io.Discard),
	}

	err := HandleListRunE(cmd, utils, nil)
	if err == nil {
		t.Fatal("expected configuration load error, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "failed to load configuration") {
		t.Fatalf("expected error to mention configuration load failure, got %q", message)
	}
}

func TestListClusters_ReturnsErrorWhenResolverFails(t *testing.T) {
	t.Parallel()

	cmd, _ := newCommandWithBuffer()

	cfgManager := configmanager.NewConfigManager(io.Discard)
	cfgManager.Config.Spec.Distribution = v1alpha1.Distribution("Unsupported")
	cfgManager.Config.Spec.DistributionConfig = "ignored"
	cfgManager.Config.Spec.Connection.Kubeconfig = "ignored"

	resolver, err := di.NewResolver(cfgManager.Config)
	if err != nil {
		t.Fatalf("expected resolver creation to succeed, got %v", err)
	}

	utils := &commandutils.CommandUtils{
		ConfigManager: cfgManager,
		Resolver:      resolver,
	}

	err = listClusters(utils, cmd)
	if err == nil {
		t.Fatal("expected resolver failure, got nil")
	}

	message := err.Error()
	if !strings.Contains(message, "failed to resolve dependencies") {
		t.Fatalf("expected resolver error to be wrapped, got %q", message)
	}
}

func TestDisplayClusterList(t *testing.T) {
	t.Parallel()

	t.Run("no clusters writes activity message", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer()

		displayClusterList(v1alpha1.DistributionKind, nil, cmd)

		got := out.String()
		want := "► no clusters found\n"

		if got != want {
			t.Fatalf("expected activity notification for empty list. want %q, got %q", want, got)
		}
	})

	t.Run("clusters are formatted per distribution", func(t *testing.T) {
		t.Parallel()

		cmd, out := newCommandWithBuffer()

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
	utils := &commandutils.CommandUtils{ConfigManager: cfgManager}

	bindAllFlag(cmd, utils)

	if err := cmd.Flags().Set(allFlag, "true"); err != nil {
		t.Fatalf("failed to set all flag: %v", err)
	}

	if !utils.ConfigManager.Viper.GetBool(allFlag) {
		t.Fatal("expected Viper binding to reflect updated flag state")
	}
}

func newCommandWithBuffer() (*cobra.Command, *bytes.Buffer) {
	tcmd := &cobra.Command{}
	var out bytes.Buffer
	tcmd.SetOut(&out)
	tcmd.SetErr(&out)

	return tcmd, &out
}
