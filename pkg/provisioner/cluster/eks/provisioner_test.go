package eksprovisioner_test

import (
	"errors"
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

var (
	errCreateClusterFailed = errors.New("create cluster failed")
	errDeleteClusterFailed = errors.New("delete cluster failed")
	errListClustersFailed  = errors.New("list clusters failed")
)

func TestCreate_Success(t *testing.T) {
	clustertestutils.RunCreateTest(t, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		runActionSuccess(
			t,
			"Create()",
			inputName,
			expectedName,
			func(_ *eks.ClusterProvider,
				clusterCreator *eksprovisioner.MockEKSClusterCreator, _ string) {
				// No longer need to mock provider construction since it's injected directly
				clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Create(name)
			},
		)
	})
}

func TestCreate_Error_CreateFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterCreator for this test
	_ = clusterProvider
	_ = clusterActions  
	_ = clusterLister
	_ = nodeGroupManager

	clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(errCreateClusterFailed)

	err := provisioner.Create("test-cluster")

	testutils.AssertErrWrappedContains(t, err, errCreateClusterFailed, "failed to create EKS cluster", "Create()")
}

func TestDelete_Success(t *testing.T) {
	cases := clustertestutils.DefaultDeleteCases()
	clustertestutils.RunStandardSuccessTest(t, cases, func(t *testing.T, inputName, expectedName string) {
		runDeleteActionSuccess(
			t,
			"Delete()",
			inputName,
			expectedName,
			func(_ *eks.ClusterProvider,
				clusterActions *eksprovisioner.MockEKSClusterActions, _ string) {
				// No longer need to mock provider construction since it's injected directly

				// Delete(ctx context.Context, waitInterval, podEvictionWaitPeriod time.Duration,
				//   wait, force, disableNodegroupEviction bool, parallel int)
				mockClusterDeleteAction(clusterActions, nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Delete(name)
			},
		)
	})
}

func TestDelete_Error_CreateFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterActions for this test
	_ = clusterProvider
	_ = clusterLister
	_ = clusterCreator
	_ = nodeGroupManager

	mockClusterDeleteAction(clusterActions, errDeleteClusterFailed)

	err := provisioner.Delete("test-cluster")

	testutils.AssertErrWrappedContains(t, err, errDeleteClusterFailed, "failed to delete EKS cluster", "Delete()")
}

func TestStart_Success(t *testing.T) {
	runNodeScalingTest(t, "Start()", true, func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
		return prov.Start(name)
	})
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test  
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	mockGetClusters(clusterLister, []cluster.Description{}, nil)

	err := provisioner.Start("test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrClusterNotFound)
}

func TestStop_Success(t *testing.T) {
	runNodeScalingTest(t, "Stop()", true, func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
		return prov.Stop(name)
	})
}

func TestList_Success(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	descriptions := []cluster.Description{
		{Name: "cluster1"},
		{Name: "cluster2"},
	}
	mockGetClusters(clusterLister, descriptions, nil)

	clusters, err := provisioner.List()

	require.NoError(t, err)
	assert.Equal(t, []string{"cluster1", "cluster2"}, clusters)
}

func TestList_Error_GetClustersFailed(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	
	mockGetClusters(clusterLister, nil, errListClustersFailed)
	
	clusters, err := provisioner.List()

	assert.Nil(t, clusters)
	testutils.AssertErrWrappedContains(t, err, errListClustersFailed, "failed to list EKS clusters", "List()")
}

func TestExists_Success_True(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	descriptions := []cluster.Description{{Name: "cfg-name"}}
	mockGetClusters(clusterLister, descriptions, nil)

	exists, err := provisioner.Exists("cfg-name")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()

	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need provisioner and clusterLister for this test
	_ = clusterProvider
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager

	mockGetClusters(clusterLister, []cluster.Description{}, nil)

	exists, err := provisioner.Exists("nonexistent")

	require.NoError(t, err)
	assert.False(t, exists)
}

// --- test helpers ---

func newProvisionerForTest(
	t *testing.T,
) (
	*eksprovisioner.EKSClusterProvisioner,
	*eks.ClusterProvider,
	*eksprovisioner.MockEKSClusterActions,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
	*eksprovisioner.MockEKSNodeGroupManager,
) {
	t.Helper()

	desiredCapacity := 2
	clusterConfig := &v1alpha5.ClusterConfig{
		Metadata: &v1alpha5.ClusterMeta{
			Name:   "cfg-name",
			Region: "us-west-2",
		},
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name: "test-nodegroup",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: &desiredCapacity,
					},
				},
			},
		},
	}

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

	return provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager
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
func mockGetClusters(clusterLister *eksprovisioner.MockEKSClusterLister, descriptions []cluster.Description, returnErr error) {
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, returnErr)
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
		provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
		// We only need provisioner, clusterLister, and nodeGroupManager for this test
		_ = clusterProvider
		_ = clusterActions
		_ = clusterCreator
		descriptions := []cluster.Description{{Name: expectedName}}
		mockGetClusters(clusterLister, descriptions, nil)

		// Mock node group manager
		setupNodeGroupScalingMock(nodeGroupManager, scaleUp)

		err := action(provisioner, inputName)
		if err != nil {
			t.Fatalf("%s unexpected error: %v", testName, err)
		}
	})
}

type expectProviderFn func(*eks.ClusterProvider, *eksprovisioner.MockEKSClusterCreator, string)
type actionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectProviderFn,
	action actionFn,
) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need clusterProvider and clusterCreator for this function
	_ = clusterActions
	_ = clusterLister
	_ = nodeGroupManager
	expect(clusterProvider, clusterCreator, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

type expectDeleteProviderFn func(*eks.ClusterProvider, *eksprovisioner.MockEKSClusterActions, string)
type deleteActionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runDeleteActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectDeleteProviderFn,
	action deleteActionFn,
) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need clusterProvider and clusterActions for this function
	_ = clusterLister
	_ = clusterCreator
	_ = nodeGroupManager
	expect(clusterProvider, clusterActions, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

type expectListProviderFn func(*eks.ClusterProvider, *eksprovisioner.MockEKSClusterLister, string)
type listActionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runListActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectListProviderFn,
	action listActionFn,
) {
	t.Helper()
	provisioner, clusterProvider, clusterActions, clusterLister, clusterCreator, nodeGroupManager := newProvisionerForTest(t)
	// We only need clusterProvider and clusterLister for this function
	_ = clusterActions
	_ = clusterCreator
	_ = nodeGroupManager
	expect(clusterProvider, clusterLister, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}
