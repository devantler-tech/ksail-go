package workload_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
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

func TestWorkloadCommandsEmitPlaceholders(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "reconcile",
			args:     []string{"reconcile"},
			expected: "ℹ Workload reconciliation coming soon.",
		},
		{
			name:     "apply",
			args:     []string{"apply"},
			expected: "ℹ Workload apply coming soon.",
		},
		{
			name:     "install",
			args:     []string{"install"},
			expected: "ℹ Workload install coming soon.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			command := workload.NewWorkloadCmd()
			command.SetOut(&out)
			command.SetErr(&out)
			command.SetArgs(testCase.args)

			err := command.Execute()
			require.NoErrorf(t, err, "expected workload %s command to succeed", testCase.name)

			actual := out.String()
			if !strings.Contains(actual, testCase.expected) {
				t.Fatalf(
					"expected placeholder output to contain %q, got %q",
					testCase.expected,
					actual,
				)
			}
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

func TestWorkloadCommandConfiguration(t *testing.T) {
	t.Parallel()

	command := workload.NewWorkloadCmd()

	require.True(t, command.SilenceErrors)
	require.True(t, command.SilenceUsage)
	require.Equal(t, helpers.SuggestionsMinimumDistance, command.SuggestionsMinimumDistance)
}
