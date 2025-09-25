package eksprovisioner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	eksprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/eks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/eksctl/pkg/actions/cluster"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
)

// Define static errors as per err113 linter requirement.
var (
	errCreateFailed       = errors.New("create failed")
	errDeleteFailed       = errors.New("delete failed")
	errScaleFailed        = errors.New("scale failed")
	errListClustersFailed = errors.New("list clusters failed")
	errListFailed         = errors.New("list failed")
)

func TestNewEKSClusterProvisioner(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		testNewEKSClusterProvisionerSuccess(t)
	})
}

func testNewEKSClusterProvisionerSuccess(t *testing.T) {
	t.Helper()

	clusterConfig, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := setupMocks(
		t,
	)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)

	require.NotNil(t, provisioner, "NewEKSClusterProvisioner should return non-nil provisioner")
}

func TestCreate(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		testCreateSuccess(t)
	})
	t.Run("invalid config", func(t *testing.T) {
		t.Parallel()
		testCreateInvalidConfig(t)
	})
	t.Run("empty cluster name", func(t *testing.T) {
		t.Parallel()
		testCreateEmptyClusterName(t)
	})
	t.Run("create error", func(t *testing.T) {
		t.Parallel()
		testCreateError(t)
	})
}

func testCreateSuccess(t *testing.T) {
	t.Helper()

	provisioner, _ := setupCreateTest(t, nil)

	err := provisioner.Create(context.Background(), "test-cluster")

	require.NoError(t, err, "Create should succeed")
}

func testCreateInvalidConfig(t *testing.T) {
	t.Helper()

	provisioner := createProvisionerWithNilConfig(t)

	err := provisioner.Create(context.Background(), "test-cluster")

	assert.ErrorIs(
		t,
		err,
		eksprovisioner.ErrInvalidClusterConfig,
		"Create should fail with invalid config",
	)
}

func testCreateEmptyClusterName(t *testing.T) {
	t.Helper()

	provisioner, clusterProvider, clusterActions, clusterLister, creator, nodeGroupManager := setupProvisioner(
		t,
	)
	_ = clusterProvider  // not used in this test
	_ = clusterActions   // not used in this test
	_ = clusterLister    // not used in this test
	_ = nodeGroupManager // not used in this test

	// Expect successful creation with default name "ksail-default"
	creator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := provisioner.Create(context.Background(), "")

	assert.NoError(t, err, "Create should succeed with empty cluster name using default")
}

func testCreateError(t *testing.T) {
	t.Helper()

	provisioner, _ := setupCreateTest(t, errCreateFailed)

	err := provisioner.Create(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errCreateFailed,
		"failed to create EKS cluster",
		"Create()",
	)
}

func TestDelete(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		testDeleteSuccess(t)
	})
	t.Run("invalid config", func(t *testing.T) {
		t.Parallel()
		testDeleteInvalidConfig(t)
	})
	t.Run("empty cluster name", func(t *testing.T) {
		t.Parallel()
		testDeleteEmptyClusterName(t)
	})
	t.Run("delete error", func(t *testing.T) {
		t.Parallel()
		testDeleteError(t)
	})
}

func testDeleteSuccess(t *testing.T) {
	t.Helper()

	provisioner, _ := setupDeleteTest(t, nil)

	err := provisioner.Delete(context.Background(), "test-cluster")

	require.NoError(t, err, "Delete should succeed")
}

func testDeleteInvalidConfig(t *testing.T) {
	t.Helper()
	provisioner := createProvisionerWithNilConfig(t)

	err := provisioner.Delete(context.Background(), "test-cluster")

	assert.ErrorIs(
		t,
		err,
		eksprovisioner.ErrInvalidClusterConfig,
		"Delete should fail with invalid config",
	)
}

