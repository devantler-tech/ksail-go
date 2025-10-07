package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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

// setupCommandTest creates common test fixtures for command tests.
func setupCommandTest(t *testing.T) (
	*cobra.Command,
	*bytes.Buffer,
	*testutils.RecordingTimer,
	*testutils.StubFactory,
	*testutils.StubProvisioner,
	*ksailconfigmanager.ConfigManager,
) {
	t.Helper()

	cmd, out := testutils.NewCommand(t)
	timerStub := &testutils.RecordingTimer{}
	provisioner := &testutils.StubProvisioner{}
	factory := &testutils.StubFactory{
		Provisioner:        provisioner,
		DistributionConfig: &v1alpha4.Cluster{Name: "kind"},
	}
	cfgManager := testutils.CreateConfigManager(t, out)

	return cmd, out, timerStub, factory, provisioner, cfgManager
}
