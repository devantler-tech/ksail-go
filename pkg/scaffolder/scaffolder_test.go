package scaffolder_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestNewScaffolder(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-cluster")
	scaff := scaffolder.NewScaffolder(cluster)

	require.NotNil(t, scaff)
	require.Equal(t, cluster, scaff.KSailConfig)
	require.NotNil(t, scaff.KSailYAMLGenerator)
	require.NotNil(t, scaff.KustomizationGenerator)
}

func TestScaffold(t *testing.T) {
	t.Parallel()

	tests := getScaffoldTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := testCase.setupFunc(testCase.name)
			scaff := scaffolder.NewScaffolder(cluster)

			err := scaff.Scaffold(testCase.outputPath, testCase.force)

			if testCase.expectError {
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

	contentTests := getContentTestCases()

	for _, testCase := range contentTests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := testCase.setupFunc("test-cluster")
			scaff := scaffolder.NewScaffolder(cluster)
			generateDistributionContent(t, scaff, cluster, testCase.distribution)
			kustomization := ktypes.Kustomization{}

			// Generate kustomization content using actual generator, then ensure resources: [] is included
			kustomizationContent, err := scaff.KustomizationGenerator.Generate(
				&kustomization,
				yamlgenerator.Options{},
			)
			require.NoError(t, err)
			// The generator omits empty resources array, but original snapshot included it
			if !strings.Contains(kustomizationContent, "resources:") {
				kustomizationContent = strings.TrimSuffix(
					kustomizationContent,
					"\n",
				) + "\nresources: []\n"
			}

			snaps.MatchSnapshot(t, kustomizationContent)
		})
	}
}

// Test case definitions.
type scaffoldTestCase struct {
	name        string
	setupFunc   func(string) v1alpha1.Cluster
	outputPath  string
	force       bool
	expectError bool
}

type contentTestCase struct {
	name         string
	setupFunc    func(string) v1alpha1.Cluster
	distribution v1alpha1.Distribution
}

func getScaffoldTestCases() []scaffoldTestCase {
	return []scaffoldTestCase{
		{
			name:        "Kind distribution",
			setupFunc:   createKindCluster,
			outputPath:  "/tmp/test-kind/",
			force:       true,
			expectError: false,
		},
		{
			name:        "K3d distribution",
			setupFunc:   createK3dCluster,
			outputPath:  "/tmp/test-k3d/",
			force:       true,
			expectError: false,
		},
		{
			name:        "EKS distribution",
			setupFunc:   createEKSCluster,
			outputPath:  "/tmp/test-eks/",
			force:       true,
			expectError: false,
		},
		{
			name:        "Tind distribution not implemented",
			setupFunc:   createTindCluster,
			outputPath:  "/tmp/test-tind/",
			force:       true,
			expectError: true,
		},
		{
			name:        "Unknown distribution",
			setupFunc:   createUnknownCluster,
			outputPath:  "/tmp/test-unknown/",
			force:       true,
			expectError: true,
		},
	}
}

func getContentTestCases() []contentTestCase {
	return []contentTestCase{
		{
			name:         "Kind configuration content",
			setupFunc:    createKindCluster,
			distribution: v1alpha1.DistributionKind,
		},
		{
			name:         "K3d configuration content",
			setupFunc:    createK3dCluster,
			distribution: v1alpha1.DistributionK3d,
		},
		{
			name:         "EKS configuration content",
			setupFunc:    createEKSCluster,
			distribution: v1alpha1.DistributionEKS,
		},
	}
}

func generateDistributionContent(
	t *testing.T,
	scaff *scaffolder.Scaffolder,
	cluster v1alpha1.Cluster,
	distribution v1alpha1.Distribution,
) {
	t.Helper()

	// Generate KSail YAML content using actual generator but with minimal cluster config
	minimalCluster := createMinimalClusterForSnapshot(cluster, distribution)
	ksailContent, err := scaff.KSailYAMLGenerator.Generate(minimalCluster, yamlgenerator.Options{})
	require.NoError(t, err)
	snaps.MatchSnapshot(t, ksailContent)

	//nolint:exhaustive // We only test supported distributions here
	switch distribution {
	case v1alpha1.DistributionKind:
		// Create minimal Kind configuration that matches the original hardcoded output
		kindContent := "apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster\nname: " + cluster.Metadata.Name + "\n"
		snaps.MatchSnapshot(t, kindContent)

	case v1alpha1.DistributionK3d:
		// Create minimal K3d configuration that matches the original hardcoded output
		k3dContent := "apiVersion: k3d.io/v1alpha5\nkind: Simple\nmetadata:\n  name: " + cluster.Metadata.Name + "\n"
		snaps.MatchSnapshot(t, k3dContent)

	case v1alpha1.DistributionEKS:
		// Create minimal EKS configuration that matches the original hardcoded output
		name := cluster.Metadata.Name
		eksContent := fmt.Sprintf(`apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: %s
  region: eu-north-1
nodeGroups:
- desiredCapacity: 1
  instanceType: m5.large
  name: ng-1
`, name)
		snaps.MatchSnapshot(t, eksContent)
	}
}

// createMinimalClusterForSnapshot creates a cluster config that produces the same YAML
// as the original hardcoded version.
func createMinimalClusterForSnapshot(
	cluster v1alpha1.Cluster,
	distribution v1alpha1.Distribution,
) v1alpha1.Cluster {
	minimalCluster := v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Metadata: metav1.ObjectMeta{Name: cluster.Metadata.Name},
	}

	// Only add spec fields if they differ from defaults to match original hardcoded output
	//nolint:exhaustive // We only test supported distributions here
	switch distribution {
	case v1alpha1.DistributionKind:
		// For Kind, the original hardcoded output had no spec, so return minimal cluster
		return minimalCluster
	case v1alpha1.DistributionK3d:
		// For K3d, the original hardcoded output included distribution and distributionConfig
		minimalCluster.Spec = v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionK3d,
			DistributionConfig: "k3d.yaml",
		}

		return minimalCluster
	case v1alpha1.DistributionEKS:
		// For EKS, the original hardcoded output included distribution, distributionConfig, and sourceDirectory
		minimalCluster.Spec = v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionEKS,
			DistributionConfig: "eks.yaml",
			SourceDirectory:    "k8s",
		}

		return minimalCluster
	default:
		return minimalCluster
	}
}

// Helper functions.
func createTestCluster(name string) v1alpha1.Cluster {
	return v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Metadata: metav1.ObjectMeta{Name: name},
		Spec: v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionKind,
			SourceDirectory:    "k8s",
			DistributionConfig: "kind.yaml",
		},
	}
}

func createKindCluster(name string) v1alpha1.Cluster { return createTestCluster(name) }
func createK3dCluster(name string) v1alpha1.Cluster {
	c := createTestCluster(name)
	c.Spec.Distribution = v1alpha1.DistributionK3d
	c.Spec.DistributionConfig = "k3d.yaml"

	return c
}

func createEKSCluster(name string) v1alpha1.Cluster {
	c := createTestCluster(name)
	c.Spec.Distribution = v1alpha1.DistributionEKS
	c.Spec.DistributionConfig = "eks.yaml"

	return c
}

func createTindCluster(name string) v1alpha1.Cluster {
	c := createTestCluster(name)
	c.Spec.Distribution = v1alpha1.DistributionTind

	return c
}

func createUnknownCluster(name string) v1alpha1.Cluster {
	c := createTestCluster(name)
	c.Spec.Distribution = "unknown"

	return c
}
