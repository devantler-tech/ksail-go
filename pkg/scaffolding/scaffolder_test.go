package scaffolding_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	kindconfig "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/scaffolding"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestMain(m *testing.M) {
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
}

func TestNewScaffolder(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-cluster")
	scaffolder := scaffolding.NewScaffolder(cluster)

	require.NotNil(t, scaffolder)
	require.Equal(t, cluster, scaffolder.KSailConfig)
	require.NotNil(t, scaffolder.KSailYAMLGenerator)
	require.NotNil(t, scaffolder.KindGenerator)
	require.NotNil(t, scaffolder.K3dGenerator)
	require.NotNil(t, scaffolder.EKSGenerator)
	require.NotNil(t, scaffolder.KustomizationGenerator)
}

func TestScaffold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		distribution v1alpha1.Distribution
		setupCluster func(cluster *v1alpha1.Cluster)
		expectError  bool
	}{
		{
			name:         "Kind distribution",
			distribution: v1alpha1.DistributionKind,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "k8s"
			},
		},
		{
			name:         "K3d distribution",
			distribution: v1alpha1.DistributionK3d,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "manifests"
			},
		},
		{
			name:         "EKS distribution",
			distribution: v1alpha1.DistributionEKS,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "workloads"
			},
		},
		{
			name:         "Tind distribution not implemented",
			distribution: v1alpha1.DistributionTind,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "k8s"
			},
			expectError: true,
		},
		{
			name:         "Unknown distribution",
			distribution: "Unknown",
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "k8s"
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cluster := createTestCluster("test-cluster")
			cluster.Spec.Distribution = tt.distribution
			tt.setupCluster(&cluster)

			scaffolder := scaffolding.NewScaffolder(cluster)

			// Test scaffolding without output path (content generation only)
			err := scaffolder.Scaffold("", false)

			if tt.expectError {
				require.Error(t, err)
				snaps.MatchSnapshot(t, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGeneratedContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		distribution v1alpha1.Distribution
		setupCluster func(cluster *v1alpha1.Cluster)
	}{
		{
			name:         "Kind configuration content",
			distribution: v1alpha1.DistributionKind,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "k8s"
			},
		},
		{
			name:         "K3d configuration content",
			distribution: v1alpha1.DistributionK3d,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "manifests"
			},
		},
		{
			name:         "EKS configuration content",
			distribution: v1alpha1.DistributionEKS,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = "workloads"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cluster := createTestCluster("test-cluster")
			cluster.Spec.Distribution = tt.distribution
			tt.setupCluster(&cluster)

			scaffolder := scaffolding.NewScaffolder(cluster)

			// Test KSail YAML generation
			ksailContent, err := scaffolder.KSailYAMLGenerator.Generate(
				cluster,
				yamlgenerator.Options{},
			)
			require.NoError(t, err)
			snaps.MatchSnapshot(t, ksailContent)

			// Test distribution-specific content
			switch tt.distribution {
			case v1alpha1.DistributionKind:
				kindConfig := createDefaultKindConfig(cluster.Metadata.Name)
				kindContent, err := scaffolder.KindGenerator.Generate(
					kindConfig,
					yamlgenerator.Options{},
				)
				require.NoError(t, err)
				snaps.MatchSnapshot(t, kindContent)

			case v1alpha1.DistributionK3d:
				k3dContent, err := scaffolder.K3dGenerator.Generate(
					&cluster,
					yamlgenerator.Options{},
				)
				require.NoError(t, err)
				snaps.MatchSnapshot(t, k3dContent)

			case v1alpha1.DistributionEKS:
				eksConfig := createDefaultEKSConfig(cluster.Metadata.Name)
				eksContent, err := scaffolder.EKSGenerator.Generate(
					eksConfig,
					yamlgenerator.Options{},
				)
				require.NoError(t, err)
				snaps.MatchSnapshot(t, eksContent)
			}

			// Test Kustomization generation
			kustomizationContent, err := scaffolder.KustomizationGenerator.Generate(
				&cluster,
				yamlgenerator.Options{},
			)
			require.NoError(t, err)
			snaps.MatchSnapshot(t, kustomizationContent)
		})
	}
}

// createTestCluster creates a test cluster configuration.
func createTestCluster(name string) v1alpha1.Cluster {
	return v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Metadata: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionKind,
			SourceDirectory:    "k8s",
			DistributionConfig: "kind.yaml",
		},
	}
}

// createDefaultKindConfig creates a default Kind cluster configuration.
func createDefaultKindConfig(name string) *v1alpha4.Cluster {
	kindCluster := kindconfig.NewKindCluster(name, "", "")
	// Add a minimal control plane node
	var node v1alpha4.Node
	node.Role = v1alpha4.ControlPlaneRole
	kindCluster.Nodes = append(kindCluster.Nodes, node)
	return kindCluster
}

// createDefaultEKSConfig creates a minimal EKS cluster configuration for testing.
func createDefaultEKSConfig(name string) *v1alpha5.ClusterConfig {
	minSize := 1
	maxSize := 3
	desiredCapacity := 2

	return &v1alpha5.ClusterConfig{
		TypeMeta: v1alpha5.ClusterConfigTypeMeta(),
		Metadata: &v1alpha5.ClusterMeta{
			Name:    name,
			Region:  "us-west-2",
			Version: "",
		},
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name:         name + "-workers",
					InstanceType: "m5.large",
					ScalingConfig: &v1alpha5.ScalingConfig{
						MinSize:         &minSize,
						MaxSize:         &maxSize,
						DesiredCapacity: &desiredCapacity,
					},
				},
			},
		},
	}
}
