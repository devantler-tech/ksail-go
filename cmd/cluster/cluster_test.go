package cluster //nolint:testpackage // Needs access to unexported helpers for coverage instrumentation.

import (
	"bytes"
	"errors"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

func TestNewClusterCmdRegistersLifecycleCommands(t *testing.T) {
	t.Parallel()

	rt := newTestRuntime()
	metadata := expectedLifecycleMetadata(t, rt)
	requireParentMetadata(t, NewClusterCmd(rt))

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

	rt := newTestRuntime()
	cmd := NewClusterCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs(nil)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected executing cluster command without subcommand to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
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

	err := handleClusterRunE(cmd, nil)
	if !errors.Is(err, errHelpFailure) {
		t.Fatalf("expected wrapped help failure error, got %v", err)
	}
}

var errHelpFailure = errors.New("help failure")

type lifecycleMetadata struct {
	short string
	long  string
}

func expectedLifecycleMetadata(
	t *testing.T,
	runtimeContainer *runtime.Runtime,
) map[string]lifecycleMetadata {
	t.Helper()

	constructors := []func() *cobra.Command{
		func() *cobra.Command { return NewCreateCmd(runtimeContainer) },
		func() *cobra.Command { return NewDeleteCmd(runtimeContainer) },
		func() *cobra.Command { return NewStartCmd(runtimeContainer) },
		func() *cobra.Command { return NewStopCmd(runtimeContainer) },
		func() *cobra.Command { return NewListCmd(runtimeContainer) },
		func() *cobra.Command { return NewInfoCmd(runtimeContainer) },
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

	rt := newTestRuntime()

	for _, subcommand := range NewClusterCmd(rt).Commands() {
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

func newTestRuntime() *runtime.Runtime {
	return runtime.NewRuntime()
}
