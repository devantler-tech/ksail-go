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
			func(providerConstructor *eksprovisioner.MockEKSProviderConstructor, clusterCreator *eksprovisioner.MockEKSClusterCreator, name string) {
				providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
				clusterCreator.On("CreateCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Create(name)
			},
		)
	})
}

func TestCreate_Error_ProviderFailed(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, _, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(nil, errBoom)

	err := provisioner.Create("test-cluster")

	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to create EKS provider", "Create()")
}

func TestCreate_Error_CreateFailed(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, _, clusterCreator := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
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
			func(providerConstructor *eksprovisioner.MockEKSProviderConstructor, clusterActionsFactory *eksprovisioner.MockEKSClusterActionsFactory, name string) {
				providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
				
				mockClusterActions := eksprovisioner.NewMockEKSClusterActions(t)
				clusterActionsFactory.On("NewClusterActions", mock.Anything, mock.Anything, mock.Anything).Return(mockClusterActions, nil)
				mockClusterActions.On("Delete", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Delete(name)
			},
		)
	})
}

func TestDelete_Error_ProviderFailed(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, _, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(nil, errBoom)

	err := provisioner.Delete("test-cluster")

	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to create EKS provider", "Delete()")
}

func TestStart_Success(t *testing.T) {
	t.Parallel()

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		runListActionSuccess(
			t,
			"Start()",
			nameCase.InputName,
			nameCase.ExpectedName,
			func(providerConstructor *eksprovisioner.MockEKSProviderConstructor, clusterLister *eksprovisioner.MockEKSClusterLister, name string) {
				providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
				descriptions := []cluster.Description{{Name: name}}
				clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Start(name)
			},
		)
	})
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, clusterLister, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]cluster.Description{}, nil)

	err := provisioner.Start("test-cluster")

	assert.ErrorIs(t, err, eksprovisioner.ErrClusterNotFound)
}

func TestStop_Success(t *testing.T) {
	t.Parallel()

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		runListActionSuccess(
			t,
			"Stop()",
			nameCase.InputName,
			nameCase.ExpectedName,
			func(providerConstructor *eksprovisioner.MockEKSProviderConstructor, clusterLister *eksprovisioner.MockEKSClusterLister, name string) {
				providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
				descriptions := []cluster.Description{{Name: name}}
				clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)
			},
			func(prov *eksprovisioner.EKSClusterProvisioner, name string) error {
				return prov.Stop(name)
			},
		)
	})
}

func TestList_Success(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, clusterLister, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
	descriptions := []cluster.Description{
		{Name: "cluster1"},
		{Name: "cluster2"},
	}
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)

	clusters, err := provisioner.List()

	require.NoError(t, err)
	assert.Equal(t, []string{"cluster1", "cluster2"}, clusters)
}

func TestList_Error_ProviderFailed(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, _, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(nil, errBoom)

	clusters, err := provisioner.List()

	assert.Nil(t, clusters)
	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to create EKS provider", "List()")
}

func TestExists_Success_True(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, clusterLister, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
	descriptions := []cluster.Description{{Name: "cfg-name"}}
	clusterLister.On("GetClusters", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(descriptions, nil)

	exists, err := provisioner.Exists("cfg-name")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()

	provisioner, providerConstructor, _, clusterLister, _ := newProvisionerForTest(t)

	providerConstructor.On("NewClusterProvider", mock.Anything, mock.Anything, mock.Anything).Return(&eks.ClusterProvider{}, nil)
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
	*eksprovisioner.MockEKSProviderConstructor,
	*eksprovisioner.MockEKSClusterActionsFactory,
	*eksprovisioner.MockEKSClusterLister,
	*eksprovisioner.MockEKSClusterCreator,
) {
	t.Helper()

	clusterConfig := &v1alpha5.ClusterConfig{
		Metadata: &v1alpha5.ClusterMeta{
			Name:   "cfg-name",
			Region: "us-west-2",
		},
	}

	providerConstructor := eksprovisioner.NewMockEKSProviderConstructor(t)
	clusterActionsFactory := eksprovisioner.NewMockEKSClusterActionsFactory(t)
	clusterLister := eksprovisioner.NewMockEKSClusterLister(t)
	clusterCreator := eksprovisioner.NewMockEKSClusterCreator(t)

	provisioner := eksprovisioner.NewEKSClusterProvisioner(
		clusterConfig,
		providerConstructor,
		clusterActionsFactory,
		clusterLister,
		clusterCreator,
	)

	return provisioner, providerConstructor, clusterActionsFactory, clusterLister, clusterCreator
}

type expectProviderFn func(*eksprovisioner.MockEKSProviderConstructor, *eksprovisioner.MockEKSClusterCreator, string)
type actionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectProviderFn,
	action actionFn,
) {
	t.Helper()
	provisioner, providerConstructor, _, _, clusterCreator := newProvisionerForTest(t)
	expect(providerConstructor, clusterCreator, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

type expectDeleteProviderFn func(*eksprovisioner.MockEKSProviderConstructor, *eksprovisioner.MockEKSClusterActionsFactory, string)
type deleteActionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runDeleteActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectDeleteProviderFn,
	action deleteActionFn,
) {
	t.Helper()
	provisioner, providerConstructor, clusterActionsFactory, _, _ := newProvisionerForTest(t)
	expect(providerConstructor, clusterActionsFactory, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

type expectListProviderFn func(*eksprovisioner.MockEKSProviderConstructor, *eksprovisioner.MockEKSClusterLister, string)
type listActionFn func(*eksprovisioner.EKSClusterProvisioner, string) error

func runListActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectListProviderFn,
	action listActionFn,
) {
	t.Helper()
	provisioner, providerConstructor, _, clusterLister, _ := newProvisionerForTest(t)
	expect(providerConstructor, clusterLister, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}