package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// testCase represents a test case for flux commands.
type testCase struct {
	args    []string
	flags   map[string]string
	wantErr bool
	errMsg  string
}

// findSourceCommand finds the source command from the create command.
func findSourceCommand(t *testing.T, createCmd *cobra.Command) *cobra.Command {
	t.Helper()

	var sourceCmd *cobra.Command

	for _, subCmd := range createCmd.Commands() {
		if subCmd.Use == sourceCommandName {
			sourceCmd = subCmd

			break
		}
	}

	require.NotNil(t, sourceCmd)

	return sourceCmd
}

// testMissingRequiredFlag tests that a command fails when a required flag is missing.
func testMissingRequiredFlag(t *testing.T, cmdPath []string, args []string) {
	t.Helper()

	var outBuf bytes.Buffer

	client := flux.NewClient(genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &outBuf,
		ErrOut: &bytes.Buffer{},
	}, "")

	createCmd := client.CreateCreateCommand("")
	createCmd.SetOut(&outBuf)
	createCmd.SetErr(&bytes.Buffer{})

	fullArgs := make([]string, 0, len(cmdPath)+len(args))
	fullArgs = append(fullArgs, cmdPath...)
	fullArgs = append(fullArgs, args...)
	createCmd.SetArgs(fullArgs)

	err := createCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required flag(s)")
}

// runFluxCommandTest executes a flux command test with the given parameters.
func runFluxCommandTest(t *testing.T, cmdPath []string, testCase testCase) {
	t.Helper()

	var outBuf bytes.Buffer

	client := flux.NewClient(genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &outBuf,
		ErrOut: &bytes.Buffer{},
	}, "")

	createCmd := client.CreateCreateCommand("")
	createCmd.SetOut(&outBuf)
	createCmd.SetErr(&bytes.Buffer{})

	// Build command line
	cmdLine := make([]string, 0, len(cmdPath)+len(testCase.args)+len(testCase.flags)*2)
	cmdLine = append(cmdLine, cmdPath...)
	cmdLine = append(cmdLine, testCase.args...)

	for flagKey, flagValue := range testCase.flags {
		if flagKey == "namespace" {
			// Namespace is a persistent flag that goes before the subcommand
			continue
		}
		// For boolean flags, only add the flag name without a value
		if flagValue == "true" {
			cmdLine = append(cmdLine, "--"+flagKey)
		} else {
			cmdLine = append(cmdLine, "--"+flagKey, flagValue)
		}
	}

	// Add namespace flag if present at the beginning
	if ns, ok := testCase.flags["namespace"]; ok {
		cmdLine = append([]string{"--namespace", ns}, cmdLine...)
	}

	createCmd.SetArgs(cmdLine)
	err := createCmd.Execute()

	if testCase.wantErr {
		require.Error(t, err)

		if testCase.errMsg != "" {
			require.Contains(t, err.Error(), testCase.errMsg)
		}

		return
	}

	require.NoError(t, err)

	// Validate YAML output
	output := outBuf.String()
	require.NotEmpty(t, output, "output should not be empty")
	require.Contains(t, output, "metadata:")
	require.Contains(t, output, "spec:")
}
