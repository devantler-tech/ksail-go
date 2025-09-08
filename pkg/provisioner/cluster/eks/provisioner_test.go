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

// setupEKSProvisioner is a helper function that creates an EKS provisioner and mock cluster creator for testing.
// This supports the shared test pattern for EKS tests.
func setupEKSProvisioner(t *testing.T) (*eksprovisioner.EKSClusterProvisioner, *eksprovisioner.MockEKSClusterCreator) {
	t.Helper()
	provisioner, _, nodeGroupManager, clusterCreator, _ := newProvisionerForTest(t)
	_ = nodeGroupManager // Explicitly ignore to satisfy dogsled

	return provisioner, clusterCreator
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	clustertestutils.RunCreateSuccessTest(
		t,
		setupEKSProvisioner,
		func(clusterCreator *eksprovisioner.MockEKSClusterCreator, _ string) {
			// No longer need to mock provider construction since it's injected directly
			clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		},
		func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
			return prov.Create(context.Background(), name)
		},
	)
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

	provisioner, _ := setupProvisionerWithEmptyClusterList(t)

	err := provisioner.Start(context.Background(), "test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrClusterNotFound)
}

func TestStart_Success_WithMinSizeZero(t *testing.T) {
	t.Parallel()

	// Create a test config with MinSize = 0
	clusterConfig := createTestProvisionerClusterConfig()
	minSize := 0
	desiredCapacity := 2
	clusterConfig.NodeGroups[0].MinSize = &minSize
	clusterConfig.NodeGroups[0].DesiredCapacity = &desiredCapacity

	provisioner, _ := setupCustomConfigWithScaling(t, clusterConfig)

	err := provisioner.Start(context.Background(), "cfg-name")

	require.NoError(t, err)
	// Verify that MinSize was updated to match DesiredCapacity
	assert.Equal(t, desiredCapacity, *clusterConfig.NodeGroups[0].MinSize)
}

func TestStart_Error_ScaleFailed(t *testing.T) {
	t.Parallel()

	provisioner, _ := setupScalingErrorTest(t, true)

	err := provisioner.Start(context.Background(), "cfg-name")

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrScaleNodeGroupFailed,
		"failed to scale node group", "Start()")
}

func TestStop_Success(t *testing.T) {
	t.Parallel()
	runNodeScalingTest(t, "Stop()", true, func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
		return prov.Stop(context.Background(), name)
	})
}

func TestStop_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()

	provisioner, _ := setupProvisionerWithEmptyClusterList(t)

	err := provisioner.Stop(context.Background(), "test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrClusterNotFound)
}

func TestStop_Success_WithNilScalingConfig(t *testing.T) {
	t.Parallel()

	// Create a test config with nil ScalingConfig
	clusterConfig := createTestProvisionerClusterConfig()
	clusterConfig.NodeGroups[0].ScalingConfig = nil // This should trigger the nil check

	provisioner, _ := setupCustomConfigWithScaling(t, clusterConfig)

	err := provisioner.Stop(context.Background(), "cfg-name")

	require.NoError(t, err)
	// Verify that ScalingConfig was created and values set to 0
	assert.NotNil(t, clusterConfig.NodeGroups[0].ScalingConfig)
	assert.Equal(t, 0, *clusterConfig.NodeGroups[0].DesiredCapacity)
	assert.Equal(t, 0, *clusterConfig.NodeGroups[0].MinSize)
}

func TestStop_Error_ScaleFailed(t *testing.T) {
	t.Parallel()

	provisioner, _ := setupScalingErrorTest(t, true)

	err := provisioner.Stop(context.Background(), "cfg-name")

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrScaleNodeGroupFailed,
		"failed to scale down node group", "Stop()")
}

func TestEnsureClusterExists_Error_ExistsCallFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	// Mock Exists() call failure (not just cluster not found)
	mockGetClusters(clusterLister, nil, clustertestutils.ErrListClustersFailed)

	err := provisioner.Start(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to check if cluster exists", "Start()")
}

