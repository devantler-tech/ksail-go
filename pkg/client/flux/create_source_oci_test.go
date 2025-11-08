package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCreateSourceOCICmd(t *testing.T) {
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

	// Find oci command
	var ociCmd *cobra.Command
	for _, subCmd := range sourceCmd.Commands() {
		if subCmd.Use == "oci [name]" {
			ociCmd = subCmd
			break
		}
	}
	require.NotNil(t, ociCmd)
	require.Equal(t, "Create or update an OCIRepository source", ociCmd.Short)

	// Verify flags
	urlFlag := ociCmd.Flags().Lookup("url")
	require.NotNil(t, urlFlag)

	tagFlag := ociCmd.Flags().Lookup("tag")
	require.NotNil(t, tagFlag)

	semverFlag := ociCmd.Flags().Lookup("tag-semver")
	require.NotNil(t, semverFlag)

	digestFlag := ociCmd.Flags().Lookup("digest")
	require.NotNil(t, digestFlag)

	secretRefFlag := ociCmd.Flags().Lookup("secret-ref")
	require.NotNil(t, secretRefFlag)

	providerFlag := ociCmd.Flags().Lookup("provider")
	require.NotNil(t, providerFlag)

	intervalFlag := ociCmd.Flags().Lookup("interval")
	require.NotNil(t, intervalFlag)

	exportFlag := ociCmd.Flags().Lookup("export")
	require.NotNil(t, exportFlag)

	insecureFlag := ociCmd.Flags().Lookup("insecure")
	require.NotNil(t, insecureFlag)
}

func TestCreateOCIRepository_Export(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args    []string
		flags   map[string]string
		wantErr bool
		errMsg  string
	}{
		"export with tag": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"tag":    "6.6.2",
				"export": "true",
			},
		},
		"export with semver": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":        "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"tag-semver": ">=6.0.0 <7.0.0",
				"export":     "true",
			},
		},
		"export with digest": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"digest": "sha256:abcdef123456",
				"export": "true",
			},
		},
		"export with secret ref": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":        "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"tag":        "6.6.2",
				"secret-ref": "oci-credentials",
				"export":     "true",
			},
		},
		"export with custom provider": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":      "oci://gcr.io/project/manifests",
				"tag":      "v1.0.0",
				"provider": "gcp",
				"export":   "true",
			},
		},
		"export with insecure": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":      "oci://localhost:5000/manifests",
				"tag":      "latest",
				"insecure": "true",
				"export":   "true",
			},
		},
		"export with custom interval": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":      "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"tag":      "6.6.2",
				"interval": "15m",
				"export":   "true",
			},
		},
		"export with namespace": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":       "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"tag":       "6.6.2",
				"namespace": "custom-ns",
				"export":    "true",
			},
		},
		"missing reference fails": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "oci://ghcr.io/stefanprodan/manifests/podinfo",
				"export": "true",
			},
			wantErr: true,
			errMsg:  "one of --tag, --tag-semver or --digest is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

			// Build command line
			cmdLine := []string{"source", "oci"}
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
			// TODO: Add snapshot test once snapshot infrastructure is properly configured
			// snaps.MatchSnapshot(t, output)
		})
	}
}

func TestCreateOCIRepository_MissingRequiredURL(t *testing.T) {
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
	createCmd.SetArgs([]string{"source", "oci", "podinfo", "--tag", "v1.0.0", "--export"})

	err := createCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required flag(s)")
}
