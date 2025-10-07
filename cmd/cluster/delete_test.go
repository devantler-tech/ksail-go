package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
"bytes"
"strings"
"testing"

runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

// TestHandleDeleteRunE_WrapperDelegation tests that the deprecated HandleDeleteRunE properly delegates to the shared lifecycle handler.
func TestHandleDeleteRunE_WrapperDelegation(t *testing.T) {
t.Parallel()

cmd, out, timerStub, factory, provisioner, cfgManager := setupCommandTest(t)

err := HandleDeleteRunE(cmd, cfgManager, DeleteDeps{Timer: timerStub, Factory: factory})
if err != nil {
t.Fatalf("expected success, got %v", err)
}

// Verify the wrapper correctly invoked the provisioner's Delete method
if provisioner.DeleteCalls != 1 {
t.Fatalf("expected Delete to be called once, got %d", provisioner.DeleteCalls)
}

// Verify output contains delete-specific messages
if !strings.Contains(out.String(), "Delete cluster...") {
t.Fatalf("expected 'Delete cluster...' in output, got %q", out.String())
}
}

func TestNewDeleteCmd(t *testing.T) {
t.Parallel()

runtimeContainer := runtime.NewRuntime()
cmd := NewDeleteCmd(runtimeContainer)

if cmd.Use != "delete" {
t.Fatalf("expected Use to be 'delete', got %q", cmd.Use)
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
