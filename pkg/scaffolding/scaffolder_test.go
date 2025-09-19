package scaffolding_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/scaffolding"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultSourceDirectory = "k8s"
)

func TestMain(m *testing.M) {
	// Clean snapshots after tests - ignore exit code
	_, _ = snaps.Clean(m, snaps.CleanOpts{Sort: true})
}

func TestNewScaffolder(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-cluster")
	scaffolder, err := scaffolding.NewScaffolder(cluster)

	require.NoError(t, err)
	require.NotNil(t, scaffolder)
	require.Equal(t, cluster, scaffolder.KSailConfig)
	require.NotNil(t, scaffolder.KSailYAMLGenerator)
	require.NotNil(t, scaffolder.DistributionGenerator)
	require.NotNil(t, scaffolder.KustomizationGenerator)
}

func TestScaffold(t *testing.T) {
	t.Parallel()

	tests := getScaffoldTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := createTestCluster("test-cluster")
			cluster.Spec.Distribution = testCase.distribution
			testCase.setupCluster(&cluster)

			scaffolder, err := scaffolding.NewScaffolder(cluster)

			if testCase.expectError {
				require.Error(t, err)
				snaps.MatchSnapshot(t, err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, scaffolder)

				// Test scaffolding without output path (content generation only)
				err = scaffolder.Scaffold("", false)
				require.NoError(t, err)
			}
		})
	}
}

// getScaffoldTestCases returns the test cases for scaffolding tests.
func getScaffoldTestCases() []struct {
	name         string
	distribution v1alpha1.Distribution
	setupCluster func(cluster *v1alpha1.Cluster)
	expectError  bool
} {
	return []struct {
		name         string
		distribution v1alpha1.Distribution
		setupCluster func(cluster *v1alpha1.Cluster)
		expectError  bool
	}{
		{
			name:         "Kind distribution",
			distribution: v1alpha1.DistributionKind,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = defaultSourceDirectory
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
				cluster.Spec.SourceDirectory = defaultSourceDirectory
			},
			expectError: true,
		},
		{
			name:         "Unknown distribution",
			distribution: "Unknown",
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = defaultSourceDirectory
			},
			expectError: true,
		},
	}
}

// getGeneratedContentTestCases returns the test cases for content generation tests.
func getGeneratedContentTestCases() []struct {
	name         string
	distribution v1alpha1.Distribution
	setupCluster func(cluster *v1alpha1.Cluster)
} {
	return []struct {
		name         string
		distribution v1alpha1.Distribution
		setupCluster func(cluster *v1alpha1.Cluster)
	}{
		{
			name:         "Kind configuration content",
			distribution: v1alpha1.DistributionKind,
			setupCluster: func(cluster *v1alpha1.Cluster) {
				cluster.Spec.SourceDirectory = defaultSourceDirectory
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
}

func TestGeneratedContent(t *testing.T) {
	t.Parallel()

	tests := getGeneratedContentTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := createTestCluster("test-cluster")
			cluster.Spec.Distribution = testCase.distribution
			testCase.setupCluster(&cluster)

			scaffolder, err := scaffolding.NewScaffolder(cluster)
			require.NoError(t, err)

			testAllContentGeneration(t, scaffolder, cluster, testCase.distribution)
		})
	}
}

// testAllContentGeneration tests all aspects of content generation for a given distribution.
func testAllContentGeneration(
	t *testing.T,
	scaffolder *scaffolding.Scaffolder,
	cluster v1alpha1.Cluster,
	distribution v1alpha1.Distribution,
) {
	t.Helper()

	// Test KSail YAML generation
	ksailContent, err := scaffolder.KSailYAMLGenerator.Generate(
		cluster,
		yamlgenerator.Options{},
	)
	require.NoError(t, err)
	snaps.MatchSnapshot(t, ksailContent)

	// Test distribution-specific content
	testDistributionSpecificContent(t, scaffolder, cluster, distribution)

	// Test Kustomization generation
	kustomizationContent, err := scaffolder.KustomizationGenerator.Generate(
		&cluster,
		yamlgenerator.Options{},
	)
	require.NoError(t, err)
	snaps.MatchSnapshot(t, kustomizationContent)
}

// testDistributionSpecificContent tests the generation of distribution-specific configurations.
func testDistributionSpecificContent(
	t *testing.T,
	scaffolder *scaffolding.Scaffolder,
	cluster v1alpha1.Cluster,
	_ v1alpha1.Distribution,
) {
	t.Helper()

	// Test distribution-specific content using the unified DistributionGenerator
	distContent, err := scaffolder.DistributionGenerator.Generate(&cluster, yamlgenerator.Options{})
	require.NoError(t, err)
	snaps.MatchSnapshot(t, distContent)
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
			SourceDirectory:    defaultSourceDirectory,
			DistributionConfig: "kind.yaml",
		},
	}
}
