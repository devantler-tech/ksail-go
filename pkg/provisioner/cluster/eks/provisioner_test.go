package eksprovisioner_test

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	eksprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/eks"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/eksctl/pkg/actions/cluster"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
)



func TestCreate_Success(t *testing.T) {
	t.Parallel()
	clustertestutils.RunCreateTest(t, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		runActionSuccess(
			t,
			"Create()",
			inputName,
			expectedName,
			func(clusterCreator *eksprovisioner.MockEKSClusterCreator, _ string) {
				// No longer need to mock provider construction since it's injected directly
				clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Create(context.Background(), name)
			},
		)
	})
}

func TestCreate_Error_CreateFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterCreator for this test
	_ = clusterActions
	_ = clusterLister
	_ = nodeGroupManager

	clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).
		Return(clustertestutils.ErrCreateClusterFailed)

	err := provisioner.Create(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrCreateClusterFailed,
		"failed to create EKS cluster", "Create()")
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	cases := clustertestutils.DefaultDeleteCases()
	clustertestutils.RunStandardSuccessTest(t, cases, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		runDeleteActionSuccess(
			t,
			"Delete()",
			inputName,
			expectedName,
			func(clusterActions *eksprovisioner.MockEKSClusterActions, _ string) {
				// No longer need to mock provider construction since it's injected directly

				// Delete(ctx context.Context, waitInterval, podEvictionWaitPeriod time.Duration,
				//   wait, force, disableNodegroupEviction bool, parallel int)
				mockClusterDeleteAction(clusterActions, nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Delete(context.Background(), name)
			},
		)
	})
}

func TestDelete_Error_CreateFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterActions for this test
	_ = clusterLister
	_ = clusterCreator
	_ = nodeGroupManager

	mockClusterDeleteAction(clusterActions, clustertestutils.ErrDeleteClusterFailed)

	err := provisioner.Delete(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrDeleteClusterFailed,
		"failed to delete EKS cluster", "Delete()")
}

func TestStart_Success(t *testing.T) {
	t.Parallel()
	runNodeScalingTest(t, "Start()", true, func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
		return prov.Start(context.Background(), name)
	})
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	mockGetClusters(clusterLister, []cluster.Description{}, nil)

	err := provisioner.Start(context.Background(), "test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrClusterNotFound)
}

func TestStop_Success(t *testing.T) {
	t.Parallel()
	runNodeScalingTest(t, "Stop()", true, func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
		return prov.Stop(context.Background(), name)
	})
}

func TestList_Success(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	descriptions := []cluster.Description{
		{Name: "cluster1", Region: "", Owned: ""},
		{Name: "cluster2", Region: "", Owned: ""},
	}
	mockGetClusters(clusterLister, descriptions, nil)

	clusters, err := provisioner.List(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"cluster1", "cluster2"}, clusters)
}

func TestList_Error_GetClustersFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	
	mockGetClusters(clusterLister, nil, clustertestutils.ErrListClustersFailed)
	
	clusters, err := provisioner.List(context.Background())

	assert.Nil(t, clusters)
	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list EKS clusters", "List()")
}

