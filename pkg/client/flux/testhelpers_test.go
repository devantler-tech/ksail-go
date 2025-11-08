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
func runFluxCommandTest(t *testing.T, cmdPath []string, tc testCase) {
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
	cmdLine := make([]string, 0, len(cmdPath)+len(tc.args)+len(tc.flags)*2)
	cmdLine = append(cmdLine, cmdPath...)
	cmdLine = append(cmdLine, tc.args...)

	for k, v := range tc.flags {
		if k == "namespace" {
			// Namespace is a persistent flag that goes before the subcommand
			continue
		}
		// For boolean flags, only add the flag name without a value
		if v == "true" {
			cmdLine = append(cmdLine, "--"+k)
		} else {
			cmdLine = append(cmdLine, "--"+k, v)
		}
	}

	// Add namespace flag if present at the beginning
	if ns, ok := tc.flags["namespace"]; ok {
		cmdLine = append([]string{"--namespace", ns}, cmdLine...)
	}

	createCmd.SetArgs(cmdLine)
	err := createCmd.Execute()

	if tc.wantErr {
		require.Error(t, err)
		if tc.errMsg != "" {
			require.Contains(t, err.Error(), tc.errMsg)
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
