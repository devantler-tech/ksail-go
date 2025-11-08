package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCreateSourceHelmCmd(t *testing.T) {
	t.Parallel()

	client := flux.NewClient(genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}, "")

	createCmd := client.CreateCreateCommand("")

	// Find source command
	var sourceCmd *cobra.Command

	for _, subCmd := range createCmd.Commands() {
		if subCmd.Use == "source" {
			sourceCmd = subCmd

			break
		}
	}

	require.NotNil(t, sourceCmd)

	// Find helm command
	var helmCmd *cobra.Command

	for _, subCmd := range sourceCmd.Commands() {
		if subCmd.Use == "helm [name]" {
			helmCmd = subCmd

			break
		}
	}

	require.NotNil(t, helmCmd)
	require.Equal(t, "Create or update a HelmRepository source", helmCmd.Short)

	// Verify flags
	urlFlag := helmCmd.Flags().Lookup("url")
	require.NotNil(t, urlFlag)

	secretRefFlag := helmCmd.Flags().Lookup("secret-ref")
	require.NotNil(t, secretRefFlag)

	intervalFlag := helmCmd.Flags().Lookup("interval")
	require.NotNil(t, intervalFlag)

	exportFlag := helmCmd.Flags().Lookup("export")
	require.NotNil(t, exportFlag)

	ociProviderFlag := helmCmd.Flags().Lookup("oci-provider")
	require.NotNil(t, ociProviderFlag)

	passCredentialsFlag := helmCmd.Flags().Lookup("pass-credentials")
	require.NotNil(t, passCredentialsFlag)
}

func TestCreateHelmRepository_Export(t *testing.T) {
	t.Parallel()

	tests := map[string]testCase{
		"export HTTPS repository": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "https://stefanprodan.github.io/podinfo",
				"export": "true",
			},
		},
		"export OCI repository": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "oci://ghcr.io/stefanprodan/charts",
				"export": "true",
			},
		},
		"export with secret ref": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":        "https://charts.example.com",
				"secret-ref": "helm-credentials",
				"export":     "true",
			},
		},
		"export OCI with provider": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":          "oci://ghcr.io/stefanprodan/charts",
				"oci-provider": "generic",
				"export":       "true",
			},
		},
		"export with pass credentials": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":              "https://charts.example.com",
				"secret-ref":       "helm-creds",
				"pass-credentials": "true",
				"export":           "true",
			},
		},
		"export with custom interval": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":      "https://stefanprodan.github.io/podinfo",
				"interval": "10m",
				"export":   "true",
			},
		},
		"export with namespace": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":       "https://stefanprodan.github.io/podinfo",
				"namespace": "custom-ns",
				"export":    "true",
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			runFluxCommandTest(t, []string{"source", "helm"}, testCase)
		})
	}
}

func TestCreateHelmRepository_MissingRequiredURL(t *testing.T) {
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
	createCmd.SetArgs([]string{"source", "helm", "podinfo", "--export"})

	err := createCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required flag(s)")
}
