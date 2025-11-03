package configmanager

import "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"

// FieldSelector defines a field and its metadata for configuration management.
type FieldSelector[T any] struct {
	Selector     func(*T) any // Function that returns a pointer to the field
	Description  string       // Human-readable description for CLI flags
	DefaultValue any          // Default value for the field
}

// AddFlagFromField returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
// Requires a default value as the second parameter, optionally accepts a description as the third parameter.
//
// Usage:
//
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, v1alpha1.DistributionKind)
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//		v1alpha1.DistributionKind, "Custom description")
func AddFlagFromField(
	selector func(*v1alpha1.Cluster) any,
	defaultValue any,
	description ...string,
) FieldSelector[v1alpha1.Cluster] {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return FieldSelector[v1alpha1.Cluster]{
		Selector:     selector,
		Description:  desc,
		DefaultValue: defaultValue,
	}
}

// DefaultDistributionFieldSelector creates a standard field selector for distribution.
func DefaultDistributionFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		Description:  "Kubernetes distribution to use",
		DefaultValue: v1alpha1.DistributionKind,
	}
}

// StandardSourceDirectoryFieldSelector creates a standard field selector for source directory.
func StandardSourceDirectoryFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
		Description:  "Directory containing workloads to deploy",
		DefaultValue: "k8s",
	}
}

// DefaultDistributionConfigFieldSelector creates a standard field selector for distribution config.
func DefaultDistributionConfigFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
		Description:  "Configuration file for the distribution",
		DefaultValue: "kind.yaml",
	}
}

// DefaultContextFieldSelector creates a standard field selector for kubernetes context.
// No default value is set as the context is distribution-specific and will be
// determined by the scaffolder based on the distribution type.
func DefaultContextFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:    func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
		Description: "Kubernetes context of cluster",
	}
}

// DefaultCNIFieldSelector creates a standard field selector for CNI.
func DefaultCNIFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
		Description:  "Container Network Interface (CNI) to use",
		DefaultValue: v1alpha1.CNIDefault,
	}
}

// DefaultGitOpsEngineFieldSelector creates a standard field selector for GitOps Engine.
func DefaultGitOpsEngineFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.GitOpsEngine },
		Description:  "GitOps engine to use",
		DefaultValue: v1alpha1.GitOpsEngineNone,
	}
}

// DefaultMetricsServerFieldSelector creates a standard field selector for Metrics Server.
func DefaultMetricsServerFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.MetricsServer },
		Description:  "Metrics Server configuration (Default: distribution decides, Enabled: install, Disabled: uninstall)",
		DefaultValue: v1alpha1.MetricsServerDefault,
	}
}

// DefaultKubeconfigFieldSelector creates a standard field selector for kubeconfig.
func DefaultKubeconfigFieldSelector() FieldSelector[v1alpha1.Cluster] {
	return FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
		Description:  "Path to kubeconfig file",
		DefaultValue: "~/.kube/config",
	}
}

// DefaultClusterFieldSelectors returns the default field selectors shared by cluster commands.
func DefaultClusterFieldSelectors() []FieldSelector[v1alpha1.Cluster] {
	return []FieldSelector[v1alpha1.Cluster]{
		DefaultDistributionFieldSelector(),
		DefaultDistributionConfigFieldSelector(),
		DefaultContextFieldSelector(),
		DefaultKubeconfigFieldSelector(),
	}
}
