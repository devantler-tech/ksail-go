package gen //nolint:testpackage // Needs access to unexported helpers for coverage instrumentation.

import (
	"bytes"
	"strings"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

func TestNewGenCmdRegistersAllResourceCommands(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewGenCmd(rt)

	expectedSubcommands := []string{
		"clusterrole",
		"clusterrolebinding",
		"configmap",
		"cronjob",
		"deployment",
		"ingress",
		"job",
		"namespace",
		"poddisruptionbudget",
		"priorityclass",
		"quota",
		"role",
		"rolebinding",
		"secret",
		"service",
		"serviceaccount",
		"token",
	}

	for _, expectedName := range expectedSubcommands {
		t.Run(expectedName, func(t *testing.T) {
			t.Parallel()

			found := false

			for _, subCmd := range cmd.Commands() {
				if subCmd.Name() == expectedName {
					found = true

					break
				}
			}

			if !found {
				t.Errorf("expected gen command to include %q subcommand", expectedName)
			}
		})
	}
}

func TestGenCommandRunEDisplaysHelp(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewGenCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected executing gen command without subcommand to succeed, got %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Generate Kubernetes resource manifests") {
		t.Errorf("expected help output to contain description, got %q", output)
	}
}

func TestGenCommandMetadata(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	cmd := NewGenCmd(rt)

	if cmd.Use != "gen" {
		t.Errorf("expected Use to be 'gen', got %q", cmd.Use)
	}

	if cmd.Short != "Generate Kubernetes resource manifests" {
		t.Errorf("expected Short description, got %q", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "kubectl create") {
		t.Errorf("expected Long description to mention kubectl create, got %q", cmd.Long)
	}

	if !strings.Contains(cmd.Long, "--dry-run=client") {
		t.Errorf("expected Long description to mention --dry-run=client, got %q", cmd.Long)
	}
}

func newTestRuntime() *runtime.Runtime {
	return runtime.NewRuntime()
}
