package eksgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
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

// createTestClusterConfig creates a minimal test EKS cluster configuration.
func createTestClusterConfig(name string) *v1alpha5.ClusterConfig {
	minNodes := 1
	maxNodes := 3
	desiredNodes := 2

	return &v1alpha5.ClusterConfig{
		TypeMeta: v1alpha5.ClusterConfigTypeMeta(),
		Metadata: &v1alpha5.ClusterMeta{
			Name:                name,
			Region:              "us-west-2",
			Version:             "",
			ForceUpdateVersion:  nil,
			Tags:                nil,
			Annotations:         nil,
			AccountID:           "",
		},
		KubernetesNetworkConfig: nil,
		AutoModeConfig:          nil,
		RemoteNetworkConfig:     nil,
		IAM:                     nil,
		IAMIdentityMappings:     nil,
		IdentityProviders:       nil,
		AccessConfig:            nil,
		VPC:                     nil,
		Addons:                  nil,
		AddonsConfig:            v1alpha5.AddonsConfig{
			AutoApplyPodIdentityAssociations: false,
			DisableDefaultAddons:             false,
		},
		PrivateCluster:          nil,
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name:                        name + "-workers",
					AMIFamily:                   "",
					InstanceType:                "m5.large",
					AvailabilityZones:           nil,
					Subnets:                     nil,
					InstancePrefix:              "",
					InstanceName:                "",
					VolumeSize:                  nil,
					SSH:                         nil,
					Labels:                      nil,
					PrivateNetworking:           false,
					Tags:                        nil,
					IAM:                         nil,
					AMI:                         "",
					SecurityGroups:              nil,
					MaxPodsPerNode:              0,
					ASGSuspendProcesses:         nil,
					EBSOptimized:                nil,
					VolumeType:                  nil,
					VolumeName:                  nil,
					VolumeEncrypted:             nil,
					VolumeKmsKeyID:              nil,
					VolumeIOPS:                  nil,
					VolumeThroughput:            nil,
					AdditionalVolumes:           nil,
					PreBootstrapCommands:        nil,
					OverrideBootstrapCommand:    nil,
					PropagateASGTags:            nil,
					DisableIMDSv1:               nil,
					DisablePodIMDS:              nil,
					Placement:                   nil,
					EFAEnabled:                  nil,
					InstanceSelector:            nil,
					AdditionalEncryptedVolume:   "",
					Bottlerocket:                nil,
					EnableDetailedMonitoring:    nil,
					CapacityReservation:         nil,
					InstanceMarketOptions:       nil,
					OutpostARN:                  "",
					ScalingConfig: &v1alpha5.ScalingConfig{
						MinSize:         &minNodes,
						MaxSize:         &maxNodes,
						DesiredCapacity: &desiredNodes,
					},
				},
			},
		},
		ManagedNodeGroups:   nil,
		FargateProfiles:     nil,
		AvailabilityZones:   nil,
		LocalZones:          nil,
		CloudWatch:          nil,
		SecretsEncryption:   nil,
		Status:              nil,
		GitOps:              nil,
		Karpenter:           nil,
		Outpost:             nil,
		ZonalShiftConfig:    nil,
	}
}

// createTestClusterConfigWithOptions creates a test cluster config with custom EKS options.
func createTestClusterConfigWithOptions(name string) *v1alpha5.ClusterConfig {
	minNodes := 2
	maxNodes := 5
	desiredNodes := 3

	return &v1alpha5.ClusterConfig{
		TypeMeta: v1alpha5.ClusterConfigTypeMeta(),
		Metadata: &v1alpha5.ClusterMeta{
			Name:                name,
			Region:              "us-east-1",
			Version:             "1.25",
			ForceUpdateVersion:  nil,
			Tags:                nil,
			Annotations:         nil,
			AccountID:           "",
		},
		KubernetesNetworkConfig: nil,
		AutoModeConfig:          nil,
		RemoteNetworkConfig:     nil,
		IAM:                     nil,
		IAMIdentityMappings:     nil,
		IdentityProviders:       nil,
		AccessConfig:            nil,
		VPC:                     nil,
		Addons:                  nil,
		AddonsConfig:            v1alpha5.AddonsConfig{
			AutoApplyPodIdentityAssociations: false,
			DisableDefaultAddons:             false,
		},
		PrivateCluster:          nil,
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name:                        name + "-workers",
					AMIFamily:                   "",
					InstanceType:                "t3.medium",
					AvailabilityZones:           nil,
					Subnets:                     nil,
					InstancePrefix:              "",
					InstanceName:                "",
					VolumeSize:                  nil,
					SSH:                         nil,
					Labels:                      nil,
					PrivateNetworking:           false,
					Tags:                        nil,
					IAM:                         nil,
					AMI:                         "",
					SecurityGroups:              nil,
					MaxPodsPerNode:              0,
					ASGSuspendProcesses:         nil,
					EBSOptimized:                nil,
					VolumeType:                  nil,
					VolumeName:                  nil,
					VolumeEncrypted:             nil,
					VolumeKmsKeyID:              nil,
					VolumeIOPS:                  nil,
					VolumeThroughput:            nil,
					AdditionalVolumes:           nil,
					PreBootstrapCommands:        nil,
					OverrideBootstrapCommand:    nil,
					PropagateASGTags:            nil,
					DisableIMDSv1:               nil,
					DisablePodIMDS:              nil,
					Placement:                   nil,
					EFAEnabled:                  nil,
					InstanceSelector:            nil,
					AdditionalEncryptedVolume:   "",
					Bottlerocket:                nil,
					EnableDetailedMonitoring:    nil,
					CapacityReservation:         nil,
					InstanceMarketOptions:       nil,
					OutpostARN:                  "",
					ScalingConfig: &v1alpha5.ScalingConfig{
						MinSize:         &minNodes,
						MaxSize:         &maxNodes,
						DesiredCapacity: &desiredNodes,
					},
				},
			},
		},
		ManagedNodeGroups:   nil,
		FargateProfiles:     nil,
		AvailabilityZones:   nil,
		LocalZones:          nil,
		CloudWatch:          nil,
		SecretsEncryption:   nil,
		Status:              nil,
		GitOps:              nil,
		Karpenter:           nil,
		Outpost:             nil,
		ZonalShiftConfig:    nil,
	}
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