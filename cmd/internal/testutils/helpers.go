package testutils

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

// CreateDefaultConfigManager creates a standard config manager for cmd tests that passes KSail validation.
func CreateDefaultConfigManager() *configmanager.ConfigManager {
	return configmanager.NewConfigManager(
		io.Discard,
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.APIVersion },
			Description:  "API version",
			DefaultValue: "ksail.dev/v1alpha1",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Kind },
			Description:  "Resource kind",
			DefaultValue: "Cluster",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to use",
			DefaultValue: v1alpha1.DistributionKind,
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			Description:  "Path to distribution configuration file",
			DefaultValue: "kind.yaml",
		},
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context name",
			DefaultValue: "kind-kind", // Using default pattern that validator expects
		},
	)
}

// SetupCommandWithOutput creates a standard cobra command with output buffer for cmd tests.
func SetupCommandWithOutput() (*cobra.Command, *bytes.Buffer) {
	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	return testCmd, &out
}

// SimpleCommandTestData holds test data for simple command testing.
type SimpleCommandTestData struct {
	CommandName   string
	NewCommand    func() *cobra.Command
	ExpectedUse   string
	ExpectedShort string
}

// TestSimpleCommandCreation tests command creation with common pattern.
func TestSimpleCommandCreation(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	cmd := data.NewCommand()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if data.ExpectedUse != "" && cmd.Use != data.ExpectedUse {
		t.Fatalf("expected Use to be %q, got %q", data.ExpectedUse, cmd.Use)
	}

	if data.ExpectedShort != "" && cmd.Short != data.ExpectedShort {
		t.Fatalf("expected Short description to be %q, got %q", data.ExpectedShort, cmd.Short)
	}
}

// TestSimpleCommandExecution tests command execution with common pattern.
func TestSimpleCommandExecution(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	var out bytes.Buffer

	cmd := data.NewCommand()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

// TestSimpleCommandHelp tests command help output with common pattern.
func TestSimpleCommandHelp(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	var out bytes.Buffer

	cmd := data.NewCommand()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

// TestCmdExecuteInCleanDir executes a command in a temporary directory with no ksail.yaml file
// and validates that it returns a configuration validation error.
func TestCmdExecuteInCleanDir(t *testing.T, cmdFactory func() *cobra.Command, cmdName string) {
	t.Helper()

	// Create a temporary directory to ensure no ksail.yaml exists
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	t.Chdir(tempDir)

	defer func() {
		t.Chdir(originalDir)
	}()

	cmd := cmdFactory()
	err = cmd.Execute()

	// Expect a validation error because no valid configuration is provided
	if err == nil {
		t.Fatalf("expected validation error for %s command, got nil", cmdName)
	}

	if !strings.Contains(err.Error(), "configuration validation failed") {
		t.Fatalf(
			"expected 'configuration validation failed' in error for %s command, got: %v",
			cmdName,
			err,
		)
	}
}
