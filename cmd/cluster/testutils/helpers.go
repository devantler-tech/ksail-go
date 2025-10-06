// Package testutils provides shared helpers for cluster command tests.
package testutils

import (
	"bytes"
	"strings"
	"testing"

	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/spf13/cobra"
)

// SetupValidWorkingDir creates a valid KSail configuration in a temporary directory and switches to it.
// The returned cleanup function restores the original working directory.
func SetupValidWorkingDir(t *testing.T) func() {
	t.Helper()

	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	t.Chdir(tempDir)

	return func() {}
}

// RunValidationErrorTest executes the provided command factory in an empty directory and validates error output.
func RunValidationErrorTest(
	t *testing.T,
	commandFactory func() *cobra.Command,
	expectedSubstrings ...string,
) {
	t.Helper()

	tempDir := t.TempDir()

	t.Chdir(tempDir)

	command := commandFactory()

	var out bytes.Buffer
	command.SetOut(&out)
	command.SetErr(&out)

	if command.RunE == nil {
		t.Fatal("command RunE must not be nil")
	}

	err := command.RunE(command, nil)
	if err == nil {
		t.Fatal("expected error but got nil")
	}

	message := err.Error()
	requiredParts := append(
		[]string{
			"failed to load cluster configuration",
			helpers.ErrConfigurationValidationFailed.Error(),
		},
		expectedSubstrings...,
	)

	for _, substring := range requiredParts {
		if !strings.Contains(message, substring) {
			t.Fatalf("expected error message to contain %q, got %q", substring, message)
		}
	}
}
