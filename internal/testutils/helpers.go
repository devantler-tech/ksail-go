package testutils

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

const (
	defaultKsailConfigContent = `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  sourceDirectory: k8s
`

	defaultKindConfigContent = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: kind
`
)

// CreateConfigManagerWithFieldSelectors builds a config manager populated with the provided field selectors.
func CreateConfigManagerWithFieldSelectors(
	writer io.Writer,
	fieldSelectors ...configmanager.FieldSelector[v1alpha1.Cluster],
) *configmanager.ConfigManager {
	return configmanager.NewConfigManager(writer, fieldSelectors...)
}

// CreateDefaultConfigManager returns a config manager configured with the standard KSail defaults used in tests.
func CreateDefaultConfigManager() *configmanager.ConfigManager {
	return CreateConfigManagerWithFieldSelectors(
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
			DefaultValue: "kind-kind",
		},
	)
}

const (
	testDirectoryPerm = 0o750
	testFilePerm      = 0o600
)

// WriteValidKsailConfig writes a minimal valid KSail configuration into the provided directory.
func WriteValidKsailConfig(t *testing.T, dir string) {
	t.Helper()

	workloadDir := filepath.Join(dir, "k8s")
	require.NoError(t, os.MkdirAll(workloadDir, testDirectoryPerm))

	configPath := filepath.Join(dir, "ksail.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(defaultKsailConfigContent), testFilePerm))

	kindConfigPath := filepath.Join(dir, "kind.yaml")
	require.NoError(t, os.WriteFile(kindConfigPath, []byte(defaultKindConfigContent), testFilePerm))
}

// SetupCommandWithOutput returns a new cobra command with its output wired to an in-memory buffer.
func SetupCommandWithOutput() (*cobra.Command, *bytes.Buffer) {
	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	return testCmd, &out
}

// SimpleCommandTestData describes expectations for reusable command creation tests.
type SimpleCommandTestData struct {
	CommandName   string
	NewCommand    func() *cobra.Command
	ExpectedUse   string
	ExpectedShort string
}

// TestSimpleCommandCreation validates the basic metadata of a command.
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

// TestSimpleCommandExecution executes a command and snapshots its standard output.
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

// TestSimpleCommandHelp runs a command with --help and snapshots the output.
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

// TestCmdExecuteInCleanDir ensures commands fail validation when executed without a ksail.yaml.
func TestCmdExecuteInCleanDir(t *testing.T, cmdFactory func() *cobra.Command, cmdName string) {
	t.Helper()
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	t.Chdir(tempDir)

	defer func() { t.Chdir(originalDir) }()

	cmd := cmdFactory()

	err = cmd.Execute()
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

// --- Generic snapshot & assertion helpers (merged from duplicate block) ---

// RunTestMainWithSnapshotCleanup runs the standard TestMain pattern with snapshot cleanup.
// Shared across packages that only need snapshot cleanup (non command-specific logic).
func RunTestMainWithSnapshotCleanup(m *testing.M) {
	exitCode := m.Run()

	_, err := snaps.Clean(m, snaps.CleanOpts{Sort: true})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to clean snapshots: " + err.Error() + "\n")

		os.Exit(1)
	}

	os.Exit(exitCode)
}

// ExpectNoError fails the test if err is not nil.
func ExpectNoError(t *testing.T, err error, description string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: unexpected error: %v", description, err)
	}
}

// ExpectErrorContains fails the test if err is nil or does not contain substr.
func ExpectErrorContains(t *testing.T, err error, substr, description string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected error containing %q but got nil", description, substr)
	}

	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("%s: expected error to contain %q, got %q", description, substr, err.Error())
	}
}

// ExpectNotNil fails the test if value is nil.
func ExpectNotNil(t *testing.T, value any, description string) {
	t.Helper()

	if value == nil {
		t.Fatalf("expected %s to be non-nil", description)
	}
}

// ExpectTrue fails the test if condition is false.
func ExpectTrue(t *testing.T, condition bool, description string) {
	t.Helper()

	if !condition {
		t.Fatalf("expected %s to be true", description)
	}
}
