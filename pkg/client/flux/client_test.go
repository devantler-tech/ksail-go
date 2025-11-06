// Package flux provides a flux client implementation using Flux Kubernetes APIs.
package flux_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/flux"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams, "")
	require.NotNil(t, client)
}

func TestCreateCreateCommand(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams, "")
	cmd := client.CreateCreateCommand("")

	require.NotNil(t, cmd)
	require.Equal(t, "flux-create", cmd.Use)
	require.Equal(t, "Create Flux resources", cmd.Short)

	// Check that sub-commands are added
	subCommands := cmd.Commands()
	require.NotEmpty(t, subCommands)

	// Check for source command (currently the only implemented command)
	var sourceCmd *cobra.Command

	for _, subCmd := range subCommands {
		if subCmd.Use == "source" {
			sourceCmd = subCmd

			break
		}
	}

	require.NotNil(t, sourceCmd)
	require.Equal(t, "Create or update Flux sources", sourceCmd.Short)
}

func TestCreateSourceGitCommand(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams, "")
	cmd := client.CreateCreateCommand("")

	// Find source command
	var sourceCmd *cobra.Command

	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "source" {
			sourceCmd = subCmd

			break
		}
	}

	require.NotNil(t, sourceCmd)

	// Find git sub-command
	var gitCmd *cobra.Command

	for _, subCmd := range sourceCmd.Commands() {
		if subCmd.Use == "git [name]" {
			gitCmd = subCmd

			break
		}
	}

	require.NotNil(t, gitCmd)
	require.Equal(t, "Create or update a GitRepository source", gitCmd.Short)

	// Check that required flags are present
	urlFlag := gitCmd.Flags().Lookup("url")
	require.NotNil(t, urlFlag)

	branchFlag := gitCmd.Flags().Lookup("branch")
	require.NotNil(t, branchFlag)

	intervalFlag := gitCmd.Flags().Lookup("interval")
	require.NotNil(t, intervalFlag)
}