func testDeleteEmptyClusterName(t *testing.T) {
	t.Helper()

	provisioner, clusterProvider, actions, clusterLister, creator, nodeGroupManager := setupProvisioner(
		t,
	)
	_ = clusterProvider  // not used in this test
	_ = clusterLister    // not used in this test
	_ = creator          // not used in this test
	_ = nodeGroupManager // not used in this test

	// Expect successful deletion with default name "ksail-default"
	actions.On("Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	err := provisioner.Delete(context.Background(), "")

	assert.NoError(t, err, "Delete should succeed with empty cluster name using default")
}

func testDeleteError(t *testing.T) {
	t.Helper()

	provisioner, _ := setupDeleteTest(t, errDeleteFailed)

	err := provisioner.Delete(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errDeleteFailed,
		"failed to delete EKS cluster",
		"Delete()",
	)
}

func TestStart(t *testing.T) {
	t.Parallel()
	runParallelSubtests(t, map[string]func(*testing.T){
		"success":                          testStartSuccess,
		"success with multiple nodegroups": testStartSuccessMultipleNodegroups,
		"cluster not found":                testStartClusterNotFound,
		"nodegroup scale error":            testStartNodegroupScaleError,
		"success without scaling config":   testStartSuccessWithoutScalingConfig,
		"exists check error":               testStartExistsCheckError,
	})
}

func testStartSuccess(t *testing.T) {
	t.Helper()

	provisioner, _, _ := setupNodegroupScaleTest(t, nil)

	err := provisioner.Start(context.Background(), "test-cluster")

	require.NoError(t, err, "Start should succeed")
}

func testStartSuccessMultipleNodegroups(t *testing.T) {
	t.Helper()

	desiredCapacity1, desiredCapacity2 := 2, 3
	minSize := 0
	provisioner, lister, nodeGroupManager := setupProvisionerForNodegroupTests(
		t,
		[]*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name: "ng1",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: &desiredCapacity1,
						MinSize:         &minSize,
					},
				},
			},
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name: "ng2",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: &desiredCapacity2,
						MinSize:         &minSize,
					},
				},
			},
		},
	)

	setupExistsSuccess(lister)
	nodeGroupManager.On("Scale", mock.Anything, mock.AnythingOfType("*v1alpha5.NodeGroupBase"), true).
		Return(nil).
		Times(2)

	err := provisioner.Start(context.Background(), "test-cluster")

	require.NoError(t, err, "Start should succeed with multiple nodegroups")
}

func testStartClusterNotFound(t *testing.T) {
	t.Helper()

	provisioner, lister := setupProvisionerForBasicTests(t)

	setupExistsFailure(lister)

	err := provisioner.Start(context.Background(), "test-cluster")

	assert.ErrorIs(
		t,
		err,
		eksprovisioner.ErrClusterNotFound,
		"Start should fail when cluster not found",
	)
}

func testStartNodegroupScaleError(t *testing.T) {
	t.Helper()

	provisioner, _, _ := setupNodegroupScaleTest(t, errScaleFailed)

	err := provisioner.Start(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errScaleFailed,
		"failed to scale node group ng1",
		"Start()",
	)
}

func testStartSuccessWithoutScalingConfig(t *testing.T) {
	t.Helper()
	provisioner, lister, _ := setupProvisionerForNodegroupTests(
		t,
		[]*v1alpha5.NodeGroup{{
			NodeGroupBase: &v1alpha5.NodeGroupBase{
				Name: "ng1",
				// No ScalingConfig
			},
		}},
	)

	setupExistsSuccess(lister)

	err := provisioner.Start(context.Background(), "test-cluster")

	require.NoError(t, err, "Start should succeed without scaling config")
}

func testStartExistsCheckError(t *testing.T) {
	t.Helper()

	provisioner, _ := setupExistsCheckErrorTest(t)

	err := provisioner.Start(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errListClustersFailed,
		"failed to check if cluster exists",
		"Start()",
	)
}

func TestStop(t *testing.T) {
	t.Parallel()
	runParallelSubtests(t, map[string]func(*testing.T){
		"success":                          testStopSuccess,
		"success with multiple nodegroups": testStopSuccessMultipleNodegroups,
		"cluster not found":                testStopClusterNotFound,
		"nodegroup scale error":            testStopNodegroupScaleError,
		"success creates scaling config":   testStopSuccessCreatesScalingConfig,
		"exists check error":               testStopExistsCheckError,
	})
}

func testStopSuccess(t *testing.T) {
	t.Helper()

	provisioner, _, _ := setupNodegroupScaleTestWithMinSize(t, 1, nil)

	err := provisioner.Stop(context.Background(), "test-cluster")

	require.NoError(t, err, "Stop should succeed")
}

func testStopSuccessMultipleNodegroups(t *testing.T) {
	t.Helper()

	desiredCapacity1, desiredCapacity2 := 2, 3
	minSize1, minSize2 := 1, 1
	provisioner, lister, nodeGroupManager := setupProvisionerForNodegroupTests(
		t,
		[]*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name: "ng1",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: &desiredCapacity1,
						MinSize:         &minSize1,
					},
				},
			},
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name: "ng2",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: &desiredCapacity2,
						MinSize:         &minSize2,
					},
				},
			},
		},
	)

	setupExistsSuccess(lister)
	nodeGroupManager.On("Scale", mock.Anything, mock.AnythingOfType("*v1alpha5.NodeGroupBase"), true).
		Return(nil).
		Times(2)

	err := provisioner.Stop(context.Background(), "test-cluster")

	require.NoError(t, err, "Stop should succeed with multiple nodegroups")
}