func TestExists_Success_True(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	descriptions := []cluster.Description{{Name: "cfg-name", Region: "", Owned: ""}}
	mockGetClusters(clusterLister, descriptions, nil)

	exists, err := provisioner.Exists(context.Background(), "cfg-name")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	mockGetClusters(clusterLister, []cluster.Description{}, nil)

	exists, err := provisioner.Exists(context.Background(), "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

// --- test helpers ---

func newProvisionerForTest(
	t *testing.T,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterActions,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()

	clusterConfig := createTestProvisionerClusterConfig()

	// For tests, we can use a nil ClusterProvider since the actual provider methods are mocked
	clusterProvider := (*eks.ClusterProvider)(nil)
	clusterActions := eksprovisioner.NewMockEKSClusterActions(t)
	clusterLister := eksprovisioner.NewMockEKSClusterLister(t)
	clusterCreator := eksprovisioner.NewMockEKSClusterCreator(t)
	nodeGroupManager := eksprovisioner.NewMockEKSNodeGroupManager(t)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)

	return provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager
}

func createTestProvisionerClusterConfig() *v1alpha5.ClusterConfig {
	desiredCapacity := 2

	return &v1alpha5.ClusterConfig{
		TypeMeta: v1alpha5.ClusterConfigTypeMeta(),
		Metadata: createTestProvisionerMetadata(),
		KubernetesNetworkConfig: nil,
		AutoModeConfig:          nil,
		RemoteNetworkConfig:     nil,
		IAM:                     nil,
		IAMIdentityMappings:     nil,
		IdentityProviders:       nil,
		AccessConfig:            nil,
		VPC:                     nil,
		Addons:                  nil,
		AddonsConfig:            createTestProvisionerAddonsConfig(),
		PrivateCluster:          nil,
		NodeGroups:              createTestProvisionerNodeGroups(desiredCapacity),
		ManagedNodeGroups:       nil,
		FargateProfiles:         nil,
		AvailabilityZones:       nil,
		LocalZones:              nil,
		CloudWatch:              nil,
		SecretsEncryption:       nil,
		Status:                  nil,
		GitOps:                  nil,
		Karpenter:               nil,
		Outpost:                 nil,
		ZonalShiftConfig:        nil,
	}
}

func createTestProvisionerMetadata() *v1alpha5.ClusterMeta {
	return &v1alpha5.ClusterMeta{
		Name:                "cfg-name",
		Region:              "us-west-2",
		Version:             "",
		ForceUpdateVersion:  nil,
		Tags:                nil,
		Annotations:         nil,
		AccountID:           "",
	}
}

func createTestProvisionerAddonsConfig() v1alpha5.AddonsConfig {
	return v1alpha5.AddonsConfig{
		AutoApplyPodIdentityAssociations: false,
		DisableDefaultAddons:             false,
	}
}

func createTestProvisionerNodeGroups(desiredCapacity int) []*v1alpha5.NodeGroup {
	return []*v1alpha5.NodeGroup{
		{
			NodeGroupBase: createTestProvisionerNodeGroupBase(desiredCapacity),
		},
	}
}

func createTestProvisionerNodeGroupBase(desiredCapacity int) *v1alpha5.NodeGroupBase {
	return &v1alpha5.NodeGroupBase{
		Name:                        "test-nodegroup",
		AMIFamily:                   "",
		InstanceType:                "",
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
			DesiredCapacity: &desiredCapacity,
			MinSize:         nil,
			MaxSize:         nil,
		},
	}
}

// mockClusterDeleteAction sets up the standard mock for Delete action on clusterActions.
func mockClusterDeleteAction(clusterActions *eksprovisioner.MockEKSClusterActions, returnErr error) {
	clusterActions.On(
		"Delete",
		mock.Anything,                         // ctx context.Context
		mock.AnythingOfType("time.Duration"),  // waitInterval time.Duration
		mock.AnythingOfType("time.Duration"),  // podEvictionWaitPeriod time.Duration
		mock.AnythingOfType("bool"),           // wait bool
		mock.AnythingOfType("bool"),           // force bool
		mock.AnythingOfType("bool"),           // disableNodegroupEviction bool
		mock.AnythingOfType("int"),            // parallel int
	).Return(returnErr)
}

// mockGetClusters sets up the standard mock for GetClusters action on clusterLister.
func mockGetClusters(
	clusterLister *eksprovisioner.MockEKSClusterLister,
	descriptions []cluster.Description,
	returnErr error,
) {
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(descriptions, returnErr)
}

// setupNodeGroupScalingMock sets up the standard mock for node group scaling operations.
func setupNodeGroupScalingMock(nodeGroupManager *eksprovisioner.MockEKSNodeGroupManager, scaleUp bool) {
	nodeGroupManager.On("Scale", mock.Anything, mock.Anything, scaleUp).Return(nil)
}

// runNodeScalingTest runs a standard test pattern for node scaling operations (Start/Stop).
func runNodeScalingTest(
	t *testing.T,
	testName string,
	scaleUp bool,
	action func(*eksprovisioner.EKSClusterProvisioner, string) error,
) {
	t.Helper()

	cases := clustertestutils.DefaultNameCases("cfg-name")
	clustertestutils.RunStandardSuccessTest(t, cases, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
			newProvisionerForTest(t)
		// We only need provisioner, clusterLister, and nodeGroupManager for this test
		_ = clusterActions
		_ = clusterCreator
		descriptions := []cluster.Description{{Name: expectedName, Region: "", Owned: ""}}
		mockGetClusters(clusterLister, descriptions, nil)

		// Mock node group manager
		setupNodeGroupScalingMock(nodeGroupManager, scaleUp)

		err := action(provisioner, inputName)
		if err != nil {
			t.Fatalf("%s unexpected error: %v", testName, err)
		}
	})
}

type expectProviderFn func(*eksprovisioner.MockEKSClusterCreator, string)
type actionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectProviderFn,
	action actionFn,
) {
	t.Helper()
	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need clusterCreator for this function
	_ = clusterActions
	_ = clusterLister
	_ = nodeGroupManager

	expect(clusterCreator, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

type expectDeleteProviderFn func(*eksprovisioner.MockEKSClusterActions, string)
type deleteActionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runDeleteActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectDeleteProviderFn,
	action deleteActionFn,
) {
	t.Helper()
	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need clusterActions for this function
	_ = clusterLister
	_ = clusterCreator
	_ = nodeGroupManager

	expect(clusterActions, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}
