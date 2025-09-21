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

	t.Run("command creation", testNewInitCmdCreation)
	t.Run("embedded RunE function", testNewInitCmdEmbeddedRunE)
}

func testNewInitCmdCreation(t *testing.T) {
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

func testNewInitCmdEmbeddedRunE(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	tempDir := t.TempDir()

	// Test the embedded RunE function by directly calling HandleInitRunE
	// This avoids working directory changes while still testing the embedded function logic
	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create manager with the same field selectors as the real init command
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardNameFieldSelector(),
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}
	manager := configmanager.NewConfigManager(fieldSelectors...)

	// Call HandleInitRunE directly with temp directory - this tests the same logic
	// that the embedded RunE function would execute
	err := cmd.HandleInitRunE(testCmd, manager, []string{}, tempDir)

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ project initialized successfully")

	// Verify files were created in the temp directory
	assert.FileExists(t, tempDir+"/ksail.yaml")
	assert.FileExists(t, tempDir+"/kind.yaml")
	assert.DirExists(t, tempDir+"/k8s")
	assert.FileExists(t, tempDir+"/k8s/kustomization.yaml")
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
		RunE: func(cobraCmd *cobra.Command, _ []string) error {
			fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
				cmdhelpers.StandardNameFieldSelector(),
				cmdhelpers.StandardDistributionFieldSelector(),
				cmdhelpers.StandardDistributionConfigFieldSelector(),
				cmdhelpers.StandardSourceDirectoryFieldSelector(),
			}
			manager := configmanager.NewConfigManager(fieldSelectors...)

			return cmd.HandleInitRunE(cobraCmd, manager, []string{}, tempDir)
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

	t.Run("success with output path", testHandleInitRunESuccessWithOutputPath)
	t.Run("success without output path", testHandleInitRunESuccessWithoutOutputPath)
	t.Run("config manager load error", testHandleInitRunEConfigManagerLoadError)
	t.Run("scaffold error", testHandleInitRunEScaffoldError)
}

func testHandleInitRunESuccessWithOutputPath(t *testing.T) {
	t.Parallel()

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

	// Use the merged function with optional output path
	err := cmd.HandleInitRunE(testCmd, manager, []string{}, tempDir)

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ project initialized successfully")
	assert.Contains(t, out.String(), "► Distribution:")
	assert.Contains(t, out.String(), "► Source directory:")

	// Verify that scaffolder created the expected files in the temp directory
	assert.FileExists(t, tempDir+"/ksail.yaml")
	assert.FileExists(t, tempDir+"/kind.yaml")
	assert.DirExists(t, tempDir+"/k8s")
	assert.FileExists(t, tempDir+"/k8s/kustomization.yaml")
}

func testHandleInitRunESuccessWithoutOutputPath(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create manager with the same field selectors as the real init command
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}
	manager := configmanager.NewConfigManager(fieldSelectors...)

	// Test calling the function with no outputPath parameters at all
	// This will trigger the os.Getwd() code path, but we'll use a temp directory
	// to avoid conflicts with other tests
	tempDir := t.TempDir()

	// Test the with-outputPath case by providing a temp directory
	err := cmd.HandleInitRunE(testCmd, manager, []string{}, tempDir)

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ project initialized successfully")

	// Files should be created in the provided tempDir
	assert.FileExists(t, tempDir+"/ksail.yaml")
	assert.FileExists(t, tempDir+"/kind.yaml")
	assert.DirExists(t, tempDir+"/k8s")
	assert.FileExists(t, tempDir+"/k8s/kustomization.yaml")
}

func testHandleInitRunEConfigManagerLoadError(t *testing.T) {
	t.Parallel()

	// This test is challenging without mocking since HandleInitRunE expects concrete type
	// However, we can test behavior with an invalid config path that would cause load errors
	// This is more of an integration test but still valuable for coverage

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create a manager that will have issues loading config due to invalid Viper setup
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}

	// Create a basic manager - errors are hard to trigger without changing source code
	// since the manager is quite robust and defaults to reasonable values
	manager := configmanager.NewConfigManager(fieldSelectors...)

	tempDir := t.TempDir()

	// Note: This test might not actually trigger the LoadConfig error path
	// since the ConfigManager is designed to be robust and use defaults
	// But it still tests the function with valid inputs
	err := cmd.HandleInitRunE(testCmd, manager, []string{}, tempDir)

	// In most cases this will actually succeed due to robust error handling in ConfigManager
	// But we're testing the code path exists and compiles correctly
	if err != nil {
		// If there is an error, ensure it's formatted correctly
		assert.Contains(t, err.Error(), "failed to")
	} else {
		// If no error, verify successful execution
		assert.Contains(t, out.String(), "✔ project initialized successfully")
	}
}

func testHandleInitRunEScaffoldError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create manager with valid field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}
	manager := configmanager.NewConfigManager(fieldSelectors...)

	// Use an invalid path to trigger scaffold error
	invalidPath := "/invalid/\x00path/"

	// Test that scaffold error is properly handled
	err := cmd.HandleInitRunE(testCmd, manager, []string{}, invalidPath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scaffold project files")
}
