package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	defaultInterval = 1 * time.Minute
)

// NewHelmReleaseCmd creates the workload gen helm-release command.
func NewHelmReleaseCmd(_ *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "helm-release [NAME]",
		Aliases: []string{"hr", "helmrelease"},
		Short:   "Generate a HelmRelease resource",
		Long:    "Generate a HelmRelease resource for a given HelmRepository, GitRepository, Bucket, or chart reference source.",
		Example: `  # Generate a HelmRelease with a chart from a HelmRepository source
  ksail workload gen helm-release podinfo \
    --interval=10m \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --chart-version=">4.0.0" \
    --export

  # Generate a HelmRelease with a chart from a GitRepository source
  ksail workload gen helm-release podinfo \
    --interval=10m \
    --source=GitRepository/podinfo \
    --chart=./charts/podinfo \
    --export

  # Generate a HelmRelease with values from local YAML files
  ksail workload gen helm-release podinfo \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --values=./my-values1.yaml \
    --values=./my-values2.yaml \
    --export

  # Generate a HelmRelease with values from a Kubernetes secret
  ksail workload gen helm-release podinfo \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --values-from=Secret/my-secret-values \
    --export

  # Generate a HelmRelease with a custom release name
  ksail workload gen helm-release podinfo \
    --release-name=podinfo-dev \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --export

  # Generate a HelmRelease targeting another namespace
  ksail workload gen helm-release podinfo \
    --target-namespace=test \
    --create-target-namespace=true \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --export

  # Generate a HelmRelease using a chart reference
  ksail workload gen helm-release podinfo \
    --namespace=default \
    --chart-ref=HelmChart/podinfo.flux-system \
    --export`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tmr := timer.New()
			tmr.Start()

			// Read flags
			name := args[0]
			namespace, _ := cmd.Flags().GetString("namespace")
			source, _ := cmd.Flags().GetString("source")
			chart, _ := cmd.Flags().GetString("chart")
			chartVersion, _ := cmd.Flags().GetString("chart-version")
			chartRef, _ := cmd.Flags().GetString("chart-ref")
			targetNamespace, _ := cmd.Flags().GetString("target-namespace")
			storageNamespace, _ := cmd.Flags().GetString("storage-namespace")
			createNamespace, _ := cmd.Flags().GetBool("create-target-namespace")
			dependsOn, _ := cmd.Flags().GetStringSlice("depends-on")
			interval, _ := cmd.Flags().GetDuration("interval")
			timeout, _ := cmd.Flags().GetDuration("timeout")
			values, _ := cmd.Flags().GetStringSlice("values")
			valuesFrom, _ := cmd.Flags().GetStringSlice("values-from")
			saName, _ := cmd.Flags().GetString("service-account")
			crdsPolicy, _ := cmd.Flags().GetString("crds")
			kubeConfigSecretRef, _ := cmd.Flags().GetString("kubeconfig-secret-ref")
			releaseName, _ := cmd.Flags().GetString("release-name")
			export, _ := cmd.Flags().GetBool("export")

			if err := validateHelmReleaseArgs(source, chart, chartRef, crdsPolicy); err != nil {
				return err
			}

			helmRelease, err := buildHelmRelease(name, namespace, source, chart, chartVersion, chartRef,
				targetNamespace, storageNamespace, createNamespace, dependsOn, interval, timeout,
				values, valuesFrom, saName, crdsPolicy, kubeConfigSecretRef, releaseName)
			if err != nil {
				return err
			}

			generator := yamlgenerator.NewYAMLGenerator[helmv2.HelmRelease]()
			opts := yamlgenerator.Options{
				Output: "",
				Force:  false,
			}

			yaml, err := generator.Generate(*helmRelease, opts)
			if err != nil {
				return fmt.Errorf("failed to generate HelmRelease YAML: %w", err)
			}

			if export {
				fmt.Fprint(cmd.OutOrStdout(), yaml)
			} else {
				// TODO: Apply the HelmRelease to the cluster
				return fmt.Errorf("applying HelmRelease to cluster is not yet implemented, use --export flag")
			}

			total, stage := tmr.GetTiming()
			timingStr := notify.FormatTiming(total, stage, false)
			notify.WriteMessage(notify.Message{
				Type:    notify.SuccessType,
				Content: "generated HelmRelease %s",
				Args:    []any{timingStr},
				Writer:  cmd.OutOrStdout(),
			})

			return nil
		},
		SilenceUsage: true,
	}

	flags := cmd.Flags()

	// Required flags
	flags.StringVar(&helmReleaseArgs.source, "source", "", "source that contains the chart (HelmRepository/name, GitRepository/name, Bucket/name)")
	flags.StringVar(&helmReleaseArgs.chart, "chart", "", "Helm chart name or path")
	flags.StringVar(&helmReleaseArgs.chartRef, "chart-ref", "", "Helm chart reference (HelmChart/name, OCIRepository/name)")

	// Optional flags
	flags.StringVarP(&helmReleaseArgs.namespace, "namespace", "n", "default", "namespace scope for the HelmRelease")
	flags.DurationVar(&helmReleaseArgs.interval, "interval", defaultInterval, "reconciliation interval")
	flags.StringVar(&helmReleaseArgs.chartVersion, "chart-version", "", "Helm chart version, accepts a semver range")
	flags.StringVar(&helmReleaseArgs.targetNamespace, "target-namespace", "", "namespace to target when performing operations")
	flags.StringVar(&helmReleaseArgs.storageNamespace, "storage-namespace", "", "namespace for Helm storage")
	flags.BoolVar(&helmReleaseArgs.createNamespace, "create-target-namespace", false, "create the target namespace if not present")
	flags.StringSliceVar(&helmReleaseArgs.dependsOn, "depends-on", nil, "HelmReleases that must be ready before this release")
	flags.DurationVar(&helmReleaseArgs.timeout, "timeout", 5*time.Minute, "timeout for any individual Kubernetes operation")
	flags.StringSliceVar(&helmReleaseArgs.values, "values", nil, "local values YAML files")
	flags.StringSliceVar(&helmReleaseArgs.valuesFrom, "values-from", nil, "values from ConfigMap or Secret")
	flags.StringVar(&helmReleaseArgs.saName, "service-account", "", "service account name to impersonate")
	flags.StringVar(&helmReleaseArgs.crdsPolicy, "crds", "", "CRDs policy (Create, CreateReplace, Skip)")
	flags.StringVar(&helmReleaseArgs.kubeConfigSecretRef, "kubeconfig-secret-ref", "", "KubeConfig secret reference for remote reconciliation")
	flags.StringVar(&helmReleaseArgs.releaseName, "release-name", "", "name used for the Helm release")
	flags.BoolVar(&helmReleaseArgs.export, "export", false, "export in YAML format to stdout")

	return cmd
}

