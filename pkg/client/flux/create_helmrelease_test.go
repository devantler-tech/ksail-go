package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCreateHelmReleaseCmd(t *testing.T) {
	t.Parallel()

	client := setupTestClient()
	createCmd := client.CreateCreateCommand("")

	// Find helmrelease command
	var helmReleaseCmd *cobra.Command

	for _, subCmd := range createCmd.Commands() {
		if subCmd.Use == "helmrelease [name]" {
			helmReleaseCmd = subCmd

			break
		}
	}

	require.NotNil(t, helmReleaseCmd)
	require.Equal(t, "Create or update a HelmRelease resource", helmReleaseCmd.Short)
	require.Contains(t, helmReleaseCmd.Aliases, "hr")

	// Verify flags
	sourceKindFlag := helmReleaseCmd.Flags().Lookup("source-kind")
	require.NotNil(t, sourceKindFlag)

	sourceFlag := helmReleaseCmd.Flags().Lookup("source")
	require.NotNil(t, sourceFlag)

	chartFlag := helmReleaseCmd.Flags().Lookup("chart")
	require.NotNil(t, chartFlag)

	chartVersionFlag := helmReleaseCmd.Flags().Lookup("chart-version")
	require.NotNil(t, chartVersionFlag)

	targetNamespaceFlag := helmReleaseCmd.Flags().Lookup("target-namespace")
	require.NotNil(t, targetNamespaceFlag)

	createNamespaceFlag := helmReleaseCmd.Flags().Lookup("create-target-namespace")
	require.NotNil(t, createNamespaceFlag)

	intervalFlag := helmReleaseCmd.Flags().Lookup("interval")
	require.NotNil(t, intervalFlag)

	exportFlag := helmReleaseCmd.Flags().Lookup("export")
	require.NotNil(t, exportFlag)

	dependsOnFlag := helmReleaseCmd.Flags().Lookup("depends-on")
	require.NotNil(t, dependsOnFlag)
}

func TestCreateHelmRelease_Export(t *testing.T) {
	t.Parallel()

	tests := map[string]testCase{
		"export basic helmrelease": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source": "HelmRepository/podinfo",
				"chart":  "podinfo",
				"export": "true",
			},
		},
		"export with chart version": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":        "HelmRepository/podinfo",
				"chart":         "podinfo",
				"chart-version": "6.6.2",
				"export":        "true",
			},
		},
		"export with target namespace": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":           "HelmRepository/podinfo",
				"chart":            "podinfo",
				"target-namespace": "production",
				"export":           "true",
			},
		},
		"export with create namespace": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":                  "HelmRepository/podinfo",
				"chart":                   "podinfo",
				"target-namespace":        "new-ns",
				"create-target-namespace": "true",
				"export":                  "true",
			},
		},
		"export with custom interval": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":   "HelmRepository/podinfo",
				"chart":    "podinfo",
				"interval": "10m",
				"export":   "true",
			},
		},
		"export with namespace flag": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":    "HelmRepository/podinfo",
				"chart":     "podinfo",
				"namespace": "custom-ns",
				"export":    "true",
			},
		},
		"export with dependencies": {
			args: []string{"app"},
			flags: map[string]string{
				"source":     "HelmRepository/app",
				"chart":      "app",
				"depends-on": "database,cache",
				"export":     "true",
			},
		},
		"export with GitRepository source": {
			args: []string{"app"},
			flags: map[string]string{
				"source-kind": "GitRepository",
				"source":      "app",
				"chart":       "./charts/app",
				"export":      "true",
			},
		},
		"export with source Kind/name format": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source": "HelmRepository/podinfo",
				"chart":  "podinfo",
				"export": "true",
			},
		},
		"export with cross-namespace source": {
			args: []string{"app"},
			flags: map[string]string{
				"source": "HelmRepository/charts.flux-system",
				"chart":  "app",
				"export": "true",
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			runFluxCommandTest(t, []string{"helmrelease"}, testCase)
		})
	}
}

func TestCreateHelmRelease_MissingRequiredFlags(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args   []string
		errMsg string
	}{
		"missing source": {
			args:   []string{"podinfo", "--chart", "podinfo", "--export"},
			errMsg: "required flag(s)",
		},
		"missing chart": {
			args:   []string{"podinfo", "--source", "HelmRepository/podinfo", "--export"},
			errMsg: "required flag(s)",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			var outBuf bytes.Buffer

			client := flux.NewClient(genericiooptions.IOStreams{
				In:     &bytes.Buffer{},
				Out:    &outBuf,
				ErrOut: &bytes.Buffer{},
			}, "")

			createCmd := client.CreateCreateCommand("")
			createCmd.SetOut(&outBuf)
			createCmd.SetErr(&bytes.Buffer{})
			createCmd.SetArgs(append([]string{"helmrelease"}, testCase.args...))

			err := createCmd.Execute()
			require.Error(t, err)
			require.Contains(t, err.Error(), testCase.errMsg)
		})
	}
}

func TestCreateHelmRelease_AliasWorks(t *testing.T) {
	t.Parallel()

	var outBuf bytes.Buffer

	client := flux.NewClient(genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &outBuf,
		ErrOut: &bytes.Buffer{},
	}, "")

	createCmd := client.CreateCreateCommand("")
	createCmd.SetOut(&outBuf)
	createCmd.SetErr(&bytes.Buffer{})
	createCmd.SetArgs(
		[]string{
			"hr",
			"podinfo",
			"--source",
			"HelmRepository/podinfo",
			"--chart",
			"podinfo",
			"--export",
		},
	)

	err := createCmd.Execute()
	require.NoError(t, err)

	output := outBuf.String()
	require.NotEmpty(t, output, "output should not be empty")
	require.Contains(t, output, "metadata:")
	require.Contains(t, output, "spec:")
}
