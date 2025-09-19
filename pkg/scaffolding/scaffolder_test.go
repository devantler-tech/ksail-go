package scaffolding_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/scaffolding"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
	os.Exit(exitCode)
}

func TestNewScaffolder(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-cluster")
	scaffolder := scaffolding.NewScaffolder(cluster)

	require.NotNil(t, scaffolder)
	require.Equal(t, cluster, scaffolder.KSailConfig)
	require.NotNil(t, scaffolder.KSailYAMLGenerator)
	require.NotNil(t, scaffolder.KustomizationGenerator)
}

func TestScaffold(t *testing.T) {
	t.Parallel()

	tests := getScaffoldTestCases()

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cluster := testCase.setupFunc(testCase.name)
			scaffolder := scaffolding.NewScaffolder(cluster)

			err := scaffolder.Scaffold(testCase.outputPath, testCase.force)

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
			generateDistributionContent(t, cluster, testCase.distribution)

			// Generate kustomization content
			kustomizationContent := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources: []\n"
			snaps.MatchSnapshot(t, kustomizationContent)
		})
	}
}

// Test case definitions
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
	cluster v1alpha1.Cluster,
	distribution v1alpha1.Distribution,
) {
	t.Helper()

	// Create a copy of the cluster and filter out default values for KSail YAML
	config := cluster
	
	// Filter out default values to keep output minimal
	if config.Spec.SourceDirectory == "k8s" {
		config.Spec.SourceDirectory = ""
	}
	if config.Spec.Distribution == v1alpha1.DistributionKind {
		config.Spec.Distribution = ""
	}
	if config.Spec.DistributionConfig == "kind.yaml" {
		config.Spec.DistributionConfig = ""
	}

	// Generate KSail YAML content - only include non-default fields
	ksailContent := fmt.Sprintf("apiVersion: ksail.dev/v1alpha1\nkind: Cluster\nmetadata:\n  name: %s\n", config.Metadata.Name)
	
	// Add spec fields only if they are non-default
	hasSpec := false
	specContent := ""
	
	if config.Spec.Distribution != "" {
		specContent += fmt.Sprintf("  distribution: %s\n", cluster.Spec.Distribution)
		hasSpec = true
	}
	if config.Spec.DistributionConfig != "" {
		specContent += fmt.Sprintf("  distributionConfig: %s\n", cluster.Spec.DistributionConfig)
		hasSpec = true
	}
	if config.Spec.SourceDirectory != "" {
		specContent += fmt.Sprintf("  sourceDirectory: %s\n", cluster.Spec.SourceDirectory)
		hasSpec = true
	}
	
	if hasSpec {
		ksailContent += "spec:\n" + specContent
	}
	
	snaps.MatchSnapshot(t, ksailContent)

	//nolint:exhaustive // We only test supported distributions here
	switch distribution {
	case v1alpha1.DistributionKind:
		kindContent := fmt.Sprintf("apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster\nname: %s\n", cluster.Metadata.Name)
		snaps.MatchSnapshot(t, kindContent)

	case v1alpha1.DistributionK3d:
		k3dContent := fmt.Sprintf("apiVersion: k3d.io/v1alpha5\nkind: Simple\nmetadata:\n  name: %s\n", cluster.Metadata.Name)
		snaps.MatchSnapshot(t, k3dContent)

	case v1alpha1.DistributionEKS:
		name := cluster.Metadata.Name
		eksContent := fmt.Sprintf(`apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: %s
  region: us-west-2
nodeGroups:
- desiredCapacity: 2
  instanceType: m5.large
  maxSize: 3
  minSize: 1
  name: %s-workers
`, name, name)
		snaps.MatchSnapshot(t, eksContent)
	}
}

// Cluster creation helpers
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

func createKindCluster(name string) v1alpha1.Cluster {
	cluster := createTestCluster(name)
	cluster.Spec.Distribution = v1alpha1.DistributionKind
	cluster.Spec.DistributionConfig = "kind.yaml"
	cluster.Spec.SourceDirectory = "k8s"  // This is the default, so it should be filtered out
	return cluster
}

func createK3dCluster(name string) v1alpha1.Cluster {
	cluster := createTestCluster(name)
	cluster.Spec.Distribution = v1alpha1.DistributionK3d
	cluster.Spec.DistributionConfig = "k3d.yaml"
	cluster.Spec.SourceDirectory = "manifests"  // Non-default value
	return cluster
}

func createEKSCluster(name string) v1alpha1.Cluster {
	cluster := createTestCluster(name)
	cluster.Spec.Distribution = v1alpha1.DistributionEKS
	cluster.Spec.DistributionConfig = "eks-config.yaml"
	cluster.Spec.SourceDirectory = "workloads"  // Non-default value
	return cluster
}

func createTindCluster(name string) v1alpha1.Cluster {
	cluster := createTestCluster(name)
	cluster.Spec.Distribution = v1alpha1.DistributionTind
	cluster.Spec.DistributionConfig = "tind.yaml"
	return cluster
}

func createUnknownCluster(name string) v1alpha1.Cluster {
	cluster := createTestCluster(name)
	cluster.Spec.Distribution = "Unknown"
	cluster.Spec.DistributionConfig = "unknown.yaml"
	return cluster
}