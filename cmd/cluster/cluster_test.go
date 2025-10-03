package cluster //nolint:testpackage // Needs access to unexported helpers for coverage instrumentation.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

func TestNewClusterCmdRegistersLifecycleCommands(t *testing.T) {
	t.Parallel()

	metadata := expectedLifecycleMetadata(t)
	requireParentMetadata(t, NewClusterCmd())

	for name, details := range metadata {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			subcommand := findLifecycleSubcommand(t, name)
			assertSubcommandMetadata(t, subcommand, details)
		})
	}
}

func TestClusterCommandRunEDisplaysHelp(t *testing.T) {
	t.Parallel()

	cmd := NewClusterCmd()
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected executing cluster command without subcommand to succeed, got %v", err)
	}

	assertOutputContains(t, buffer.String(), "Usage:")
	assertOutputContains(t, buffer.String(), "Available Commands:")
}

//nolint:paralleltest // Alters package-level helper.
func TestHandleClusterRunEWrapsHelpError(t *testing.T) {
	originalRunner := helpRunner
	helpRunner = func(*cobra.Command) error {
		return errHelpFailure
	}

	defer func() {
		helpRunner = originalRunner
	}()

	cmd := &cobra.Command{Use: "cluster"}

	err := handleClusterRunE(cmd, nil, nil)
	if !errors.Is(err, errHelpFailure) {
		t.Fatalf("expected wrapped help failure error, got %v", err)
	}
}

var errHelpFailure = errors.New("help failure")

type lifecycleMetadata struct {
	short string
	long  string
}

func expectedLifecycleMetadata(t *testing.T) map[string]lifecycleMetadata {
	t.Helper()

	constructors := []func() *cobra.Command{
		NewUpCmd,
		NewDownCmd,
		NewStartCmd,
		NewStopCmd,
		NewStatusCmd,
		NewListCmd,
	}

	metadata := make(map[string]lifecycleMetadata, len(constructors))

	for _, constructor := range constructors {
		cmd := constructor()
		metadata[cmd.Use] = lifecycleMetadata{
			short: cmd.Short,
			long:  cmd.Long,
		}
	}

	return metadata
}

func requireParentMetadata(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	const expectedDescription = "Manage cluster lifecycle"

	if cmd.Short != expectedDescription {
		t.Fatalf(
			"short description mismatch for parent command. want %q, got %q",
			expectedDescription,
			cmd.Short,
		)
	}
}

func findLifecycleSubcommand(t *testing.T, name string) *cobra.Command {
	t.Helper()

	for _, subcommand := range NewClusterCmd().Commands() {
		if subcommand.Use == name {
			return subcommand
		}
	}

	t.Fatalf("expected cluster command to include %q subcommand", name)

	return nil
}

func assertSubcommandMetadata(t *testing.T, cmd *cobra.Command, metadata lifecycleMetadata) {
	t.Helper()

	if cmd.Short != metadata.short {
		t.Fatalf(
			"short description mismatch for %q. want %q, got %q",
			cmd.Use,
			metadata.short,
			cmd.Short,
		)
	}

	if cmd.Long != metadata.long {
		t.Fatalf(
			"long description mismatch for %q. want %q, got %q",
			cmd.Use,
			metadata.long,
			cmd.Long,
		)
	}
}

func assertOutputContains(t *testing.T, output, expected string) {
	t.Helper()

	if !strings.Contains(output, expected) {
		t.Fatalf("expected output to contain %q, got %q", expected, output)
	}
}

type lifecycleHandlerFunc func(*cobra.Command, *configmanager.ConfigManager, []string) error

func runLifecycleSuccessCase(
	t *testing.T,
	use string,
	handler lifecycleHandlerFunc,
	expectedMessage string,
) {
	t.Helper()

	cmd, manager, output := testutils.NewCommandAndManager(t, use)
	testutils.SeedValidClusterConfig(manager)

	err := handler(cmd, manager, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assertOutputContains(t, output.String(), expectedMessage)
}

func runLifecycleValidationErrorCase(
	t *testing.T,
	use string,
	handler lifecycleHandlerFunc,
	expectedSubstrings ...string,
) {
	t.Helper()

	testutils.RunValidationErrorTest(t, use, func(
		cmd *cobra.Command,
		manager *configmanager.ConfigManager,
		args []string,
	) error {
		err := handler(cmd, manager, args)
		if err == nil {
			t.Fatal("expected error but got nil")
		}

		message := err.Error()
		for _, substring := range expectedSubstrings {
			if !strings.Contains(message, substring) {
				t.Fatalf("expected error message to contain %q, got %q", substring, message)
			}
		}

		return err
	})
}
