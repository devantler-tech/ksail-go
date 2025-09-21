package cmd_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInitCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewInitCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "init" {
		t.Fatalf("expected Use to be 'init', got %q", cmd.Use)
	}

	if cmd.Short != "Initialize a new project" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestInitCmdExecute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "init",
		NewCommand:  cmd.NewInitCmd,
	})
}

func TestInitCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		NewCommand: cmd.NewInitCmd,
	})
}

func TestInitCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewInitCmd()

	// Act & Assert
	distributionFlag := cmd.Flags().Lookup("distribution")
	if distributionFlag == nil {
		t.Fatal("expected distribution flag to exist")
	}

	// Verify that CLI flags show appropriate defaults for better UX
	// Distribution should show its default value
	if distributionFlag.DefValue != "Kind" {
		t.Fatalf(
			"expected distribution default to be 'Kind' for help display, got %q",
			distributionFlag.DefValue,
		)
	}

	sourceDirectoryFlag := cmd.Flags().Lookup("source-directory")
	if sourceDirectoryFlag == nil {
		t.Fatal("expected source-directory flag to exist")
	}

	// Source directory should show its default value
	if sourceDirectoryFlag.DefValue != "k8s" {
		t.Fatalf(
			"expected source-directory default to be 'k8s' for help display, got %q",
			sourceDirectoryFlag.DefValue,
		)
	}
}

//nolint:paralleltest // Cannot use t.Parallel() with t.Chdir()
func TestHandleInitRunE(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var out bytes.Buffer

		testCmd := &cobra.Command{}
		testCmd.SetOut(&out)

		// Create manager with the same field selectors as the real init command
		fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
			cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to use"),
			cmdhelpers.StandardSourceDirectoryFieldSelector(),
		}
		manager := configmanager.NewConfigManager(fieldSelectors...)

		// Create a temporary directory for testing
		tempDir := t.TempDir()

		// Change to temp directory for the test
		origDir, _ := os.Getwd()

		t.Chdir(tempDir)
		defer t.Chdir(origDir)

		err := cmd.HandleInitRunE(testCmd, manager, []string{})

		require.NoError(t, err)
		assert.Contains(t, out.String(), "✔ project initialized successfully")
		assert.Contains(t, out.String(), "► Distribution:")
		assert.Contains(t, out.String(), "► Source directory:")

		// Verify that scaffolder created the expected files
		assert.FileExists(t, "ksail.yaml")
		assert.FileExists(t, "kind.yaml")
		assert.DirExists(t, "k8s")
		assert.FileExists(t, "k8s/kustomization.yaml")
	})
}
