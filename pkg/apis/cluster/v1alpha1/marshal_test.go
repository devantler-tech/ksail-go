package v1alpha1_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestClusterMarshalDefaults verifies that default values are pruned from marshaled output.
func TestClusterMarshalDefaults(t *testing.T) {
	t.Parallel()

	t.Run("all_defaults_produces_minimal_output", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Distribution = v1alpha1.DefaultDistribution
		cluster.Spec.DistributionConfig = v1alpha1.DefaultDistributionConfig
		cluster.Spec.SourceDirectory = v1alpha1.DefaultSourceDirectory
		cluster.Spec.Connection.Kubeconfig = v1alpha1.DefaultKubeconfigPath
		cluster.Spec.CNI = v1alpha1.DefaultCNI
		cluster.Spec.CSI = v1alpha1.DefaultCSI
		cluster.Spec.MetricsServer = v1alpha1.DefaultMetricsServer
		cluster.Spec.LocalRegistry = v1alpha1.DefaultLocalRegistry
		cluster.Spec.GitOpsEngine = v1alpha1.DefaultGitOpsEngine

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Should only contain apiVersion and kind
		require.Contains(t, output, "apiVersion: ksail.dev/v1alpha1")
		require.Contains(t, output, "kind: Cluster")

		// Should NOT contain any spec fields when all are defaults
		require.NotContains(t, output, "spec:")
		require.NotContains(t, output, "distribution:")
		require.NotContains(t, output, "distributionConfig:")
		require.NotContains(t, output, "sourceDirectory:")
		require.NotContains(t, output, "cni:")
		require.NotContains(t, output, "gitOpsEngine:")
		require.NotContains(t, output, "localRegistry:")
	})

	t.Run("non_default_distribution_preserved", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Distribution = v1alpha1.DistributionK3d

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Should contain the non-default distribution
		require.Contains(t, output, "distribution: K3d")
	})

	t.Run("k3d_distribution_config_preserved", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Distribution = v1alpha1.DistributionK3d
		cluster.Spec.DistributionConfig = "k3d.yaml"

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// K3d with k3d.yaml config should preserve both
		// This is the key fix: derived default for non-default distribution
		require.Contains(t, output, "distribution: K3d")
		require.Contains(t, output, "distributionConfig: k3d.yaml")
	})

	t.Run("kind_distribution_config_pruned", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Distribution = v1alpha1.DistributionKind
		cluster.Spec.DistributionConfig = "kind.yaml"

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Kind with kind.yaml should be fully pruned (both are base defaults)
		require.NotContains(t, output, "distribution:")
		require.NotContains(t, output, "distributionConfig:")
	})

	t.Run("custom_distribution_config_preserved", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.DistributionConfig = "custom.yaml"

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Custom config should be preserved
		require.Contains(t, output, "distributionConfig: custom.yaml")
	})

	t.Run("non_default_gitops_with_auto_registry", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.GitOpsEngine = v1alpha1.GitOpsEngineFlux
		cluster.Spec.LocalRegistry = v1alpha1.LocalRegistryEnabled

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Non-default GitOps and LocalRegistry should be preserved
		require.Contains(t, output, "gitOpsEngine: Flux")
		require.Contains(t, output, "localRegistry: Enabled")
	})

	t.Run("k3d_with_flux_full_scenario", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Distribution = v1alpha1.DistributionK3d
		cluster.Spec.DistributionConfig = "k3d.yaml"
		cluster.Spec.GitOpsEngine = v1alpha1.GitOpsEngineFlux
		cluster.Spec.LocalRegistry = v1alpha1.LocalRegistryEnabled

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// All non-default values should be preserved
		require.Contains(t, output, "distribution: K3d")
		require.Contains(t, output, "distributionConfig: k3d.yaml")
		require.Contains(t, output, "gitOpsEngine: Flux")
		require.Contains(t, output, "localRegistry: Enabled")
	})

	t.Run("context_pruned_for_distribution", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Distribution = v1alpha1.DistributionK3d
		cluster.Spec.Connection.Context = v1alpha1.ExpectedContextName(v1alpha1.DistributionK3d)

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Expected context for the distribution should be pruned
		require.NotContains(t, output, "context:")
		require.NotContains(t, output, "k3d-k3d-default")
	})

	t.Run("custom_context_preserved", func(t *testing.T) {
		t.Parallel()

		cluster := v1alpha1.NewCluster()
		cluster.Spec.Connection.Context = "my-custom-context"

		data, err := yaml.Marshal(cluster)
		require.NoError(t, err)

		output := string(data)

		// Custom context should be preserved
		require.Contains(t, output, "context: my-custom-context")
	})
}

// TestClusterRoundTrip verifies that marshaling and unmarshaling preserves non-default values.
func TestClusterRoundTrip(t *testing.T) {
	t.Parallel()

	original := v1alpha1.NewCluster()
	original.Spec.Distribution = v1alpha1.DistributionK3d
	original.Spec.DistributionConfig = "k3d.yaml"
	original.Spec.GitOpsEngine = v1alpha1.GitOpsEngineFlux
	original.Spec.LocalRegistry = v1alpha1.LocalRegistryEnabled
	original.Spec.SourceDirectory = "manifests"

	// Marshal
	data, err := yaml.Marshal(original)
	require.NoError(t, err)

	// Unmarshal
	var restored v1alpha1.Cluster
	err = yaml.Unmarshal(data, &restored)
	require.NoError(t, err)

	// Verify marshaled output contains expected fields
	output := string(data)
	require.Contains(t, output, "distribution: K3d")
	require.Contains(t, output, "distributionConfig: k3d.yaml")
	require.Contains(t, output, "gitOpsEngine: Flux")
	require.Contains(t, output, "localRegistry: Enabled")
	require.Contains(t, output, "sourceDirectory: manifests")
}
