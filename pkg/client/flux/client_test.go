// Package flux provides a flux client implementation that wraps the flux CLI.
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

	client := flux.NewClient(ioStreams)
	require.NotNil(t, client)
}

func TestCreateCreateCommand(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams)
	cmd := client.CreateCreateCommand("")

	require.NotNil(t, cmd)
	require.Equal(t, "flux-create", cmd.Use)
	require.Equal(t, "Create Flux resources", cmd.Short)

	// Check that sub-commands are added
	subCommands := cmd.Commands()
	require.NotEmpty(t, subCommands)

	// Check for expected sub-commands
	expectedSubCommands := []string{
		"source", "secret", "kustomization", "helmrelease",
		"image", "alert", "alert-provider", "receiver", "tenant",
	}

	for _, expectedCmd := range expectedSubCommands {
		found := false
		for _, subCmd := range subCommands {
			if subCmd.Use == expectedCmd {
				found = true

				break
			}
		}
		require.True(t, found, "Expected sub-command %s not found", expectedCmd)
	}
}

func TestCreateSourceCommand(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams)
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
	require.Equal(t, "Create or update Flux sources", sourceCmd.Short)

	// Check for source sub-commands
	sourceSubCommands := sourceCmd.Commands()
	require.NotEmpty(t, sourceSubCommands)

	expectedSourceSubCommands := []string{"git", "helm", "bucket", "chart", "oci"}

	for _, expectedCmd := range expectedSourceSubCommands {
		found := false
		for _, subCmd := range sourceSubCommands {
			if subCmd.Use == expectedCmd {
				found = true

				break
			}
		}
		require.True(t, found, "Expected source sub-command %s not found", expectedCmd)
	}
}

func TestCreateSecretCommand(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams)
	cmd := client.CreateCreateCommand("")

	// Find secret command
	var secretCmd *cobra.Command

	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "secret" {
			secretCmd = subCmd

			break
		}
	}

	require.NotNil(t, secretCmd)
	require.Equal(t, "Create or update Flux secrets", secretCmd.Short)

	// Check for secret sub-commands
	secretSubCommands := secretCmd.Commands()
	require.NotEmpty(t, secretSubCommands)

	expectedSecretSubCommands := []string{
		"git", "helm", "oci", "tls", "github-app", "notation", "proxy",
	}

	for _, expectedCmd := range expectedSecretSubCommands {
		found := false
		for _, subCmd := range secretSubCommands {
			if subCmd.Use == expectedCmd {
				found = true

				break
			}
		}
		require.True(t, found, "Expected secret sub-command %s not found", expectedCmd)
	}
}

func TestCreateImageCommand(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}

	client := flux.NewClient(ioStreams)
	cmd := client.CreateCreateCommand("")

	// Find image command
	var imageCmd *cobra.Command

	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "image" {
			imageCmd = subCmd

			break
		}
	}

	require.NotNil(t, imageCmd)
	require.Equal(t, "Create or update Flux image automation objects", imageCmd.Short)

	// Check for image sub-commands
	imageSubCommands := imageCmd.Commands()
	require.NotEmpty(t, imageSubCommands)

	expectedImageSubCommands := []string{"repository", "policy", "update"}

	for _, expectedCmd := range expectedImageSubCommands {
		found := false
		for _, subCmd := range imageSubCommands {
			if subCmd.Use == expectedCmd {
				found = true

				break
			}
		}
		require.True(t, found, "Expected image sub-command %s not found", expectedCmd)
	}
}
