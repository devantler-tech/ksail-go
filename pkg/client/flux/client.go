// Package flux provides a flux client implementation that wraps the flux CLI.
package flux

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// Client wraps flux command functionality.
type Client struct {
	ioStreams genericiooptions.IOStreams
	fluxBin   string
}

// NewClient creates a new flux client instance.
func NewClient(ioStreams genericiooptions.IOStreams) *Client {
	// Try to find flux binary in PATH
	fluxPath, err := exec.LookPath("flux")
	if err != nil {
		// flux not found, we'll provide a helpful error later
		fluxPath = "flux"
	}

	return &Client{
		ioStreams: ioStreams,
		fluxBin:   fluxPath,
	}
}

// replaceFluxInExamples replaces "flux" with "ksail workload" in command examples.
func replaceFluxInExamples(cmd *cobra.Command) {
	if cmd.Example != "" {
		cmd.Example = strings.ReplaceAll(cmd.Example, "flux", "ksail workload")
	}
	// Recursively update examples for sub-commands
	for _, subCmd := range cmd.Commands() {
		replaceFluxInExamples(subCmd)
	}
}

// CreateCreateCommand creates a flux create command wrapper that executes flux CLI commands.
func (c *Client) CreateCreateCommand(kubeConfigPath string) *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "flux-create",
		Short: "Create Flux resources",
		Long:  "Create or update Flux sources and resources.",
	}

	// Add sub-commands for flux create
	createCmd.AddCommand(c.createSourceCommand(kubeConfigPath))
	createCmd.AddCommand(c.createSecretCommand(kubeConfigPath))
	createCmd.AddCommand(c.createKustomizationCommand(kubeConfigPath))
	createCmd.AddCommand(c.createHelmReleaseCommand(kubeConfigPath))
	createCmd.AddCommand(c.createImageCommand(kubeConfigPath))
	createCmd.AddCommand(c.createAlertCommand(kubeConfigPath))
	createCmd.AddCommand(c.createAlertProviderCommand(kubeConfigPath))
	createCmd.AddCommand(c.createReceiverCommand(kubeConfigPath))
	createCmd.AddCommand(c.createTenantCommand(kubeConfigPath))

	replaceFluxInExamples(createCmd)

	return createCmd
}

// createFluxCommand creates a generic flux command wrapper.
func (c *Client) createFluxCommand(
	name string, use string, short string, kubeConfigPath string,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  short, // Use short description as long for consistency
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if flux is available before executing
			_, lookupErr := exec.LookPath("flux")
			if lookupErr != nil {
				cmd.PrintErrln("Error: flux CLI is not installed or not in PATH")
				cmd.PrintErrln("Please install flux from https://fluxcd.io/flux/installation/")

				return fmt.Errorf("flux CLI not found: %w", lookupErr)
			}

			// Build the flux command
			fluxArgs := []string{"create", name}
			fluxArgs = append(fluxArgs, args...)

			// Add kubeconfig if provided
			if kubeConfigPath != "" {
				fluxArgs = append([]string{"--kubeconfig", kubeConfigPath}, fluxArgs...)
			}

			// Execute flux command
			// #nosec G204 - flux binary path is controlled and fluxArgs are from validated cobra args
			fluxCmd := exec.Command(c.fluxBin, fluxArgs...)
			fluxCmd.Stdin = os.Stdin
			fluxCmd.Stdout = c.ioStreams.Out
			fluxCmd.Stderr = c.ioStreams.ErrOut

			return fluxCmd.Run()
		},
		// Don't disable flag parsing for help to work
		// DisableFlagParsing: true, // Let flux handle all flags
	}

	return cmd
}

// createSourceCommand creates the flux create source command.
func (c *Client) createSourceCommand(kubeConfigPath string) *cobra.Command {
	sourceCmd := &cobra.Command{
		Use:   "source",
		Short: "Create or update Flux sources",
	}

	// Add source sub-commands
	sourceCmd.AddCommand(
		c.createFluxCommand(
			"source git",
			"git",
			"Create or update a GitRepository source",
			kubeConfigPath,
		),
	)
	sourceCmd.AddCommand(
		c.createFluxCommand(
			"source helm",
			"helm",
			"Create or update a HelmRepository source",
			kubeConfigPath,
		),
	)
	sourceCmd.AddCommand(
		c.createFluxCommand(
			"source bucket",
			"bucket",
			"Create or update a Bucket source",
			kubeConfigPath,
		),
	)
	sourceCmd.AddCommand(
		c.createFluxCommand(
			"source chart",
			"chart",
			"Create or update a HelmChart source",
			kubeConfigPath,
		),
	)
	sourceCmd.AddCommand(
		c.createFluxCommand(
			"source oci",
			"oci",
			"Create or update an OCIRepository source",
			kubeConfigPath,
		),
	)

	return sourceCmd
}

