package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/spf13/cobra"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

const (
	defaultInterval = 1 * time.Minute
)

// resourceReference represents a parsed Kubernetes resource reference.
type resourceReference struct {
	Kind      string
	Name      string
	Namespace string
}

// parseResourceReference parses a string in format "Kind/name" or "Kind/name.namespace".
func parseResourceReference(ref, defaultNamespace, errorContext string) (*resourceReference, error) {
	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid %s format, expected Kind/name or Kind/name.namespace", errorContext)
	}

	rr := &resourceReference{
		Kind:      parts[0],
		Name:      parts[1],
		Namespace: defaultNamespace,
	}

	// Check if namespace is included in the name
	if strings.Contains(rr.Name, ".") {
		nameParts := strings.SplitN(rr.Name, ".", 2)
		rr.Name = nameParts[0]
		rr.Namespace = nameParts[1]
	}

	return rr, nil
}

// validateKind checks if a kind is in the list of valid kinds.
func validateKind(kind string, validKinds []string, errorContext string) error {
	for _, validKind := range validKinds {
		if kind == validKind {
			return nil
		}
	}
	return fmt.Errorf("invalid %s kind %q, must be one of: %s", errorContext, kind, strings.Join(validKinds, ", "))
}

// validateKindCaseInsensitive checks if a kind matches (case-insensitive) one of the valid kinds and returns the canonical form.
func validateKindCaseInsensitive(kind string, validKinds []string, errorContext string) (string, error) {
	for _, validKind := range validKinds {
		if strings.EqualFold(kind, validKind) {
			return validKind, nil
		}
	}
	return "", fmt.Errorf("invalid %s kind %q, must be one of: %s", errorContext, kind, strings.Join(validKinds, ", "))
}

// parseDependency parses a depends-on reference in format "name" or "namespace/name".
func parseDependency(dep string) (*helmv2.DependencyReference, error) {
	parts := strings.Split(dep, "/")
	if len(parts) == 1 {
		// Same namespace
		return &helmv2.DependencyReference{
			Name: parts[0],
		}, nil
	} else if len(parts) == 2 {
		// Different namespace
		return &helmv2.DependencyReference{
			Namespace: parts[0],
			Name:      parts[1],
		}, nil
	}
	return nil, fmt.Errorf("invalid depends-on format %q, expected name or namespace/name", dep)
}


