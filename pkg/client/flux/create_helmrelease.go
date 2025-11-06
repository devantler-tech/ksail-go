package flux

import (
	"context"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type helmReleaseFlags struct {
	sourceKind      string
	sourceName      string
	sourceNamespace string
	chart           string
	chartVersion    string
	targetNamespace string
	createNamespace bool
	interval        time.Duration
	export          bool
	dependsOn       []string
}

func (c *Client) newCreateHelmReleaseCmd() *cobra.Command {
	flags := &helmReleaseFlags{
		interval: time.Minute,
	}

	cmd := &cobra.Command{
		Use:     "helmrelease [name]",
		Aliases: []string{"hr"},
		Short:   "Create or update a HelmRelease resource",
		Long:    "Create or update a HelmRelease resource using Flux APIs",
		Example: `  # Create a HelmRelease with a chart from a HelmRepository source
  ksail workload create helmrelease podinfo \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --chart-version=6.6.2 \
    --namespace=flux-system

  # Create a HelmRelease targeting a different namespace
  ksail workload create helmrelease podinfo \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --target-namespace=production \
    --create-target-namespace=true`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			namespace := cmd.Flag("namespace").Value.String()
			if namespace == "" {
				namespace = "flux-system"
			}

			return c.createHelmRelease(cmd.Context(), name, namespace, flags)
		},
	}

	cmd.Flags().
		StringVar(&flags.sourceKind, "source-kind", "HelmRepository", "source kind (HelmRepository, GitRepository, Bucket)")
	cmd.Flags().
		StringVar(&flags.sourceName, "source", "", "source name in format 'Kind/name' or 'Kind/name.namespace'")
	cmd.Flags().StringVar(&flags.chart, "chart", "", "Helm chart name or path")
	cmd.Flags().StringVar(&flags.chartVersion, "chart-version", "", "Helm chart version")
	cmd.Flags().
		StringVar(&flags.targetNamespace, "target-namespace", "", "namespace to install the Helm release")
	cmd.Flags().
		BoolVar(&flags.createNamespace, "create-target-namespace", false, "create the target namespace if it doesn't exist")
	cmd.Flags().DurationVar(&flags.interval, "interval", time.Minute, "reconciliation interval")
	cmd.Flags().BoolVar(&flags.export, "export", false, "export in YAML format to stdout")
	cmd.Flags().
		StringSliceVar(&flags.dependsOn, "depends-on", nil, "HelmRelease that must be ready before this one")

	_ = cmd.MarkFlagRequired("source")
	_ = cmd.MarkFlagRequired("chart")

	return cmd
}

func (c *Client) createHelmRelease(
	ctx context.Context,
	name, namespace string,
	flags *helmReleaseFlags,
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

	helmRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: flags.interval},
			Chart: &helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: flags.chart,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourceKind,
						Name:      sourceName,
						Namespace: sourceNs,
					},
				},
			},
		},
	}

	if flags.chartVersion != "" {
		helmRelease.Spec.Chart.Spec.Version = flags.chartVersion
	}

	if flags.targetNamespace != "" {
		helmRelease.Spec.TargetNamespace = flags.targetNamespace
	}

	if flags.createNamespace {
		helmRelease.Spec.Install = &helmv2.Install{
			CreateNamespace: true,
		}
	}

	// Set dependencies
	if len(flags.dependsOn) > 0 {
		deps := make([]helmv2.DependencyReference, 0, len(flags.dependsOn))
		for _, dep := range flags.dependsOn {
			depName := dep
			depNs := namespace
			if strings.Contains(dep, "/") {
				parts := strings.SplitN(dep, "/", 2)
				depNs = parts[0]
				depName = parts[1]
			}
			deps = append(deps, helmv2.DependencyReference{
				Name:      depName,
				Namespace: depNs,
			})
		}
		helmRelease.Spec.DependsOn = deps
	}

	// Export mode
	if flags.export {
		return c.exportResource(helmRelease)
	}

	// Create or update the resource
	return c.upsertResource(ctx, helmRelease, &helmv2.HelmRelease{}, "HelmRelease")
}
