package eksprovisioner_test

import (
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

var errBoom = errors.New("boom")

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		runActionSuccess(
			t,
			"Create()",
			nameCase.InputName,
			nameCase.ExpectedName,
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

	provisioner, _, _, _, clusterCreator, _ := newProvisionerForTest(t)

	clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(errBoom)

	err := provisioner.Create("test-cluster")

	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to create EKS cluster", "Create()")
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	cases := []testutils.NameCase{
		{Name: "without name uses cfg", InputName: "", ExpectedName: "cfg-name"},
		{Name: "with name", InputName: "custom", ExpectedName: "custom"},
	}

	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		runDeleteActionSuccess(
			t,
			"Delete()",
			nameCase.InputName,
			nameCase.ExpectedName,
			func(_ *eks.ClusterProvider,
				clusterActions *eksprovisioner.MockEKSClusterActions, _ string) {
				// No longer need to mock provider construction since it's injected directly

				// Delete(ctx context.Context, waitInterval, podEvictionWaitPeriod time.Duration,
				//   wait, force, disableNodegroupEviction bool, parallel int)
				clusterActions.On(
					"Delete",
					mock.Anything,                         // ctx context.Context
					mock.AnythingOfType("time.Duration"),  // waitInterval time.Duration
					mock.AnythingOfType("time.Duration"),  // podEvictionWaitPeriod time.Duration
					mock.AnythingOfType("bool"),           // wait bool
					mock.AnythingOfType("bool"),           // force bool
					mock.AnythingOfType("bool"),           // disableNodegroupEviction bool
					mock.AnythingOfType("int"),            // parallel int
				).Return(nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Delete(name)
			},
		)
	})
}

func TestDelete_Error_CreateFailed(t *testing.T) {
	t.Parallel()

	provisioner, _, clusterActions, _, _, _ := newProvisionerForTest(t)

	clusterActions.On(
		"Delete",
		mock.Anything,                         // ctx context.Context
		mock.AnythingOfType("time.Duration"),  // waitInterval time.Duration
		mock.AnythingOfType("time.Duration"),  // podEvictionWaitPeriod time.Duration
		mock.AnythingOfType("bool"),           // wait bool
		mock.AnythingOfType("bool"),           // force bool
		mock.AnythingOfType("bool"),           // disableNodegroupEviction bool
		mock.AnythingOfType("int"),            // parallel int
	).Return(errBoom)

	err := provisioner.Delete("test-cluster")

	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to delete EKS cluster", "Delete()")
}

func TestStart_Success(t *testing.T) {
	t.Parallel()

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		provisioner, _, _, clusterLister, _, nodeGroupManager := newProvisionerForTest(t)
		descriptions := []cluster.Description{{Name: nameCase.ExpectedName}}
		clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)

		// Mock node group manager
		nodeGroupManager.On("Scale", mock.Anything, mock.Anything, true).Return(nil)

		err := provisioner.Start(nameCase.InputName)
		if err != nil {
			t.Fatalf("Start() unexpected error: %v", err)
		}
	})
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()

	provisioner, _, _, clusterLister, _, _ := newProvisionerForTest(t)

	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]cluster.Description{}, nil)

	err := provisioner.Start("test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrClusterNotFound)
}

func TestStop_Success(t *testing.T) {
	t.Parallel()

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		provisioner, _, _, clusterLister, _, nodeGroupManager := newProvisionerForTest(t)
		descriptions := []cluster.Description{{Name: nameCase.ExpectedName}}
		clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)

		// Mock node group manager
		nodeGroupManager.On("Scale", mock.Anything, mock.Anything, true).Return(nil)

		err := provisioner.Stop(nameCase.InputName)
		if err != nil {
			t.Fatalf("Stop() unexpected error: %v", err)
		}
	})
}

func TestList_Success(t *testing.T) {
	t.Parallel()

	provisioner, _, _, clusterLister, _, _ := newProvisionerForTest(t)
	descriptions := []cluster.Description{
		{Name: "cluster1"},
		{Name: "cluster2"},
	}
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)

	clusters, err := provisioner.List()

	require.NoError(t, err)
	assert.Equal(t, []string{"cluster1", "cluster2"}, clusters)
}

func TestList_Error_GetClustersFailed(t *testing.T) {
	t.Parallel()

	provisioner, _, _, clusterLister, _, _ := newProvisionerForTest(t)
	
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errBoom)
	
	clusters, err := provisioner.List()

	assert.Nil(t, clusters)
	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to list EKS clusters", "List()")
}

func TestExists_Success_True(t *testing.T) {
	t.Parallel()

	provisioner, _, _, clusterLister, _, _ := newProvisionerForTest(t)
	descriptions := []cluster.Description{{Name: "cfg-name"}}
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)

	exists, err := provisioner.Exists("cfg-name")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()

	provisioner, _, _, clusterLister, _, _ := newProvisionerForTest(t)

	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]cluster.Description{}, nil)

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
	provisioner, clusterProvider, _, _, clusterCreator, _ := newProvisionerForTest(t)
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
	provisioner, clusterProvider, clusterActions, _, _, _ := newProvisionerForTest(t)
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
	provisioner, clusterProvider, _, clusterLister, _, _ := newProvisionerForTest(t)
	expect(clusterProvider, clusterLister, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}
