package flux

import (
	"context"
	"fmt"
	"strings"
	"time"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type kustomizationFlags struct {
	sourceKind      string
	sourceName      string
	sourceNamespace string
	path            string
	prune           bool
	targetNamespace string
	interval        time.Duration
	export          bool
	wait            bool
	dependsOn       []string
}

func (c *Client) newCreateKustomizationCmd() *cobra.Command {
	flags := &kustomizationFlags{
		interval: time.Minute,
		path:     "./",
	}

	cmd := &cobra.Command{
		Use:   "kustomization [name]",
		Short: "Create or update a Kustomization resource",
		Long:  "Create or update a Kustomization resource using Flux APIs",
		Example: `  # Create a Kustomization from a GitRepository source
  ksail workload create kustomization podinfo \
    --source=GitRepository/podinfo \
    --path="./kustomize" \
    --prune=true \
    --interval=5m

  # Create a Kustomization with a target namespace
  ksail workload create kustomization podinfo \
    --source=GitRepository/podinfo \
    --path="./kustomize" \
    --target-namespace=default \
    --prune=true`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			namespace := cmd.Flag("namespace").Value.String()
			if namespace == "" {
				namespace = "flux-system"
			}

			return c.createKustomization(cmd.Context(), name, namespace, flags)
		},
	}

	cmd.Flags().StringVar(&flags.sourceKind, "source-kind", "GitRepository", "source kind (GitRepository, OCIRepository, Bucket)")
	cmd.Flags().StringVar(&flags.sourceName, "source", "", "source name in format 'Kind/name' or 'Kind/name.namespace'")
	cmd.Flags().StringVar(&flags.path, "path", "./", "path to the directory containing a kustomization.yaml file")
	cmd.Flags().BoolVar(&flags.prune, "prune", false, "enable garbage collection")
	cmd.Flags().BoolVar(&flags.wait, "wait", false, "enable health checking")
	cmd.Flags().StringVar(&flags.targetNamespace, "target-namespace", "", "overrides the namespace of all Kustomization objects")
	cmd.Flags().DurationVar(&flags.interval, "interval", time.Minute, "reconciliation interval")
	cmd.Flags().BoolVar(&flags.export, "export", false, "export in YAML format to stdout")
	cmd.Flags().StringSliceVar(&flags.dependsOn, "depends-on", nil, "Kustomization that must be ready before this one")

	_ = cmd.MarkFlagRequired("source")

	return cmd
}

func (c *Client) createKustomization(
	ctx context.Context,
	name, namespace string,
	flags *kustomizationFlags,
) error {
	// Parse source
	sourceKind := flags.sourceKind
	sourceName := flags.sourceName
	sourceNs := namespace

	if strings.Contains(sourceName, "/") {
		parts := strings.SplitN(sourceName, "/", 2)
		sourceKind = parts[0]
		sourceName = parts[1]
	}

	if strings.Contains(sourceName, ".") {
		parts := strings.SplitN(sourceName, ".", 2)
		sourceName = parts[0]
		sourceNs = parts[1]
	}

	kustomization := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kustomizev1.KustomizationSpec{
			Interval: metav1.Duration{Duration: flags.interval},
			Path:     flags.path,
			Prune:    flags.prune,
			Wait:     flags.wait,
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind:      sourceKind,
				Name:      sourceName,
				Namespace: sourceNs,
			},
		},
	}

	if flags.targetNamespace != "" {
		kustomization.Spec.TargetNamespace = flags.targetNamespace
	}

	// Set dependencies
	if len(flags.dependsOn) > 0 {
		deps := make([]kustomizev1.DependencyReference, 0, len(flags.dependsOn))
		for _, dep := range flags.dependsOn {
			depName := dep
			depNs := namespace
			if strings.Contains(dep, "/") {
				parts := strings.SplitN(dep, "/", 2)
				depNs = parts[0]
				depName = parts[1]
			}
			deps = append(deps, kustomizev1.DependencyReference{
				Name:      depName,
				Namespace: depNs,
			})
		}
		kustomization.Spec.DependsOn = deps
	}

	// Export mode
	if flags.export {
		return c.exportResource(kustomization)
	}

	// Create or update the resource
	k8sClient, err := c.getClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	err = k8sClient.Create(ctx, kustomization)
	if err != nil {
		if client.IgnoreAlreadyExists(err) == nil {
			// Resource exists, update it
			existing := &kustomizev1.Kustomization{}
			if err := k8sClient.Get(ctx, client.ObjectKey{
				Name:      name,
				Namespace: namespace,
			}, existing); err != nil {
				return fmt.Errorf("failed to get existing Kustomization: %w", err)
			}

			existing.Spec = kustomization.Spec
			if err := k8sClient.Update(ctx, existing); err != nil {
				return fmt.Errorf("failed to update Kustomization: %w", err)
			}

			fmt.Fprintf(c.ioStreams.Out, "✓ Kustomization %s/%s updated\n", namespace, name)
			return nil
		}
		return fmt.Errorf("failed to create Kustomization: %w", err)
	}

	fmt.Fprintf(c.ioStreams.Out, "✓ Kustomization %s/%s created\n", namespace, name)
	return nil
}
