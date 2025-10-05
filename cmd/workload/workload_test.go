package workload_test

// cspell:words cmdtestutils

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils" // cspell:ignore cmdtestutils
	"github.com/devantler-tech/ksail-go/cmd/workload"
	internaltestutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
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
		{name: "install", args: []string{"workload", "install", "--help"}},
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

func TestWorkloadCommandsLoadConfigOnly(t *testing.T) {
	commands := []string{"reconcile", "apply", "install"}

	for _, commandName := range commands {
		t.Run(commandName, func(t *testing.T) {
			var out bytes.Buffer

			tempDir := t.TempDir()
			cmdtestutils.WriteValidKsailConfig(t, tempDir)

			originalDir, err := os.Getwd()
			require.NoError(t, err)

			t.Cleanup(func() {
				require.NoError(t, os.Chdir(originalDir))
			})

			require.NoError(t, os.Chdir(tempDir))

			root := cmd.NewRootCmd("test", "test", "test")
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs([]string{"workload", commandName})

			err = root.Execute()
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

	var out bytes.Buffer

	command := workload.NewWorkloadCmd()
	command.SetOut(&out)
	command.SetErr(&out)

	err := command.Execute()
	require.NoError(t, err)

	output := out.String()
	if !strings.Contains(output, "Group workload commands under a single namespace") {
		t.Fatalf("expected help output to mention workload namespace details, got %q", output)
	}
}
