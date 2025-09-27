package cluster_test

import (
	"testing"

	cluster "github.com/devantler-tech/ksail-go/cmd/cluster"
	"github.com/spf13/cobra"
)

func TestNewClusterCmdRegistersLifecycleCommands(t *testing.T) {
	t.Parallel()

	cmd := cluster.NewClusterCmd()

	if cmd.Short != "Manage cluster lifecycle commands" {
		t.Fatalf(
			"short description mismatch for parent command. want %q, got %q",
			"Manage cluster lifecycle commands",
			cmd.Short,
		)
	}

	expected := map[string]struct {
		short string
		long  string
	}{
		"up": {
			short: "Start the Kubernetes cluster",
			long:  "Start the Kubernetes cluster defined in the project configuration.",
		},
		"down": {
			short: "Destroy a cluster",
			long:  "Destroy a cluster.",
		},
		"start": {
			short: "Start a stopped cluster",
			long:  "Start a previously stopped cluster.",
		},
		"stop": {
			short: "Stop the Kubernetes cluster",
			long:  "Stop the Kubernetes cluster without removing it.",
		},
		"status": {
			short: "Show status of the Kubernetes cluster",
			long:  "Show the current status of the Kubernetes cluster.",
		},
		"list": {
			short: "List clusters",
			long:  "List all Kubernetes clusters managed by KSail.",
		},
	}

	validateSubcommandExists := func(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
		t.Helper()

		sub := parent.Commands()
		for _, c := range sub {
			if c.Use == name {
				return c
			}
		}

		t.Fatalf("expected cluster command to include %q subcommand", name)

		return nil
	}

	for use, metadata := range expected {
		sub := validateSubcommandExists(t, cmd, use)

		if sub.Short != metadata.short {
			t.Fatalf(
				"short description mismatch for %q. want %q, got %q",
				use,
				metadata.short,
				sub.Short,
			)
		}

		if sub.Long != metadata.long {
			t.Fatalf(
				"long description mismatch for %q. want %q, got %q",
				use,
				metadata.long,
				sub.Long,
			)
		}
	}
}
