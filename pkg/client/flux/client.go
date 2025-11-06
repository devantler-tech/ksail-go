// Package flux provides a flux client implementation using Flux Kubernetes APIs.
package flux

import (
	"context"
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

// upsertResource creates or updates a Flux resource using the Kubernetes API.
// It takes a context, the resource object to create/update, the resource name,
// namespace, and a resourceKind string for logging purposes.
// The obj parameter must be a pointer to a Flux resource (e.g., *sourcev1.GitRepository).
// The existing parameter must be a pointer to the same type for fetching existing resources.
func (c *Client) upsertResource(
	ctx context.Context,
	obj client.Object,
	existing client.Object,
	resourceKind string,
) error {
	name := obj.GetName()
	namespace := obj.GetNamespace()

	// Get Kubernetes client
	k8sClient, err := c.getClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Try to create the resource
	err = k8sClient.Create(ctx, obj)
	if err != nil {
		if client.IgnoreAlreadyExists(err) == nil {
			// Resource exists, update it
			if err := k8sClient.Get(ctx, client.ObjectKey{
				Name:      name,
				Namespace: namespace,
			}, existing); err != nil {
				return fmt.Errorf("failed to get existing %s: %w", resourceKind, err)
			}

			// Copy spec from obj to existing
			// This assumes the objects have a Spec field that can be copied
			if err := copySpec(obj, existing); err != nil {
				return fmt.Errorf("failed to copy spec: %w", err)
			}

			if err := k8sClient.Update(ctx, existing); err != nil {
				return fmt.Errorf("failed to update %s: %w", resourceKind, err)
			}

			fmt.Fprintf(c.ioStreams.Out, "✓ %s %s/%s updated\n", resourceKind, namespace, name)
			return nil
		}
		return fmt.Errorf("failed to create %s: %w", resourceKind, err)
	}

	fmt.Fprintf(c.ioStreams.Out, "✓ %s %s/%s created\n", resourceKind, namespace, name)
	return nil
}

// copySpec copies the Spec field from src to dst using type assertions.
func copySpec(src, dst client.Object) error {
	switch s := src.(type) {
	case *sourcev1.GitRepository:
		d, ok := dst.(*sourcev1.GitRepository)
		if !ok {
			return fmt.Errorf("type mismatch: expected *sourcev1.GitRepository")
		}
		d.Spec = s.Spec
	case *sourcev1.HelmRepository:
		d, ok := dst.(*sourcev1.HelmRepository)
		if !ok {
			return fmt.Errorf("type mismatch: expected *sourcev1.HelmRepository")
		}
		d.Spec = s.Spec
	case *sourcev1.OCIRepository:
		d, ok := dst.(*sourcev1.OCIRepository)
		if !ok {
			return fmt.Errorf("type mismatch: expected *sourcev1.OCIRepository")
		}
		d.Spec = s.Spec
	case *kustomizev1.Kustomization:
		d, ok := dst.(*kustomizev1.Kustomization)
		if !ok {
			return fmt.Errorf("type mismatch: expected *kustomizev1.Kustomization")
		}
		d.Spec = s.Spec
	case *helmv2.HelmRelease:
		d, ok := dst.(*helmv2.HelmRelease)
		if !ok {
			return fmt.Errorf("type mismatch: expected *helmv2.HelmRelease")
		}
		d.Spec = s.Spec
	default:
		return fmt.Errorf("unsupported resource type: %T", src)
	}
	return nil
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
	createCmd.PersistentFlags().
		StringP("namespace", "n", "flux-system", "the namespace scope for this operation")

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
