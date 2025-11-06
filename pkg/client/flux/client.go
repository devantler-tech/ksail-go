// Package flux provides a flux client implementation using Flux Kubernetes APIs.
package flux

import (
	"fmt"
	"io"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// Client wraps flux API functionality.
type Client struct {
	ioStreams      genericiooptions.IOStreams
	kubeconfigPath string
	client         client.Client
}

// NewClient creates a new flux client instance.
func NewClient(ioStreams genericiooptions.IOStreams, kubeconfigPath string) *Client {
	return &Client{
		ioStreams:      ioStreams,
		kubeconfigPath: kubeconfigPath,
	}
}

// getClient returns a controller-runtime client configured for Flux APIs.
func (c *Client) getClient() (client.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	config, err := c.getRestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := sourcev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add source-controller scheme: %w", err)
	}
	if err := kustomizev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add kustomize-controller scheme: %w", err)
	}
	if err := helmv2.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add helm-controller scheme: %w", err)
	}

	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	c.client = k8sClient
	return k8sClient, nil
}

// getRestConfig returns a REST config for the Kubernetes cluster.
func (c *Client) getRestConfig() (*rest.Config, error) {
	if c.kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", c.kubeconfigPath)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)
	return kubeConfig.ClientConfig()
}

// exportResource exports a resource as YAML to stdout.
func (c *Client) exportResource(obj runtime.Object) error {
	data, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	_, err = io.WriteString(c.ioStreams.Out, string(data))
	return err
}

// CreateCreateCommand creates a flux create command that uses Flux Kubernetes APIs.
func (c *Client) CreateCreateCommand(kubeconfigPath string) *cobra.Command {
	c.kubeconfigPath = kubeconfigPath

	createCmd := &cobra.Command{
		Use:   "flux-create",
		Short: "Create Flux resources",
		Long:  "Create or update Flux sources and resources using Kubernetes APIs.",
	}

	// Add namespace flag to all commands
	createCmd.PersistentFlags().StringP("namespace", "n", "flux-system", "the namespace scope for this operation")

	// Add sub-commands for flux create
	createCmd.AddCommand(c.createSourceCommand())
	createCmd.AddCommand(c.newCreateKustomizationCmd())
	createCmd.AddCommand(c.newCreateHelmReleaseCmd())

	return createCmd
}

// createSourceCommand creates the flux create source command.
func (c *Client) createSourceCommand() *cobra.Command {
	sourceCmd := &cobra.Command{
		Use:   "source",
		Short: "Create or update Flux sources",
	}

	// Add source sub-commands
	sourceCmd.AddCommand(c.newCreateSourceGitCmd())
	sourceCmd.AddCommand(c.newCreateSourceHelmCmd())
	sourceCmd.AddCommand(c.newCreateSourceOCICmd())

	return sourceCmd
}
