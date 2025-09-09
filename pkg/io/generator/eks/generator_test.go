package eksgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

func TestEKSGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfig("test-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cfg, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "test-cluster")
}

func TestEKSGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfig("file-cluster")
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "eks-config.yaml")
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cfg, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "file-cluster")

	// Verify file was written
	testutils.AssertFileEquals(t, tempDir, outputPath, result)
}

func TestEKSGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfig("existing-no-force")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cfg,
		"eks-config.yaml",
		assertEKSYAML,
		"existing-no-force",
		false,
	)
}

func TestEKSGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfig("existing-with-force")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cfg,
		"eks-config.yaml",
		assertEKSYAML,
		"existing-with-force",
		true,
	)
}

func TestEKSGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfig("error-cluster")

	// Use an invalid file path that will cause a write error
	invalidPath := "/dev/null/invalid/path/eks-config.yaml"
	opts := yamlgenerator.Options{
		Output: invalidPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cfg, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), "write EKS config", "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}

func TestEKSGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Act & Assert
	testEKSMarshalError(
		t,
		createTestClusterConfig,
		"marshal EKS config",
	)
}

func TestEKSGenerator_Generate_WithCustomOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfigWithOptions("custom-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cfg, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "custom-cluster")

	// Verify custom options are applied
	assert.Contains(t, result, "us-east-1", "YAML should contain custom region")
	assert.Contains(t, result, "t3.medium", "YAML should contain custom instance type")
	assert.Contains(t, result, "\"1.25\"", "YAML should contain custom Kubernetes version")
}

func TestEKSGenerator_Generate_DefaultValues(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cfg := createTestClusterConfig("default-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cfg, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "default-cluster")
}

// createTestClusterConfigBase creates a test EKS cluster configuration with customizable parameters.
func createTestClusterConfigBase(
	name, region, version, instanceType string,
	minNodes, maxNodes, desiredNodes int,
) *v1alpha5.ClusterConfig {
	return &v1alpha5.ClusterConfig{
		TypeMeta:                v1alpha5.ClusterConfigTypeMeta(),
		Metadata:                createTestMetadata(name, region, version),
		KubernetesNetworkConfig: nil,
		AutoModeConfig:          nil,
		RemoteNetworkConfig:     nil,
		IAM:                     nil,
		IAMIdentityMappings:     nil,
		IdentityProviders:       nil,
		AccessConfig:            nil,
		VPC:                     nil,
		Addons:                  nil,
		AddonsConfig:            createTestAddonsConfig(),
		PrivateCluster:          nil,
		NodeGroups: createTestNodeGroups(
			name,
			instanceType,
			minNodes,
			maxNodes,
			desiredNodes,
		),
		ManagedNodeGroups: nil,
		FargateProfiles:   nil,
		AvailabilityZones: nil,
		LocalZones:        nil,
		CloudWatch:        nil,
		SecretsEncryption: nil,
		Status:            nil,
		GitOps:            nil,
		Karpenter:         nil,
		Outpost:           nil,
		ZonalShiftConfig:  nil,
	}
}

func createTestMetadata(name, region, version string) *v1alpha5.ClusterMeta {
	return &v1alpha5.ClusterMeta{
		Name:               name,
		Region:             region,
		Version:            version,
		ForceUpdateVersion: nil,
		Tags:               nil,
		Annotations:        nil,
		AccountID:          "",
	}
}

func createTestAddonsConfig() v1alpha5.AddonsConfig {
	return v1alpha5.AddonsConfig{
		AutoApplyPodIdentityAssociations: false,
		DisableDefaultAddons:             false,
	}
}

func createTestNodeGroups(
	name, instanceType string,
	minNodes, maxNodes, desiredNodes int,
) []*v1alpha5.NodeGroup {
	return []*v1alpha5.NodeGroup{
		{
			NodeGroupBase: createTestNodeGroupBase(
				name,
				instanceType,
				minNodes,
				maxNodes,
				desiredNodes,
			),
		},
	}
}

func createTestNodeGroupBase(
	name, instanceType string,
	minNodes, maxNodes, desiredNodes int,
) *v1alpha5.NodeGroupBase {
	return clustertestutils.CreateTestEKSNodeGroupBase(clustertestutils.EKSNodeGroupBaseOptions{
		Name:            name + "-workers",
		InstanceType:    instanceType,
		MinSize:         &minNodes,
		MaxSize:         &maxNodes,
		DesiredCapacity: &desiredNodes,
	})
}

// createTestClusterConfig creates a minimal test EKS cluster configuration.
func createTestClusterConfig(name string) *v1alpha5.ClusterConfig {
	return createTestClusterConfigBase(name, "us-west-2", "", "m5.large", 1, 3, 2)
}

// createTestClusterConfigWithOptions creates a test cluster config with custom EKS options.
func createTestClusterConfigWithOptions(name string) *v1alpha5.ClusterConfig {
	return createTestClusterConfigBase(name, "us-east-1", "1.25", "t3.medium", 2, 5, 3)
}

// assertEKSYAML ensures the generated YAML contains the expected boilerplate and cluster name.
func assertEKSYAML(t *testing.T, result string, clusterName string) {
	t.Helper()
	assert.Contains(t, result, "apiVersion: eksctl.io/v1alpha5", "YAML should contain API version")
	assert.Contains(t, result, "kind: ClusterConfig", "YAML should contain kind")
	assert.Contains(t, result, "name: "+clusterName, "YAML should contain cluster name")
	assert.Contains(t, result, "nodeGroups:", "YAML should contain node groups")
	assert.Contains(t, result, clusterName+"-workers", "YAML should contain node group name")
}

// testEKSMarshalError runs a test pattern for EKS generator marshal errors.
func testEKSMarshalError(
	t *testing.T,
	createClusterConfig func(string) *v1alpha5.ClusterConfig,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange
	gen := generator.NewEKSGenerator()
	gen.Marshaller = generatortestutils.MarshalFailer[*v1alpha5.ClusterConfig]{
		Marshaller: nil,
	}
	cfg := createClusterConfig("marshal-error-cluster")

	// Act & Assert
	generatortestutils.TestGeneratorMarshalError[*v1alpha5.ClusterConfig, *v1alpha5.ClusterConfig](
		t,
		gen,
		cfg,
		expectedErrorContains,
	)
}