func testStopClusterNotFound(t *testing.T) {
	t.Helper()
	provisioner, lister := setupProvisionerForBasicTests(t)

	setupExistsFailure(lister)

	err := provisioner.Stop(context.Background(), "test-cluster")

	assert.ErrorIs(
		t,
		err,
		eksprovisioner.ErrClusterNotFound,
		"Stop should fail when cluster not found",
	)
}

func testStopNodegroupScaleError(t *testing.T) {
	t.Helper()

	provisioner, _, _ := setupNodegroupScaleTestWithMinSize(t, 1, errScaleFailed)

	err := provisioner.Stop(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errScaleFailed,
		"failed to scale down node group ng1",
		"Stop()",
	)
}

func testStopSuccessCreatesScalingConfig(t *testing.T) {
	t.Helper()
	provisioner, lister, nodeGroupManager := setupProvisionerForNodegroupTests(
		t,
		[]*v1alpha5.NodeGroup{{
			NodeGroupBase: &v1alpha5.NodeGroupBase{
				Name: "ng1",
				// No ScalingConfig - will be created
			},
		}},
	)

	setupExistsSuccess(lister)
	nodeGroupManager.On("Scale", mock.Anything, mock.AnythingOfType("*v1alpha5.NodeGroupBase"), true).
		Return(nil)

	err := provisioner.Stop(context.Background(), "test-cluster")

	require.NoError(t, err, "Stop should succeed and create scaling config")
}

func testStopExistsCheckError(t *testing.T) {
	t.Helper()
	provisioner, _ := setupExistsCheckErrorTest(t)

	err := provisioner.Stop(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errListClustersFailed,
		"failed to check if cluster exists",
		"Stop()",
	)
}

func TestList(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		testListSuccess(t)
	})
	t.Run("success empty", func(t *testing.T) {
		t.Parallel()
		testListSuccessEmpty(t)
	})
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		testListError(t)
	})
}

func testListSuccess(t *testing.T) {
	t.Helper()

	expectedClusters := []cluster.Description{
		{Name: "cluster1"},
		{Name: "cluster2"},
	}
	provisioner, _ := setupListerTest(t, expectedClusters)

	clusters, err := provisioner.List(context.Background())

	require.NoError(t, err, "List should succeed")
	assert.Equal(t, []string{"cluster1", "cluster2"}, clusters, "List should return cluster names")
}

func testListSuccessEmpty(t *testing.T) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, lister, clusterCreator, nodeGroupManager := setupProvisioner(
		t,
	)
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).
		Return([]cluster.Description{}, nil)

	clusters, err := provisioner.List(context.Background())

	require.NoError(t, err, "List should succeed")
	assert.Empty(t, clusters, "List should return empty slice")
}

func testListError(t *testing.T) {
	t.Helper()
	provisioner, _ := setupListerErrorTest(t)

	_, err := provisioner.List(context.Background())

	testutils.AssertErrWrappedContains(
		t,
		err,
		errListFailed,
		"failed to list EKS clusters",
		"List()",
	)
}

func TestExists(t *testing.T) {
	t.Parallel()
	t.Run("true with name", func(t *testing.T) {
		t.Parallel()
		testExistsTrueWithName(t)
	})
	t.Run("true default name", func(t *testing.T) {
		t.Parallel()
		testExistsTrueDefaultName(t)
	})
	t.Run("false", func(t *testing.T) {
		t.Parallel()
		testExistsFalse(t)
	})
	t.Run("list error", func(t *testing.T) {
		t.Parallel()
		testExistsListError(t)
	})
}

func testExistsTrueWithName(t *testing.T) {
	t.Helper()

	expectedClusters := []cluster.Description{
		{Name: "test-cluster"},
		{Name: "other-cluster"},
	}
	provisioner, _ := setupListerTest(t, expectedClusters)

	exists, err := provisioner.Exists(context.Background(), "test-cluster")

	require.NoError(t, err, "Exists should succeed")
	assert.True(t, exists, "Exists should return true")
}

func testExistsTrueDefaultName(t *testing.T) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, lister, clusterCreator, nodeGroupManager := setupProvisioner(
		t,
	)
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	expectedClusters := []cluster.Description{
		{Name: "default-cluster"}, // This matches what setupMocks creates
		{Name: "other-cluster"},
	}
	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).Return(expectedClusters, nil)

	exists, err := provisioner.Exists(context.Background(), "")

	require.NoError(t, err, "Exists should succeed")
	assert.True(t, exists, "Exists should return true for default name")
}