func TestSetupNodeGroupManager_Error_SetupClusterOperationFailed(t *testing.T) {
	t.Parallel()

	// Test the setupClusterOperation failure in setupNodeGroupManager by 
	// creating a provisioner with empty metadata name and empty input name
	clusterConfig := &v1alpha5.ClusterConfig{
		Metadata: &v1alpha5.ClusterMeta{
			Name: "", // empty name in metadata
		},
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name: "test",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: new(int), // non-nil to trigger scaling
					},
				},
			},
		},
	}

	clusterLister := eksprovisioner.NewMockEKSClusterLister(t)
	// Mock that cluster exists (this succeeds in ensureClusterExists)
	descriptions := []cluster.Description{{Name: "", Region: "", Owned: ""}}
	mockGetClusters(clusterLister, descriptions, nil)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		nil,
		eksprovisioner.NewMockEKSClusterActions(t),
		clusterLister,
		eksprovisioner.NewMockEKSClusterCreator(t),
		eksprovisioner.NewMockEKSNodeGroupManager(t),
	)

	err := provisioner.Start(context.Background(), "") // empty input name with empty metadata name

	assert.ErrorIs(t, err, eksprovisioner.ErrEmptyClusterName)
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

	provisioner, _ := setupProvisionerWithEmptyClusterList(t)

	exists, err := provisioner.Exists(context.Background(), "nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestExists_Error_ListFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	mockGetClusters(clusterLister, nil, clustertestutils.ErrListClustersFailed)

	exists, err := provisioner.Exists(context.Background(), "test-cluster")

	assert.False(t, exists)
	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list clusters", "Exists()")
}

func TestCreate_Error_InvalidClusterConfig(t *testing.T) {
	t.Parallel()

	provisioner := createProvisionerWithInvalidConfig(t)

	err := provisioner.Create(context.Background(), "test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrInvalidClusterConfig)
}

func TestCreate_Error_NilMetadata(t *testing.T) {
	t.Parallel()

	provisioner := createProvisionerWithNilMetadata(t)

	err := provisioner.Create(context.Background(), "test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrInvalidClusterConfig)
}

func TestCreate_Error_EmptyClusterName(t *testing.T) {
	t.Parallel()

	provisioner := createProvisionerWithEmptyName(t)

	err := provisioner.Create(context.Background(), "") // empty input name too

	assert.ErrorIs(t, err, eksprovisioner.ErrEmptyClusterName)
}

func TestDelete_Error_InvalidClusterConfig(t *testing.T) {
	t.Parallel()

	provisioner := createProvisionerWithInvalidConfig(t)

	err := provisioner.Delete(context.Background(), "test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrInvalidClusterConfig)
}

// --- test helpers ---

// createProvisionerWithInvalidConfig creates a provisioner with nil cluster config for error testing.
func createProvisionerWithInvalidConfig(t *testing.T) *eksprovisioner.EKSClusterProvisioner {
	t.Helper()
	clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := createMocksForTest(t)
	return eksprovisioner.NewEKSClusterProvisioner(
		nil, // nil clusterConfig
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)
}

// createProvisionerWithNilMetadata creates a provisioner with nil metadata for error testing.
func createProvisionerWithNilMetadata(t *testing.T) *eksprovisioner.EKSClusterProvisioner {
	t.Helper()
	clusterConfig := &v1alpha5.ClusterConfig{
		Metadata: nil, // nil metadata
	}
	clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := createMocksForTest(t)
	return eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)
}

// createProvisionerWithEmptyName creates a provisioner with empty name in metadata for error testing.
func createProvisionerWithEmptyName(t *testing.T) *eksprovisioner.EKSClusterProvisioner {
	t.Helper()
	clusterConfig := &v1alpha5.ClusterConfig{
		Metadata: &v1alpha5.ClusterMeta{
			Name: "", // empty name
		},
	}
	clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := createMocksForTest(t)
	return eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)
}

// setupCustomConfigWithScaling sets up a provisioner with custom config and mocks cluster exists and scaling.
func setupCustomConfigWithScaling(
	t *testing.T,
	clusterConfig *v1alpha5.ClusterConfig,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	provisioner, clusterLister, nodeGroupManager := setupProvisionerWithCustomConfig(t, clusterConfig)

	// Mock cluster exists
	descriptions := []cluster.Description{{Name: "cfg-name", Region: "", Owned: ""}}
	mockGetClusters(clusterLister, descriptions, nil)

	// Mock node group scaling
	setupNodeGroupScalingMock(nodeGroupManager, true)

	return provisioner, nodeGroupManager
}

// createMocksForTest creates all the mock dependencies for an EKS provisioner.
func createMocksForTest(t *testing.T) (
	*eks.ClusterProvider,
	*eksprovisioner.MockEKSClusterActions,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	// For tests, we can use a nil ClusterProvider since the actual provider methods are mocked
	clusterProvider := (*eks.ClusterProvider)(nil)
	clusterActions := eksprovisioner.NewMockEKSClusterActions(t)
	clusterLister := eksprovisioner.NewMockEKSClusterLister(t)
	clusterCreator := eksprovisioner.NewMockEKSClusterCreator(t)
	nodeGroupManager := eksprovisioner.NewMockEKSNodeGroupManager(t)

	return clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager
}

// setupProvisionerWithCustomConfig creates a provisioner with a custom cluster config and returns it with mocks.
func setupProvisionerWithCustomConfig(
	t *testing.T,
	clusterConfig *v1alpha5.ClusterConfig,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := createMocksForTest(t)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)

	return provisioner, clusterLister, nodeGroupManager
}

