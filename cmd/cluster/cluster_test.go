package cluster_test

import (
	"bytes"
	"strings"
	"testing"

	cluster "github.com/devantler-tech/ksail-go/cmd/cluster"
	"github.com/spf13/cobra"
)

func TestNewClusterCmdRegistersLifecycleCommands(t *testing.T) {
	t.Parallel()

	cmd := cluster.NewClusterCmd()

	requireParentMetadata(t, cmd)

	for name, metadata := range expectedLifecycleMetadata(t) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			subcommand := findSubcommand(t, cmd, name)
			assertSubcommandMetadata(t, subcommand, metadata)
		})
	}
}

func TestClusterCommandRunEDisplaysHelp(t *testing.T) {
	t.Parallel()

	cmd := cluster.NewClusterCmd()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected executing cluster command without subcommand to succeed, got %v", err)
	}

	output := out.String()

	if !strings.Contains(output, "Usage:") {
		t.Fatalf("expected help output to contain Usage section, got %q", output)
	}

	if !strings.Contains(output, "Available Commands:") {
		t.Fatalf("expected help output to list available commands, got %q", output)
	}
}

type lifecycleMetadata struct {
	short string
	long  string
}

func expectedLifecycleMetadata(t *testing.T) map[string]lifecycleMetadata {
	t.Helper()

	commandConstructors := []func() *cobra.Command{
		cluster.NewUpCmd,
		cluster.NewDownCmd,
		cluster.NewStartCmd,
		cluster.NewStopCmd,
		cluster.NewStatusCmd,
		cluster.NewListCmd,
	}

	metadata := make(map[string]lifecycleMetadata, len(commandConstructors))

	for _, constructor := range commandConstructors {
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

	if cmd.Short != "Manage cluster lifecycle commands" {
		t.Fatalf(
			"short description mismatch for parent command. want %q, got %q",
			"Manage cluster lifecycle commands",
			cmd.Short,
		)
	}
}

func findSubcommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()

	for _, subcommand := range parent.Commands() {
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