func testExistsFalse(t *testing.T) {
	t.Helper()

	expectedClusters := []cluster.Description{
		{Name: "other-cluster"},
	}
	provisioner, _ := setupListerTest(t, expectedClusters)

	exists, err := provisioner.Exists(context.Background(), "test-cluster")

	require.NoError(t, err, "Exists should succeed")
	assert.False(t, exists, "Exists should return false")
}

func testExistsListError(t *testing.T) {
	t.Helper()
	provisioner, _ := setupListerErrorTest(t)

	_, err := provisioner.Exists(context.Background(), "test-cluster")

	testutils.AssertErrWrappedContains(t, err, errListFailed, "failed to list clusters", "Exists()")
}

// Helper functions

// runParallelSubtests runs a set of parallel subtests.
func runParallelSubtests(t *testing.T, tests map[string]func(*testing.T)) {
	t.Helper()

	for name, testFunc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testFunc(t)
		})
	}
}

func setupMocks(t *testing.T) (
	*v1alpha5.ClusterConfig,
	*eks.ClusterProvider,
	*eksprovisioner.MockEKSClusterActions,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()

	clusterConfig := &v1alpha5.ClusterConfig{
		Metadata: &v1alpha5.ClusterMeta{Name: "default-cluster"},
	}
	clusterProvider := &eks.ClusterProvider{}
	clusterActions := eksprovisioner.NewMockEKSClusterActions(t)
	clusterLister := eksprovisioner.NewMockEKSClusterLister(t)
	clusterCreator := eksprovisioner.NewMockEKSClusterCreator(t)
	nodeGroupManager := eksprovisioner.NewMockEKSNodeGroupManager(t)

	return clusterConfig, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager
}

func setupProvisioner(t *testing.T) (
	*eksprovisioner.EKSClusterProvisioner,
	*eks.ClusterProvider,
	*eksprovisioner.MockEKSClusterActions,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	clusterConfig, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := setupMocks(
		t,
	)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)

	return provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager
}

// setupProvisionerForNodegroupTests returns only what's needed for nodegroup scaling tests.
func setupProvisionerForNodegroupTests(t *testing.T, nodeGroups []*v1alpha5.NodeGroup) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	clusterConfig, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := setupMocks(
		t,
	)
	clusterConfig.NodeGroups = nodeGroups

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

// setupProvisionerForBasicTests creates a basic provisioner and returns only
// the components commonly used by basic tests (no nodegroups).
func setupProvisionerForBasicTests(
	t *testing.T,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterLister,
) {
	t.Helper()

	clusterConfig, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := setupMocks(
		t,
	)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		clusterProvider,
		clusterActions,
		clusterLister,
		clusterCreator,
		nodeGroupManager,
	)

	return provisioner, clusterLister
}

func setupExistsSuccess(lister *eksprovisioner.MockEKSClusterLister) {
	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).Return([]cluster.Description{
		{Name: "default-cluster"},
		{Name: "test-cluster"},
	}, nil)
}

func setupExistsFailure(lister *eksprovisioner.MockEKSClusterLister) {
	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).Return([]cluster.Description{
		{Name: "other-cluster"},
	}, nil)
}

// setupProvisionerWithUsedMocks returns a provisioner and the specific mocks that will be used
// This helper reduces duplication for tests that only need specific mocks.
func setupProvisionerWithUsedMocks(t *testing.T, usedMocks ...string) (
	*eksprovisioner.EKSClusterProvisioner,
	*eks.ClusterProvider,
	*eksprovisioner.MockEKSClusterActions,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, clusterLister, creator, nodeGroupManager := setupProvisioner(
		t,
	)

	// Mark unused mocks to avoid "not used" comments
	usedMockSet := make(map[string]bool)
	for _, mock := range usedMocks {
		usedMockSet[mock] = true
	}

	if !usedMockSet["clusterProvider"] {
		_ = clusterProvider
	}

	if !usedMockSet["clusterActions"] {
		_ = clusterActions
	}

	if !usedMockSet["clusterLister"] {
		_ = clusterLister
	}

	if !usedMockSet["creator"] {
		_ = creator
	}

	if !usedMockSet["nodeGroupManager"] {
		_ = nodeGroupManager
	}

	return provisioner, clusterProvider, clusterActions, clusterLister, creator, nodeGroupManager
}