// createSecretCommand creates the flux create secret command.
func (c *Client) createSecretCommand(kubeConfigPath string) *cobra.Command {
	secretCmd := &cobra.Command{
		Use:   "secret",
		Short: "Create or update Flux secrets",
	}

	// Add secret sub-commands
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret git",
			"git",
			"Create or update a Kubernetes secret for Git authentication",
			kubeConfigPath,
		),
	)
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret helm",
			"helm",
			"Create or update a Kubernetes secret for Helm repository authentication",
			kubeConfigPath,
		),
	)
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret oci",
			"oci",
			"Create or update a Kubernetes secret for OCI authentication",
			kubeConfigPath,
		),
	)
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret tls",
			"tls",
			"Create or update a Kubernetes secret with TLS certificates",
			kubeConfigPath,
		),
	)
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret github-app",
			"github-app",
			"Create or update a Kubernetes secret for GitHub App authentication",
			kubeConfigPath,
		),
	)
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret notation",
			"notation",
			"Create or update a Kubernetes secret for Notation trust policy",
			kubeConfigPath,
		),
	)
	secretCmd.AddCommand(
		c.createFluxCommand(
			"secret proxy",
			"proxy",
			"Create or update a Kubernetes secret for proxy authentication",
			kubeConfigPath,
		),
	)

	return secretCmd
}

// createKustomizationCommand creates the flux create kustomization command.
func (c *Client) createKustomizationCommand(kubeConfigPath string) *cobra.Command {
	return c.createFluxCommand(
		"kustomization",
		"kustomization",
		"Create or update a Kustomization resource",
		kubeConfigPath,
	)
}

// createHelmReleaseCommand creates the flux create helmrelease command.
func (c *Client) createHelmReleaseCommand(kubeConfigPath string) *cobra.Command {
	return c.createFluxCommand(
		"helmrelease",
		"helmrelease",
		"Create or update a HelmRelease resource",
		kubeConfigPath,
	)
}

// createImageCommand creates the flux create image command.
func (c *Client) createImageCommand(kubeConfigPath string) *cobra.Command {
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Create or update Flux image automation objects",
	}

	// Add image sub-commands
	imageCmd.AddCommand(
		c.createFluxCommand(
			"image repository",
			"repository",
			"Create or update an ImageRepository object",
			kubeConfigPath,
		),
	)
	imageCmd.AddCommand(
		c.createFluxCommand(
			"image policy",
			"policy",
			"Create or update an ImagePolicy object",
			kubeConfigPath,
		),
	)
	imageCmd.AddCommand(
		c.createFluxCommand(
			"image update",
			"update",
			"Create or update an ImageUpdateAutomation object",
			kubeConfigPath,
		),
	)

	return imageCmd
}

// createAlertCommand creates the flux create alert command.
func (c *Client) createAlertCommand(kubeConfigPath string) *cobra.Command {
	return c.createFluxCommand(
		"alert",
		"alert",
		"Create or update a Alert resource",
		kubeConfigPath,
	)
}

// createAlertProviderCommand creates the flux create alert-provider command.
func (c *Client) createAlertProviderCommand(kubeConfigPath string) *cobra.Command {
	return c.createFluxCommand(
		"alert-provider",
		"alert-provider",
		"Create or update a Provider resource",
		kubeConfigPath,
	)
}

// createReceiverCommand creates the flux create receiver command.
func (c *Client) createReceiverCommand(kubeConfigPath string) *cobra.Command {
	return c.createFluxCommand(
		"receiver",
		"receiver",
		"Create or update a Receiver resource",
		kubeConfigPath,
	)
}

// createTenantCommand creates the flux create tenant command.
func (c *Client) createTenantCommand(kubeConfigPath string) *cobra.Command {
	return c.createFluxCommand(
		"tenant",
		"tenant",
		"Create or update a tenant",
		kubeConfigPath,
	)
}
