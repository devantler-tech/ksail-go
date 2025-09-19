package scaffolding_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/scaffolding"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewScaffolder(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-cluster")
	scaffolder := scaffolding.NewScaffolder(cluster)

	assert.NotNil(t, scaffolder)
	assert.Equal(t, cluster, scaffolder.KSailConfig)
	assert.NotNil(t, scaffolder.KSailYAMLGenerator)
	assert.NotNil(t, scaffolder.KindGenerator)
	assert.NotNil(t, scaffolder.K3dGenerator)
	assert.NotNil(t, scaffolder.EKSGenerator)
	assert.NotNil(t, scaffolder.KustomizationGenerator)
}

func TestScaffoldKind(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-kind-cluster")
	cluster.Spec.Distribution = v1alpha1.DistributionKind
	cluster.Spec.SourceDirectory = "k8s"

	scaffolder := scaffolding.NewScaffolder(cluster)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := tempDir + "/"

	err := scaffolder.Scaffold(outputPath, false)
	require.NoError(t, err)

	// Verify files were created
	assertFileExists(t, filepath.Join(tempDir, "ksail.yaml"))
	assertFileExists(t, filepath.Join(tempDir, "kind.yaml"))
	assertFileExists(t, filepath.Join(tempDir, "k8s"))

	// Verify content of ksail.yaml
	ksailContent, err := os.ReadFile(filepath.Join(tempDir, "ksail.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(ksailContent), "apiVersion: ksail.dev/v1alpha1")
	assert.Contains(t, string(ksailContent), "kind: Cluster")

	// Verify content of kind.yaml
	kindContent, err := os.ReadFile(filepath.Join(tempDir, "kind.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(kindContent), "apiVersion: kind.x-k8s.io/v1alpha4")
	assert.Contains(t, string(kindContent), "kind: Cluster")
}

func TestScaffoldK3d(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-k3d-cluster")
	cluster.Spec.Distribution = v1alpha1.DistributionK3d
	cluster.Spec.SourceDirectory = "manifests"

	scaffolder := scaffolding.NewScaffolder(cluster)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := tempDir + "/"

	err := scaffolder.Scaffold(outputPath, false)
	require.NoError(t, err)

	// Verify files were created
	assertFileExists(t, filepath.Join(tempDir, "ksail.yaml"))
	assertFileExists(t, filepath.Join(tempDir, "k3d.yaml"))
	assertFileExists(t, filepath.Join(tempDir, "manifests"))

	// Verify content of k3d.yaml
	k3dContent, err := os.ReadFile(filepath.Join(tempDir, "k3d.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(k3dContent), "apiVersion: k3d.io/v1alpha5")
	assert.Contains(t, string(k3dContent), "kind: Simple")
}

func TestScaffoldEKS(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-eks-cluster")
	cluster.Spec.Distribution = v1alpha1.DistributionEKS
	cluster.Spec.SourceDirectory = "manifests"

	scaffolder := scaffolding.NewScaffolder(cluster)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := tempDir + "/"

	err := scaffolder.Scaffold(outputPath, false)
	require.NoError(t, err)

	// Verify files were created
	assertFileExists(t, filepath.Join(tempDir, "ksail.yaml"))
	assertFileExists(t, filepath.Join(tempDir, "eks-config.yaml"))
	assertFileExists(t, filepath.Join(tempDir, "manifests"))

	// Verify content of eks-config.yaml
	eksContent, err := os.ReadFile(filepath.Join(tempDir, "eks-config.yaml"))
	require.NoError(t, err)
	assert.Contains(t, string(eksContent), "apiVersion: eksctl.io/v1alpha5")
	assert.Contains(t, string(eksContent), "kind: ClusterConfig")
}

func TestScaffoldTindError(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-tind-cluster")
	cluster.Spec.Distribution = v1alpha1.DistributionTind

	scaffolder := scaffolding.NewScaffolder(cluster)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := tempDir + "/"

	err := scaffolder.Scaffold(outputPath, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "talos-in-docker distribution is not yet implemented")
}

func TestScaffoldUnknownDistributionError(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-unknown-cluster")
	cluster.Spec.Distribution = "Unknown"

	scaffolder := scaffolding.NewScaffolder(cluster)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := tempDir + "/"

	err := scaffolder.Scaffold(outputPath, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "provided distribution is unknown")
}

func TestScaffoldWithForce(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("test-force-cluster")
	cluster.Spec.Distribution = v1alpha1.DistributionKind
	cluster.Spec.SourceDirectory = "k8s"

	scaffolder := scaffolding.NewScaffolder(cluster)

	// Create temp directory for output
	tempDir := t.TempDir()
	outputPath := tempDir + "/"

	// Create existing files
	ksailPath := filepath.Join(tempDir, "ksail.yaml")
	require.NoError(t, os.WriteFile(ksailPath, []byte("existing content"), 0o644))

	err := scaffolder.Scaffold(outputPath, true) // force = true
	require.NoError(t, err)

	// Verify files were overwritten
	assertFileExists(t, ksailPath)
	content, err := os.ReadFile(ksailPath)
	require.NoError(t, err)
	assert.NotEqual(t, "existing content", string(content))
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

// assertFileExists checks if a file exists and fails the test if it doesn't.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "file should exist: %s", path)
}