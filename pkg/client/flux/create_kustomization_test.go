package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCreateKustomizationCmd(t *testing.T) {
	t.Parallel()

	client := flux.NewClient(genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}, "")

	createCmd := client.CreateCreateCommand("")
	
	// Find kustomization command
	var kustomizationCmd *cobra.Command
	for _, subCmd := range createCmd.Commands() {
		if subCmd.Use == "kustomization [name]" {
			kustomizationCmd = subCmd
			break
		}
	}
	require.NotNil(t, kustomizationCmd)
	require.Equal(t, "Create or update a Kustomization resource", kustomizationCmd.Short)

	// Verify flags
	sourceKindFlag := kustomizationCmd.Flags().Lookup("source-kind")
	require.NotNil(t, sourceKindFlag)
	
	sourceFlag := kustomizationCmd.Flags().Lookup("source")
	require.NotNil(t, sourceFlag)
	
	pathFlag := kustomizationCmd.Flags().Lookup("path")
	require.NotNil(t, pathFlag)
	
	pruneFlag := kustomizationCmd.Flags().Lookup("prune")
	require.NotNil(t, pruneFlag)

	waitFlag := kustomizationCmd.Flags().Lookup("wait")
	require.NotNil(t, waitFlag)
	
	targetNamespaceFlag := kustomizationCmd.Flags().Lookup("target-namespace")
	require.NotNil(t, targetNamespaceFlag)
	
	intervalFlag := kustomizationCmd.Flags().Lookup("interval")
	require.NotNil(t, intervalFlag)
	
	exportFlag := kustomizationCmd.Flags().Lookup("export")
	require.NotNil(t, exportFlag)

	dependsOnFlag := kustomizationCmd.Flags().Lookup("depends-on")
	require.NotNil(t, dependsOnFlag)
}

func TestCreateKustomization_Export(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		args    []string
		flags   map[string]string
		wantErr bool
	}{
		"export basic kustomization": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source": "GitRepository/podinfo",
				"path":   "./kustomize",
				"export": "true",
			},
		},
		"export with prune": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source": "GitRepository/podinfo",
				"path":   "./kustomize",
				"prune":  "true",
				"export": "true",
			},
		},
		"export with wait": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source": "GitRepository/podinfo",
				"path":   "./deploy",
				"wait":   "true",
				"export": "true",
			},
		},
		"export with target namespace": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":           "GitRepository/podinfo",
				"path":             "./kustomize",
				"target-namespace": "production",
				"export":           "true",
			},
		},
		"export with custom interval": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":   "GitRepository/podinfo",
				"path":     "./kustomize",
				"interval": "5m",
				"export":   "true",
			},
		},
		"export with namespace": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source":    "GitRepository/podinfo",
				"path":      "./kustomize",
				"namespace": "custom-ns",
				"export":    "true",
			},
		},
		"export with dependencies": {
			args: []string{"app"},
			flags: map[string]string{
				"source":     "GitRepository/app",
				"path":       "./kustomize",
				"depends-on": "infra,database",
				"export":     "true",
			},
		},
		"export with source Kind/name format": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source": "GitRepository/podinfo",
				"path":   "./",
				"export": "true",
			},
		},
		"export with OCIRepository source": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"source-kind": "OCIRepository",
				"source":      "podinfo",
				"path":        "./kustomize",
				"export":      "true",
			},
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
			cmdLine := []string{"kustomization"}
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

func TestCreateKustomization_MissingRequiredSource(t *testing.T) {
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
	createCmd.SetArgs([]string{"kustomization", "podinfo", "--path", "./kustomize", "--export"})

	err := createCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required flag(s)")
}
