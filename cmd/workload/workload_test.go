package workload_test

// cspell:words cmdtestutils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils" // cspell:ignore cmdtestutils
	"github.com/devantler-tech/ksail-go/cmd/workload"
	internaltestutils "github.com/devantler-tech/ksail-go/internal/testutils"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/samber/do/v2"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { internaltestutils.RunTestMainWithSnapshotCleanup(m) }

func TestWorkloadHelpSnapshots(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{name: "namespace", args: []string{"workload", "--help"}},
		{name: "reconcile", args: []string{"workload", "reconcile", "--help"}},
		{name: "apply", args: []string{"workload", "apply", "--help"}},
		{name: "create", args: []string{"workload", "create", "--help"}},
		{name: "delete", args: []string{"workload", "delete", "--help"}},
		{name: "describe", args: []string{"workload", "describe", "--help"}},
		{name: "edit", args: []string{"workload", "edit", "--help"}},
		{name: "explain", args: []string{"workload", "explain", "--help"}},
		{name: "expose", args: []string{"workload", "expose", "--help"}},
		{name: "get", args: []string{"workload", "get", "--help"}},
		{name: "install", args: []string{"workload", "install", "--help"}},
		{name: "rollout", args: []string{"workload", "rollout", "--help"}},
		{name: "scale", args: []string{"workload", "scale", "--help"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			root := cmd.NewRootCmd("test", "test", "test")
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(testCase.args)

			err := root.Execute()
			require.NoErrorf(
				t,
				err,
				"expected no error executing %s help",
				strings.Join(testCase.args, " "),
			)

			snaps.MatchSnapshot(t, out.String())
		})
	}
}

//nolint:paralleltest // Uses t.Chdir which is incompatible with parallel tests.
func TestWorkloadCommandsLoadConfigOnly(t *testing.T) {
	// Note: "apply" is excluded as it's a full kubectl wrapper, not a config-only placeholder
	commands := []string{"reconcile", "install"}

	for _, commandName := range commands {
		t.Run(commandName, func(t *testing.T) {
			var out bytes.Buffer

			tempDir := t.TempDir()
			cmdtestutils.WriteValidKsailConfig(t, tempDir)

			t.Chdir(tempDir)

			root := cmd.NewRootCmd("test", "test", "test")
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs([]string{"workload", commandName})

			err := root.Execute()
			require.NoErrorf(t, err, "expected workload %s handler to succeed", commandName)

			actual := out.String()
			require.Contains(t, actual, "config loaded")
			require.NotContains(t, actual, "coming soon")
			require.NotContains(t, actual, "ℹ")
		})
	}
}

func TestNewWorkloadCmdRunETriggersHelp(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.New(func(injector do.Injector) error {
		do.Provide(injector, func(do.Injector) (timer.Timer, error) {
			return timer.New(), nil
		})

		return nil
	})

	var out bytes.Buffer

	command := workload.NewWorkloadCmd(runtimeContainer)
	command.SetOut(&out)
	command.SetErr(&out)

	err := command.Execute()
	require.NoError(t, err)

	output := out.String()
	if !strings.Contains(output, "Group workload commands under a single namespace") {
		t.Fatalf("expected help output to mention workload namespace details, got %q", output)
	}

	if !strings.Contains(output, "edit") {
		t.Fatalf("expected help output to mention edit command, got %q", output)
	}
}