// NewHelmReleaseCmd creates the workload gen helmrelease command.
func NewHelmReleaseCmd(_ *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "helmrelease [NAME]",
		Aliases: []string{"hr"},
		Short:   "Generate a HelmRelease resource",
		Long:    "Generate a HelmRelease resource for a given HelmRepository, GitRepository, Bucket, or chart reference source.",
		Example: `  # Generate a HelmRelease with a chart from a HelmRepository source
  ksail workload gen helmrelease podinfo \
    --interval=10m \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --chart-version=">4.0.0" \
    --export

  # Generate a HelmRelease with a chart from a GitRepository source
  ksail workload gen helmrelease podinfo \
    --interval=10m \
    --source=GitRepository/podinfo \
    --chart=./charts/podinfo \
    --export

  # Generate a HelmRelease with values from local YAML files
  ksail workload gen helmrelease podinfo \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --values=./my-values1.yaml \
    --values=./my-values2.yaml \
    --export

  # Generate a HelmRelease with values from a Kubernetes secret
  ksail workload gen helmrelease podinfo \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --values-from=Secret/my-secret-values \
    --export

  # Generate a HelmRelease with a custom release name
  ksail workload gen helmrelease podinfo \
    --release-name=podinfo-dev \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --export

  # Generate a HelmRelease targeting another namespace
  ksail workload gen helmrelease podinfo \
    --target-namespace=test \
    --create-target-namespace=true \
    --source=HelmRepository/podinfo \
    --chart=podinfo \
    --export

  # Generate a HelmRelease using a chart reference
  ksail workload gen helmrelease podinfo \
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

			helmRelease, err := buildHelmRelease(
				name,
				namespace,
				source,
				chart,
				chartVersion,
				chartRef,
				targetNamespace,
				storageNamespace,
				createNamespace,
				dependsOn,
				interval,
				timeout,
				values,
				valuesFrom,
				saName,
				crdsPolicy,
				kubeConfigSecretRef,
				releaseName,
			)
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
				_, err = fmt.Fprint(cmd.OutOrStdout(), yaml)
				if err != nil {
					return fmt.Errorf("failed to write YAML: %w", err)
				}
			} else {
				return fmt.Errorf("applying HelmRelease to cluster is not yet implemented, use --export flag to generate YAML")
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
	flags.String(
		"source",
		"",
		"source that contains the chart (HelmRepository/name, GitRepository/name, Bucket/name)",
	)
	flags.String("chart", "", "Helm chart name or path")
	flags.String("chart-ref", "", "Helm chart reference (HelmChart/name, OCIRepository/name)")

	// Optional flags
	flags.StringP("namespace", "n", "default", "namespace scope for the HelmRelease")
	flags.Duration("interval", defaultInterval, "reconciliation interval")
	flags.String("chart-version", "", "Helm chart version, accepts a semver range")
	flags.String("target-namespace", "", "namespace to target when performing operations")
	flags.String("storage-namespace", "", "namespace for Helm storage")
	flags.Bool("create-target-namespace", false, "create the target namespace if not present")
	flags.StringSlice("depends-on", nil, "HelmReleases that must be ready before this release")
	flags.Duration("timeout", 5*time.Minute, "timeout for any individual Kubernetes operation")
	flags.StringSlice("values", nil, "local values YAML files")
	flags.StringSlice("values-from", nil, "values from ConfigMap or Secret")
	flags.String("service-account", "", "service account name to impersonate")
	flags.String("crds", "", "CRDs policy (Create, CreateReplace, Skip)")
	flags.String(
		"kubeconfig-secret-ref",
		"",
		"KubeConfig secret reference for remote reconciliation",
	)
	flags.String("release-name", "", "name used for the Helm release")
	flags.Bool("export", false, "export in YAML format to stdout")

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
		if err := validateKind(crdsPolicy, validPolicies, "crds policy"); err != nil {
			return err
		}
	}

	return nil
}

func buildHelmRelease(name, namespace, source, chart, chartVersion, chartRef,
	targetNamespace, storageNamespace string, createNamespace bool, dependsOn []string,
	interval, timeout time.Duration, values, valuesFrom []string, saName, crdsPolicy,
	kubeConfigSecretRef, releaseName string,
) (*helmv2.HelmRelease, error) {
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
		sourceRef, err := parseResourceReference(source, namespace, "source")
		if err != nil {
			return nil, err
		}

		// Validate source kind
		validSourceKinds := []string{
			sourcev1.HelmRepositoryKind,
			sourcev1.GitRepositoryKind,
			sourcev1.BucketKind,
		}
		if err := validateKind(sourceRef.Kind, validSourceKinds, "source"); err != nil {
			return nil, err
		}

		helmRelease.Spec.Chart = &helmv2.HelmChartTemplate{
			Spec: helmv2.HelmChartTemplateSpec{
				Chart: chart,
				SourceRef: helmv2.CrossNamespaceObjectReference{
					Kind:      sourceRef.Kind,
					Name:      sourceRef.Name,
					Namespace: sourceRef.Namespace,
				},
			},
		}

		if chartVersion != "" {
			helmRelease.Spec.Chart.Spec.Version = chartVersion
		}
	} else if chartRef != "" {
		chartReference, err := parseResourceReference(chartRef, namespace, "chart-ref")
		if err != nil {
			return nil, err
		}

		// Validate chartRef kind
		validChartRefKinds := []string{sourcev1.OCIRepositoryKind, sourcev1.HelmChartKind}
		if err := validateKind(chartReference.Kind, validChartRefKinds, "chart-ref"); err != nil {
			return nil, err
		}

		helmRelease.Spec.ChartRef = &helmv2.CrossNamespaceSourceReference{
			Kind:      chartReference.Kind,
			Name:      chartReference.Name,
			Namespace: chartReference.Namespace,
		}
	}

	// Set optional fields
	if releaseName != "" {
		helmRelease.Spec.ReleaseName = releaseName
	}

	if targetNamespace != "" {
		helmRelease.Spec.TargetNamespace = targetNamespace
	}

	if storageNamespace != "" {
		helmRelease.Spec.StorageNamespace = storageNamespace
	}

	if createNamespace {
		if helmRelease.Spec.Install == nil {
			helmRelease.Spec.Install = &helmv2.Install{}
		}
		helmRelease.Spec.Install.CreateNamespace = true
	}

	if len(dependsOn) > 0 {
		deps := []helmv2.DependencyReference{}
		for _, dep := range dependsOn {
			depRef, err := parseDependency(dep)
			if err != nil {
				return nil, err
			}
			deps = append(deps, *depRef)
		}
		helmRelease.Spec.DependsOn = deps
	}

	if timeout > 0 {
		helmRelease.Spec.Timeout = &metav1.Duration{Duration: timeout}
	}

	if saName != "" {
		helmRelease.Spec.ServiceAccountName = saName
	}

	if crdsPolicy != "" {
		if helmRelease.Spec.Install == nil {
			helmRelease.Spec.Install = &helmv2.Install{}
		}
		helmRelease.Spec.Install.CRDs = helmv2.Create

		if helmRelease.Spec.Upgrade == nil {
			helmRelease.Spec.Upgrade = &helmv2.Upgrade{}
		}
		helmRelease.Spec.Upgrade.CRDs = helmv2.CRDsPolicy(crdsPolicy)
	}

	if kubeConfigSecretRef != "" {
		helmRelease.Spec.KubeConfig = &meta.KubeConfigReference{
			SecretRef: &meta.SecretKeyReference{
				Name: kubeConfigSecretRef,
			},
		}
	}

	// Handle values files
	if len(values) > 0 {
		valuesMap := make(map[string]interface{})
		for _, vFile := range values {
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
	if len(valuesFrom) > 0 {
		valsFrom := []helmv2.ValuesReference{}
		for _, vFrom := range valuesFrom {
			parts := strings.Split(vFrom, "/")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid values-from format %q, expected Kind/name", vFrom)
			}

			kind := parts[0]
			name := parts[1]

			// Validate kind (case-insensitive)
			validKinds := []string{"ConfigMap", "Secret"}
			canonicalKind, err := validateKindCaseInsensitive(kind, validKinds, "values-from")
			if err != nil {
				return nil, err
			}

			valsFrom = append(valsFrom, helmv2.ValuesReference{
				Kind: canonicalKind,
				Name: name,
			})
		}
		helmRelease.Spec.ValuesFrom = valsFrom
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
