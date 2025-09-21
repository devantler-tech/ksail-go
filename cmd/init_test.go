package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/gkampitakis/go-snaps/snaps"
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

	var out bytes.Buffer

	// Create a temp directory for this test
	tempDir := t.TempDir()

	// Create a custom cobra command that uses the temp directory
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project",
		Long:  "Initialize a new project.",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
				cmdhelpers.StandardNameFieldSelector(),
				cmdhelpers.StandardDistributionFieldSelector(),
				cmdhelpers.StandardDistributionConfigFieldSelector(),
				cmdhelpers.StandardSourceDirectoryFieldSelector(),
			}
			manager := configmanager.NewConfigManager(fieldSelectors...)
			return cmd.HandleInitRunEWithOutputPath(cobraCmd, manager, tempDir)
		},
	}
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Capture the output as a snapshot
	snaps.MatchSnapshot(t, out.String())

	// Verify files were created in temp directory, not current directory
	assert.FileExists(t, tempDir+"/ksail.yaml")
	assert.FileExists(t, tempDir+"/kind.yaml")
	assert.DirExists(t, tempDir+"/k8s")
	assert.FileExists(t, tempDir+"/k8s/kustomization.yaml")

	// Verify files were NOT created in current directory
	assert.NoFileExists(t, "./ksail.yaml")
	assert.NoFileExists(t, "./kind.yaml")
	assert.NoDirExists(t, "./k8s")
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

func TestHandleInitRunE(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		var out bytes.Buffer

		testCmd := &cobra.Command{}
		testCmd.SetOut(&out)

		// Create manager with the same field selectors as the real init command
		fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
			cmdhelpers.StandardDistributionFieldSelector(),
			cmdhelpers.StandardSourceDirectoryFieldSelector(),
		}
		manager := configmanager.NewConfigManager(fieldSelectors...)

		// Create a temporary directory for testing
		tempDir := t.TempDir()

		// Use the new function that accepts an output path
		err := cmd.HandleInitRunEWithOutputPath(testCmd, manager, tempDir)

		require.NoError(t, err)
		assert.Contains(t, out.String(), "✔ project initialized successfully")
		assert.Contains(t, out.String(), "► Distribution:")
		assert.Contains(t, out.String(), "► Source directory:")

		// Verify that scaffolder created the expected files in the temp directory
		assert.FileExists(t, tempDir+"/ksail.yaml")
		assert.FileExists(t, tempDir+"/kind.yaml")
		assert.DirExists(t, tempDir+"/k8s")
		assert.FileExists(t, tempDir+"/k8s/kustomization.yaml")
	})
}