func validateHelmReleaseArgs(source, chart, chartRef, crdsPolicy string) error {
	// Either source+chart or chartRef must be specified
	hasSource := source != "" && chart != ""
	hasChartRef := chartRef != ""

	if !hasSource && !hasChartRef {
		return fmt.Errorf("either --source with --chart or --chart-ref must be specified")
	}

	if hasSource && hasChartRef {
		return fmt.Errorf("cannot specify both --source/--chart and --chart-ref")
	}

	// Validate CRDs policy if specified
	if crdsPolicy != "" {
		validPolicies := []string{"Create", "CreateReplace", "Skip"}
		found := false
		for _, p := range validPolicies {
			if crdsPolicy == p {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid crds policy %q, must be one of: %s", crdsPolicy, strings.Join(validPolicies, ", "))
		}
	}

	return nil
}

func buildHelmRelease(name, namespace, source, chart, chartVersion, chartRef,
	targetNamespace, storageNamespace string, createNamespace bool, dependsOn []string,
	interval, timeout time.Duration, values, valuesFrom []string, saName, crdsPolicy,
	kubeConfigSecretRef, releaseName string) (*helmv2.HelmRelease, error) {
	helmRelease := &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			APIVersion: helmv2.GroupVersion.String(),
			Kind:       helmv2.HelmReleaseKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Interval: metav1.Duration{Duration: interval},
		},
	}

	// Set chart or chartRef
	if source != "" && chart != "" {
		parts := strings.Split(source, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid source format, expected Kind/name or Kind/name.namespace")
		}

		sourceKind := parts[0]
		sourceName := parts[1]
		sourceNamespace := namespace

		// Check if namespace is included in source name
		if strings.Contains(sourceName, ".") {
			nameParts := strings.SplitN(sourceName, ".", 2)
			sourceName = nameParts[0]
			sourceNamespace = nameParts[1]
		}

		// Validate source kind
		validSourceKinds := []string{sourcev1.HelmRepositoryKind, sourcev1.GitRepositoryKind, sourcev1.BucketKind}
		found := false
		for _, kind := range validSourceKinds {
			if sourceKind == kind {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid source kind %q, must be one of: %s", sourceKind, strings.Join(validSourceKinds, ", "))
		}

		helmRelease.Spec.Chart = &helmv2.HelmChartTemplate{
			Spec: helmv2.HelmChartTemplateSpec{
				Chart: helmReleaseArgs.chart,
				SourceRef: helmv2.CrossNamespaceObjectReference{
					Kind:      sourceKind,
					Name:      sourceName,
					Namespace: sourceNamespace,
				},
			},
		}

		if helmReleaseArgs.chartVersion != "" {
			helmRelease.Spec.Chart.Spec.Version = helmReleaseArgs.chartVersion
		}
	} else if helmReleaseArgs.chartRef != "" {
		parts := strings.Split(helmReleaseArgs.chartRef, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid chart-ref format, expected Kind/name or Kind/name.namespace")
		}

		chartRefKind := parts[0]
		chartRefName := parts[1]
		chartRefNamespace := helmReleaseArgs.namespace

		// Check if namespace is included
		if strings.Contains(chartRefName, ".") {
			nameParts := strings.SplitN(chartRefName, ".", 2)
			chartRefName = nameParts[0]
			chartRefNamespace = nameParts[1]
		}

		// Validate chartRef kind
		validChartRefKinds := []string{sourcev1.OCIRepositoryKind, sourcev1.HelmChartKind}
		found := false
		for _, kind := range validChartRefKinds {
			if chartRefKind == kind {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid chart-ref kind %q, must be one of: %s", chartRefKind, strings.Join(validChartRefKinds, ", "))
		}

		helmRelease.Spec.ChartRef = &helmv2.CrossNamespaceSourceReference{
			Kind:      chartRefKind,
			Name:      chartRefName,
			Namespace: chartRefNamespace,
		}
	}

	// Set optional fields
	if helmReleaseArgs.releaseName != "" {
		helmRelease.Spec.ReleaseName = helmReleaseArgs.releaseName
	}

	if helmReleaseArgs.targetNamespace != "" {
		helmRelease.Spec.TargetNamespace = helmReleaseArgs.targetNamespace
	}

	if helmReleaseArgs.storageNamespace != "" {
		helmRelease.Spec.StorageNamespace = helmReleaseArgs.storageNamespace
	}

	if helmReleaseArgs.createNamespace {
		if helmRelease.Spec.Install == nil {
			helmRelease.Spec.Install = &helmv2.Install{}
		}
		helmRelease.Spec.Install.CreateNamespace = true
	}

	if len(helmReleaseArgs.dependsOn) > 0 {
		dependsOn := []helmv2.DependencyReference{}
		for _, dep := range helmReleaseArgs.dependsOn {
			parts := strings.Split(dep, "/")
			if len(parts) == 1 {
				// Same namespace
				dependsOn = append(dependsOn, helmv2.DependencyReference{
					Name: parts[0],
				})
			} else if len(parts) == 2 {
				// Different namespace
				dependsOn = append(dependsOn, helmv2.DependencyReference{
					Namespace: parts[0],
					Name:      parts[1],
				})
			} else {
				return nil, fmt.Errorf("invalid depends-on format %q, expected name or namespace/name", dep)
			}
		}
		helmRelease.Spec.DependsOn = dependsOn
	}

	if helmReleaseArgs.timeout > 0 {
		helmRelease.Spec.Timeout = &metav1.Duration{Duration: helmReleaseArgs.timeout}
	}

	if helmReleaseArgs.saName != "" {
		helmRelease.Spec.ServiceAccountName = helmReleaseArgs.saName
	}

	if helmReleaseArgs.crdsPolicy != "" {
		if helmRelease.Spec.Install == nil {
			helmRelease.Spec.Install = &helmv2.Install{}
		}
		helmRelease.Spec.Install.CRDs = helmv2.Create

		if helmRelease.Spec.Upgrade == nil {
			helmRelease.Spec.Upgrade = &helmv2.Upgrade{}
		}
		helmRelease.Spec.Upgrade.CRDs = helmv2.CRDsPolicy(helmReleaseArgs.crdsPolicy)
	}

	if helmReleaseArgs.kubeConfigSecretRef != "" {
		helmRelease.Spec.KubeConfig = &meta.KubeConfigReference{
			SecretRef: &meta.SecretKeyReference{
				Name: helmReleaseArgs.kubeConfigSecretRef,
			},
		}
	}

	// Handle values files
	if len(helmReleaseArgs.values) > 0 {
		valuesMap := make(map[string]interface{})
		for _, vFile := range helmReleaseArgs.values {
			data, err := os.ReadFile(vFile)
			if err != nil {
				return nil, fmt.Errorf("reading values from %s: %w", vFile, err)
			}

			jsonBytes, err := yaml.YAMLToJSON(data)
			if err != nil {
				return nil, fmt.Errorf("converting values to JSON from %s: %w", vFile, err)
			}

			jsonMap := make(map[string]interface{})
			if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
				return nil, fmt.Errorf("unmarshaling values from %s: %w", vFile, err)
			}

			valuesMap = mergeMaps(valuesMap, jsonMap)
		}

		jsonRaw, err := json.Marshal(valuesMap)
		if err != nil {
			return nil, fmt.Errorf("marshaling values: %w", err)
		}

		helmRelease.Spec.Values = &apiextensionsv1.JSON{Raw: jsonRaw}
	}

	// Handle values from ConfigMap/Secret
	if len(helmReleaseArgs.valuesFrom) > 0 {
		valuesFrom := []helmv2.ValuesReference{}
		for _, vFrom := range helmReleaseArgs.valuesFrom {
			parts := strings.Split(vFrom, "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid values-from format %q, expected Kind/name", vFrom)
			}

			kind := parts[0]
			name := parts[1]

			// Validate kind
			validKinds := []string{"ConfigMap", "Secret"}
			found := false
			for _, validKind := range validKinds {
				if strings.EqualFold(kind, validKind) {
					kind = validKind
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("invalid values-from kind %q, must be one of: %s", kind, strings.Join(validKinds, ", "))
			}

			valuesFrom = append(valuesFrom, helmv2.ValuesReference{
				Kind: kind,
				Name: name,
			})
		}
		helmRelease.Spec.ValuesFrom = valuesFrom
	}

	return helmRelease, nil
}

// mergeMaps merges two maps, with values from the second map taking precedence
func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
