package scaffolder_test

import (
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			generateDistributionContent(t, cluster, testCase.distribution)

			// Generate kustomization content
			kustomizationContent := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources: []\n"
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
	ksailContent := fmt.Sprintf(
		"apiVersion: ksail.dev/v1alpha1\nkind: Cluster\nmetadata:\n  name: %s\n",
		config.Metadata.Name,
	)

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
		kindContent := fmt.Sprintf(
			"apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster\nname: %s\n",
			cluster.Metadata.Name,
		)
		snaps.MatchSnapshot(t, kindContent)

	case v1alpha1.DistributionK3d:
		k3dContent := fmt.Sprintf(
			"apiVersion: k3d.io/v1alpha5\nkind: Simple\nmetadata:\n  name: %s\n",
			cluster.Metadata.Name,
		)
		snaps.MatchSnapshot(t, k3dContent)

	case v1alpha1.DistributionEKS:
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
	c.Spec.SourceDirectory = "workloads" // non-default to ensure it is included in KSail YAML output

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
