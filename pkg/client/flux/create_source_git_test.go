package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewCreateSourceGitCmd(t *testing.T) {
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

	// Find git command
	var gitCmd *cobra.Command
	for _, subCmd := range sourceCmd.Commands() {
		if subCmd.Use == "git [name]" {
			gitCmd = subCmd
			break
		}
	}
	require.NotNil(t, gitCmd)
	require.Equal(t, "Create or update a GitRepository source", gitCmd.Short)

	// Verify required flags
	urlFlag := gitCmd.Flags().Lookup("url")
	require.NotNil(t, urlFlag)

	branchFlag := gitCmd.Flags().Lookup("branch")
	require.NotNil(t, branchFlag)

	tagFlag := gitCmd.Flags().Lookup("tag")
	require.NotNil(t, tagFlag)

	semverFlag := gitCmd.Flags().Lookup("tag-semver")
	require.NotNil(t, semverFlag)

	commitFlag := gitCmd.Flags().Lookup("commit")
	require.NotNil(t, commitFlag)

	secretRefFlag := gitCmd.Flags().Lookup("secret-ref")
	require.NotNil(t, secretRefFlag)

	intervalFlag := gitCmd.Flags().Lookup("interval")
	require.NotNil(t, intervalFlag)

	exportFlag := gitCmd.Flags().Lookup("export")
	require.NotNil(t, exportFlag)
}

func TestCreateGitRepository_Export(t *testing.T) {
	t.Parallel()

	tests := map[string]testCase{
		"export with branch": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "https://github.com/stefanprodan/podinfo",
				"branch": "master",
				"export": "true",
			},
		},
		"export with tag": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "https://github.com/stefanprodan/podinfo",
				"tag":    "6.6.2",
				"export": "true",
			},
		},
		"export with semver": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":        "https://github.com/stefanprodan/podinfo",
				"tag-semver": ">=6.0.0",
				"export":     "true",
			},
		},
		"export with commit": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":    "https://github.com/stefanprodan/podinfo",
				"commit": "abc123",
				"export": "true",
			},
		},
		"export with secret ref": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":        "ssh://git@github.com/stefanprodan/podinfo",
				"branch":     "main",
				"secret-ref": "git-credentials",
				"export":     "true",
			},
		},
		"export with namespace flag": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":       "https://github.com/stefanprodan/podinfo",
				"branch":    "master",
				"namespace": "custom-ns",
				"export":    "true",
			},
		},
		"export with custom interval": {
			args: []string{"podinfo"},
			flags: map[string]string{
				"url":      "https://github.com/stefanprodan/podinfo",
				"branch":   "master",
				"interval": "5m",
				"export":   "true",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runFluxCommandTest(t, []string{"source", "git"}, tc)
		})
	}
}

func TestCreateGitRepository_MissingRequiredURL(t *testing.T) {
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
	createCmd.SetArgs([]string{"source", "git", "podinfo", "--branch", "main", "--export"})

	err := createCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required flag(s)")
}