// setupScalingErrorTest sets up a test for node group scaling errors.
func setupScalingErrorTest(
	t *testing.T,
	scaleUp bool,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner, clusterLister, and nodeGroupManager for scaling error tests
	_ = clusterActions
	_ = clusterCreator

	// Mock cluster exists
	descriptions := []cluster.Description{{Name: "cfg-name", Region: "", Owned: ""}}
	mockGetClusters(clusterLister, descriptions, nil)

	// Mock node group scaling failure
	nodeGroupManager.On("Scale", mock.Anything, mock.Anything, scaleUp).
		Return(clustertestutils.ErrScaleNodeGroupFailed)

	return provisioner, nodeGroupManager
}

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
	clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := createMocksForTest(t)

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
		TypeMeta:     v1alpha5.ClusterConfigTypeMeta(),
		Metadata:     createTestProvisionerMetadata(),
		AddonsConfig: createTestProvisionerAddonsConfig(),
		NodeGroups:   createTestProvisionerNodeGroups(desiredCapacity),
	}
}

func createTestProvisionerMetadata() *v1alpha5.ClusterMeta {
	return &v1alpha5.ClusterMeta{
		Name:   "cfg-name",
		Region: "us-west-2",
	}
}

func createTestProvisionerAddonsConfig() v1alpha5.AddonsConfig {
	return v1alpha5.AddonsConfig{}
}

func createTestProvisionerNodeGroups(desiredCapacity int) []*v1alpha5.NodeGroup {
	return []*v1alpha5.NodeGroup{
		{
			NodeGroupBase: createTestProvisionerNodeGroupBase(desiredCapacity),
		},
	}
}

func createTestProvisionerNodeGroupBase(desiredCapacity int) *v1alpha5.NodeGroupBase {
	return clustertestutils.CreateTestEKSNodeGroupBase(clustertestutils.EKSNodeGroupBaseOptions{
		Name:            "test-nodegroup",
		InstanceType:    "",
		MinSize:         nil,
		MaxSize:         nil,
		DesiredCapacity: &desiredCapacity,
	})
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

// setupProvisionerWithEmptyClusterList creates a test provisioner and mocks an empty cluster list.
// This helper eliminates duplication for tests that need to verify behavior when no clusters exist.
func setupProvisionerWithEmptyClusterList(t *testing.T) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterLister,
) {
	t.Helper()

	provisioner, clusterActions, clusterLister, clusterCreator, nodeGroupManager :=
		newProvisionerForTest(t)
	// We only need provisioner and clusterLister for empty cluster list scenarios
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	mockGetClusters(clusterLister, []cluster.Description{}, nil)

	return provisioner, clusterLister
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