// createProvisionerWithNilConfig creates a provisioner with nil config for error testing.
func createProvisionerWithNilConfig(t *testing.T) *eksprovisioner.EKSClusterProvisioner {
	t.Helper()

	return eksprovisioner.NewEKSClusterProvisioner(
		nil, // nil config
		&eks.ClusterProvider{},
		eksprovisioner.NewMockEKSClusterActions(t),
		eksprovisioner.NewMockEKSClusterLister(t),
		eksprovisioner.NewMockEKSClusterCreator(t),
		eksprovisioner.NewMockEKSNodeGroupManager(t),
	)
}

// setupCreateTest sets up provisioner and creator mock for create tests.
func setupCreateTest(
	t *testing.T,
	returnError error,
) (*eksprovisioner.EKSClusterProvisioner, *eksprovisioner.MockEKSClusterCreator) {
	t.Helper()
	prov, cp, ca, cl, cr, ng := setupProvisionerWithUsedMocks(t, "creator")
	_ = cp
	_ = ca
	_ = cl
	_ = ng
	provisioner, creator := prov, cr

	creator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(returnError)

	return provisioner, creator
}

// setupDeleteTest sets up provisioner and actions mock for delete tests.
func setupDeleteTest(
	t *testing.T,
	returnError error,
) (*eksprovisioner.EKSClusterProvisioner, *eksprovisioner.MockEKSClusterActions) {
	t.Helper()
	prov, cp, actions, cl, cr, ng := setupProvisionerWithUsedMocks(t, "actions")
	_ = cp
	_ = cl
	_ = cr
	_ = ng
	provisioner := prov

	actions.On("Delete",
		mock.Anything,                        // context
		mock.AnythingOfType("time.Duration"), // waitInterval
		mock.AnythingOfType("time.Duration"), // podEvictionWaitPeriod
		mock.AnythingOfType("bool"),          // wait
		mock.AnythingOfType("bool"),          // force
		mock.AnythingOfType("bool"),          // disableNodegroupEviction
		mock.AnythingOfType("int"),           // parallel
	).Return(returnError)

	return provisioner, actions
}

// setupNodegroupScaleTest sets up provisioner with single nodegroup for scale tests.
func setupNodegroupScaleTest(
	t *testing.T,
	returnError error,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()

	return setupNodegroupScaleTestWithMinSize(t, 0, returnError)
}

// setupNodegroupScaleTestWithMinSize sets up provisioner with single nodegroup for scale tests
// with configurable minSize.
func setupNodegroupScaleTestWithMinSize(
	t *testing.T,
	minSize int,
	returnError error,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()

	desiredCapacity := 1
	provisioner, lister, nodeGroupManager := setupProvisionerForNodegroupTests(
		t,
		[]*v1alpha5.NodeGroup{{
			NodeGroupBase: &v1alpha5.NodeGroupBase{
				Name: "ng1",
				ScalingConfig: &v1alpha5.ScalingConfig{
					DesiredCapacity: &desiredCapacity,
					MinSize:         &minSize,
				},
			},
		}},
	)

	setupExistsSuccess(lister)
	nodeGroupManager.On("Scale", mock.Anything, mock.AnythingOfType("*v1alpha5.NodeGroupBase"), true).
		Return(returnError)

	return provisioner, lister, nodeGroupManager
}

// setupListerTest sets up provisioner and lister with expected clusters for list/exists tests.
func setupListerTest(
	t *testing.T,
	expectedClusters []cluster.Description,
) (*eksprovisioner.EKSClusterProvisioner, *eksprovisioner.MockEKSClusterLister) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, lister, creator, nodeGroupManager := setupProvisionerWithUsedMocks(
		t,
		"lister",
	)
	_ = clusterProvider
	_ = clusterActions
	_ = creator
	_ = nodeGroupManager

	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).Return(expectedClusters, nil)

	return provisioner, lister
}

// setupListerErrorTest sets up provisioner and lister for testing GetClusters errors.
func setupListerErrorTest(
	t *testing.T,
) (*eksprovisioner.EKSClusterProvisioner, *eksprovisioner.MockEKSClusterLister) {
	t.Helper()
	prov, cp, ca, lister, cr, ng := setupProvisionerWithUsedMocks(t, "lister")
	_ = cp
	_ = ca
	_ = cr
	_ = ng
	provisioner := prov

	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).Return(nil, errListFailed)

	return provisioner, lister
}

// setupExistsCheckErrorTest sets up provisioner and lister for exists check error tests.
func setupExistsCheckErrorTest(
	t *testing.T,
) (*eksprovisioner.EKSClusterProvisioner, *eksprovisioner.MockEKSClusterLister) {
	t.Helper()
	provisioner, lister := setupProvisionerForBasicTests(t)
	lister.On("GetClusters", mock.Anything, mock.Anything, false, 100).
		Return(nil, errListClustersFailed)

	return provisioner, lister
}
