package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
"bytes"
"strings"
"testing"

"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
runtime "github.com/devantler-tech/ksail-go/pkg/di"
"github.com/spf13/cobra"
"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// TestHandleCreateRunE_WrapperDelegation tests that the deprecated HandleCreateRunE properly delegates to the shared lifecycle handler.
func TestHandleCreateRunE_WrapperDelegation(t *testing.T) {
t.Parallel()

cmd, out, timerStub, factory, provisioner, cfgManager := setupCommandTest(t)

err := HandleCreateRunE(cmd, cfgManager, CreateDeps{Timer: timerStub, Factory: factory})
if err != nil {
t.Fatalf("expected success, got %v", err)
}

// Verify the wrapper correctly invoked the provisioner's Create method
if provisioner.CreateCalls != 1 {
t.Fatalf("expected Create to be called once, got %d", provisioner.CreateCalls)
}

// Verify output contains create-specific messages
if !strings.Contains(out.String(), "Create cluster...") {
t.Fatalf("expected 'Create cluster...' in output, got %q", out.String())
}
}

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
