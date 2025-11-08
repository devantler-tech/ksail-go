package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
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
