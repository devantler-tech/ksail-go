package kubectlinstaller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	k8sutils "github.com/devantler-tech/ksail-go/internal/utils/k8s"
	kubectlinstaller "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

// Static errors for testing to satisfy err113 linter.
var (
	errCRDCreationFailed      = errors.New("failed to create CRD")
	errApplySetCreationFailed = errors.New("failed to create ApplySet")
	errAPIServerError         = errors.New("api server error")
	errServerError            = errors.New("server error")
	errGetError               = errors.New("get error")
	errUpdateError            = errors.New("update error")
	errCreateFailed           = errors.New("create failed")
)

// testSetup provides common setup for kubectl installer tests.
func testSetup(
	t *testing.T,
) (*kubectlinstaller.MockCustomResourceDefinitionInterface, *kubectlinstaller.MockResourceInterface) {
	t.Helper()
	apiExtClient := kubectlinstaller.NewMockCustomResourceDefinitionInterface(t)
	dynClient := kubectlinstaller.NewMockResourceInterface(t)

	return apiExtClient, dynClient
}

// createTestInstaller creates a kubectl installer with common test configuration.
func createTestInstaller(
	apiExtClient apiextensionsv1client.CustomResourceDefinitionInterface,
	dynClient dynamic.ResourceInterface,
) *kubectlinstaller.KubectlInstaller {
	return kubectlinstaller.NewKubectlInstaller(
		5*time.Second,
		apiExtClient,
		dynClient,
	)
}

// expectCRDNotFound sets up expectation for CRD not found.
func expectCRDNotFound(apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface) {
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
}

// expectCRDExists sets up expectation for CRD already exists.
func expectCRDExists(apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface) {
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
}

// expectCRDCreateSuccess sets up expectation for successful CRD creation.
func expectCRDCreateSuccess(apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface) {
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
}

// expectCRDCreateAlreadyExists sets up expectation for CRD creation returning AlreadyExists.
func expectCRDCreateAlreadyExists(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
) {
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
}

// expectCRDEstablishmentSuccess sets up expectations for successful CRD establishment polling.
func expectCRDEstablishmentSuccess(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
) {
	establishedCRD := createEstablishedCRD()
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(establishedCRD, nil).
		Maybe()
}

// expectApplySetNotFound sets up expectation for ApplySet not found.
func expectApplySetNotFound(dynClient *kubectlinstaller.MockResourceInterface) {
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
}

// expectApplySetCreateSuccess sets up expectation for successful ApplySet creation.
func expectApplySetCreateSuccess(dynClient *kubectlinstaller.MockResourceInterface) {
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{Object: map[string]any{}}, nil).
		Times(1)
}

// expectApplySetCreateAlreadyExists sets up expectation for ApplySet creation returning AlreadyExists.
func expectApplySetCreateAlreadyExists(dynClient *kubectlinstaller.MockResourceInterface) {
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "ksail")).
		Times(1)
}

// createEstablishedCRD creates a CRD with established condition for testing.
func createEstablishedCRD() *apiextensionsv1.CustomResourceDefinition {
	crd := createDefaultCRD()
	crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:               apiextensionsv1.Established,
			Status:             apiextensionsv1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "EstablishedSuccessfully",
			Message:            "CRD is established",
		},
	}

	return crd
}

// createExistingApplySet creates an existing ApplySet object for testing.
func createExistingApplySet() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]any{
			"metadata": map[string]any{
				"name":            "ksail",
				"resourceVersion": "applyset-version-456",
			},
		},
	}
}

// expectCRDEstablishmentError sets up expectations for CRD establishment that fails with server error.
func expectCRDEstablishmentError(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
) {
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, errServerError).
		Maybe() // Allow multiple calls during polling
}

// createCRDWithNamesNotAccepted creates a CRD with NamesAccepted = false for testing.
func createCRDWithNamesNotAccepted() *apiextensionsv1.CustomResourceDefinition {
	crd := createDefaultCRD()
	crd.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:               apiextensionsv1.NamesAccepted,
			Status:             apiextensionsv1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MultipleNamesNotAllowed",
			Message:            "names conflict with existing CRD",
		},
	}

	return crd
}

// expectCRDEstablishmentWithNamesNotAccepted sets up expectations for CRD establishment with NamesAccepted = false.
func expectCRDEstablishmentWithNamesNotAccepted(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
) {
	crdWithNamesNotAccepted := createCRDWithNamesNotAccepted()
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(crdWithNamesNotAccepted, nil).
		Maybe()
}

// createShortTimeoutInstaller creates an installer with very short timeout for testing failures.
func createShortTimeoutInstaller(
	apiExtClient apiextensionsv1client.CustomResourceDefinitionInterface,
	dynClient dynamic.ResourceInterface,
) *kubectlinstaller.KubectlInstaller {
	return kubectlinstaller.NewKubectlInstaller(
		10*time.Millisecond,
		apiExtClient,
		dynClient,
	)
}

// expectCRDGetForUpdate sets up expectation for getting existing CRD for update.
func expectCRDGetForUpdate(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
) *apiextensionsv1.CustomResourceDefinition {
	existingCRD := createDefaultCRD()
	existingCRD.Name = "applysets.k8s.devantler.tech"
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(existingCRD, nil).
		Times(1)

	return existingCRD
}

// expectCRDUpdateSuccess sets up expectation for successful CRD update.
func expectCRDUpdateSuccess(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
	crd *apiextensionsv1.CustomResourceDefinition,
) {
	apiExtClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(crd, nil).
		Times(1)
}

// expectApplySetGetForUpdate sets up expectation for getting existing ApplySet for update.
func expectApplySetGetForUpdate(
	dynClient *kubectlinstaller.MockResourceInterface,
) *unstructured.Unstructured {
	existingCR := createExistingApplySet()
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(existingCR, nil).
		Times(1)

	return existingCR
}

// expectApplySetUpdateSuccess sets up expectation for successful ApplySet update.
func expectApplySetUpdateSuccess(
	dynClient *kubectlinstaller.MockResourceInterface,
	applyset *unstructured.Unstructured,
) {
	dynClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(applyset, nil).
		Times(1)
}

// expectCRDEstablishmentWithPolling sets up expectations for CRD establishment
// that requires polling (NotFound -> Established).
func expectCRDEstablishmentWithPolling(
	apiExtClient *kubectlinstaller.MockCustomResourceDefinitionInterface,
) {
	// During establishment waiting - first call returns NotFound (should continue polling)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)

	// Second call returns established CRD
	expectCRDEstablishmentSuccess(apiExtClient)
}

// runInstallTestExpectingError runs an install test expecting a specific error message.
func runInstallTestExpectingError(
	t *testing.T,
	installer *kubectlinstaller.KubectlInstaller,
	expectedErrorMessage string,
) {
	t.Helper()

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedErrorMessage)
}

// runInstallTestExpectingSuccess runs an install test expecting success.
func runInstallTestExpectingSuccess(
	t *testing.T,
	installer *kubectlinstaller.KubectlInstaller,
) {
	t.Helper()

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

func TestNewKubectlInstaller(t *testing.T) {
	t.Parallel()

	timeout := 5 * time.Minute
	apiExtClient := kubectlinstaller.NewMockCustomResourceDefinitionInterface(t)
	dynClient := kubectlinstaller.NewMockResourceInterface(t)

	installer := kubectlinstaller.NewKubectlInstaller(timeout, apiExtClient, dynClient)

	assert.NotNil(t, installer)
}

func TestKubectlInstaller_Install_Success(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip creation and establishment)
	expectCRDExists(apiExtClient)

	// Setup ApplySet CR not found and successful creation
	expectApplySetNotFound(dynClient)
	expectApplySetCreateSuccess(dynClient)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingSuccess(t, installer)
}

func TestKubectlInstaller_Install_Error_CRDCreation(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	expectCRDNotFound(apiExtClient)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errCRDCreationFailed)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to create CRD")
}

func TestKubectlInstaller_Install_CRDEstablishmentTimeout(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found and successful creation
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech"))
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(createDefaultCRD(), nil)

	// CRD establishment polling returns not established (times out)
	crdNotEstablished := createDefaultCRD()
	crdNotEstablished.Status = apiextensionsv1.CustomResourceDefinitionStatus{
		Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
			{
				Type:               apiextensionsv1.Established,
				Status:             apiextensionsv1.ConditionFalse,
				LastTransitionTime: metav1.Time{Time: time.Time{}},
				Reason:             "",
				Message:            "",
			},
		},
		AcceptedNames:  createDefaultCRDNames(),
		StoredVersions: []string{},
	}
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(crdNotEstablished, nil).Maybe()

	installer := kubectlinstaller.NewKubectlInstaller(
		1*time.Millisecond, // Very short timeout to fail quickly
		apiExtClient,
		dynClient,
	)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to wait for CRD to be established")
}

func TestKubectlInstaller_Install_ApplySetCRCreateError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists
	expectCRDExists(apiExtClient)

	// ApplySet CR not found and creation fails
	expectApplySetNotFound(dynClient)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errApplySetCreationFailed)

	installer := createTestInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create ApplySet CR")
}

func TestKubectlInstaller_Uninstall_Success(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// Both deletions succeed
	dynClient.EXPECT().Delete(mock.Anything, "ksail", mock.Anything).Return(nil)
	apiExtClient.EXPECT().
		Delete(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil)

	installer := createTestInstaller(apiExtClient, dynClient)

	err := installer.Uninstall(context.Background())

	require.NoError(t, err)
}

func TestKubectlInstaller_Install_CRDGetError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD Get operation fails with non-NotFound error
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, errAPIServerError)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to check CRD existence")
}

func TestKubectlInstaller_Install_ApplySetGetError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists
	expectCRDExists(apiExtClient)

	// ApplySet Get operation fails with non-NotFound error
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, errAPIServerError)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to get ApplySet CR")
}

func TestKubectlInstaller_EmbeddedCRAsset(t *testing.T) {
	t.Parallel()

	// Test that the embedded CR YAML can be unmarshaled properly
	testCRYAML := `
apiVersion: k8s.devantler.tech/v1
kind: ApplySet
metadata:
  name: test-applyset
  annotations:
    applyset.k8s.devantler.tech/contains-group-kinds: "v1/ConfigMap,v1/Secret,v1/Service,apps/v1/Deployment"
spec: {}
`

	var applySetObj unstructured.Unstructured

	err := yaml.Unmarshal([]byte(testCRYAML), &applySetObj.Object)

	require.NoError(t, err)
	assert.Equal(t, "test-applyset", applySetObj.GetName())
	assert.Equal(t, "ApplySet", applySetObj.GetKind())
}

// createDefaultCRD creates a default CustomResourceDefinition for testing.
func createDefaultCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: k8sutils.NewEmptyObjectMeta(),
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Names: createDefaultCRDNames(),
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			AcceptedNames: createDefaultCRDNames(),
		},
	}
}

// createDefaultCRDNames creates a default CustomResourceDefinitionNames for testing.
func createDefaultCRDNames() apiextensionsv1.CustomResourceDefinitionNames {
	return apiextensionsv1.CustomResourceDefinitionNames{}
}

// createDefaultGroupResource creates a default GroupResource for testing.
func createDefaultGroupResource() schema.GroupResource {
	return schema.GroupResource{
		Group:    "",
		Resource: "",
	}
}

// Test to cover the waitForCRDEstablished error path when Get fails.
func TestKubectlInstaller_WaitForCRDEstablished_GetError_Direct(t *testing.T) {
	t.Parallel()

	// Create a simple test that only targets the CRD establishment error path
	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD creation succeeds
	expectCRDCreateSuccess(apiExtClient)

	// During establishment waiting, Get always returns a server error
	expectCRDEstablishmentError(apiExtClient)

	// Use a very short timeout to make the test fast
	installer := createShortTimeoutInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get CRD")
}

// Test to cover the NamesAccepted false condition in waitForCRDEstablished.
func TestKubectlInstaller_WaitForCRDEstablished_NamesNotAccepted_Direct(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD creation succeeds
	expectCRDCreateSuccess(apiExtClient)

	// During establishment waiting, return CRD with NamesAccepted = false
	expectCRDEstablishmentWithNamesNotAccepted(apiExtClient)

	installer := createShortTimeoutInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "crd names not accepted")
	assert.Contains(t, err.Error(), "names conflict with existing CRD")
}

// Test to cover the CRD update path (AlreadyExists -> Update) which triggers createDefaultUpdateOptions.
func TestKubectlInstaller_ApplyCRD_UpdatePath_Success(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD Create returns AlreadyExists (race condition)
	expectCRDCreateAlreadyExists(apiExtClient)

	// Get existing CRD for update
	existingCRD := expectCRDGetForUpdate(apiExtClient)
	existingCRD.ResourceVersion = "test-version-123"

	// Update succeeds (this triggers createDefaultUpdateOptions)
	expectCRDUpdateSuccess(apiExtClient, existingCRD)

	// Establishment check - CRD is already established
	expectCRDEstablishmentSuccess(apiExtClient)

	// ApplySet CR creation
	expectApplySetNotFound(dynClient)
	expectApplySetCreateSuccess(dynClient)

	installer := createTestInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

// Test to cover the ApplySet CR update path.
func TestKubectlInstaller_ApplyApplySetCR_UpdatePath_Success(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	expectCRDExists(apiExtClient)

	// ApplySet CR not found initially
	expectApplySetNotFound(dynClient)

	// ApplySet Create returns AlreadyExists (race condition)
	expectApplySetCreateAlreadyExists(dynClient)

	// Get existing ApplySet for update
	existingCR := expectApplySetGetForUpdate(dynClient)

	// Update succeeds (this also triggers createDefaultUpdateOptions)
	expectApplySetUpdateSuccess(dynClient, existingCR)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingSuccess(t, installer)
}

// Test to cover Get error in CRD update path.
func TestKubectlInstaller_ApplyCRD_GetErrorInUpdate(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD Create returns AlreadyExists (race condition)
	expectCRDCreateAlreadyExists(apiExtClient)

	// Get existing CRD fails
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, errGetError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to get existing CRD for update")
}

// Test to cover Update error in CRD update path.
func TestKubectlInstaller_ApplyCRD_UpdateError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD Create returns AlreadyExists (race condition)
	expectCRDCreateAlreadyExists(apiExtClient)

	// Get existing CRD succeeds
	existingCRD := expectCRDGetForUpdate(apiExtClient)
	existingCRD.ResourceVersion = "test-version-123"

	// Update fails
	apiExtClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errUpdateError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to update CRD")
}

// Test to cover Get error in ApplySet CR update path.
func TestKubectlInstaller_ApplyApplySetCR_GetErrorInUpdate(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	expectCRDExists(apiExtClient)

	// ApplySet CR not found initially
	expectApplySetNotFound(dynClient)

	// ApplySet Create returns AlreadyExists (race condition)
	expectApplySetCreateAlreadyExists(dynClient)

	// Get existing ApplySet fails
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, errGetError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to get existing ApplySet")
}

// Test to cover Update error in ApplySet CR update path.
func TestKubectlInstaller_ApplyApplySetCR_UpdateError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	expectCRDExists(apiExtClient)

	// ApplySet CR not found initially
	expectApplySetNotFound(dynClient)

	// ApplySet Create returns AlreadyExists (race condition)
	expectApplySetCreateAlreadyExists(dynClient)

	// Get existing ApplySet succeeds
	expectApplySetGetForUpdate(dynClient)

	// Update fails
	dynClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errUpdateError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to update ApplySet")
}

// Test to cover the Create failure path that's not AlreadyExists.
func TestKubectlInstaller_ApplyCRD_CreateFailure(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD Create fails with some other error (not AlreadyExists)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errCreateFailed).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to create CRD")
}

// Test to cover the Create failure path in ApplySet CR that's not AlreadyExists.
func TestKubectlInstaller_ApplyApplySetCR_CreateFailure(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	expectCRDExists(apiExtClient)

	// ApplySet CR not found initially
	expectApplySetNotFound(dynClient)

	// ApplySet Create fails with some other error (not AlreadyExists)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errCreateFailed).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	runInstallTestExpectingError(t, installer, "failed to create ApplySet CR")
}

// Test to cover NotFound during CRD establishment polling.
func TestKubectlInstaller_WaitForCRDEstablished_NotFoundDuringPolling(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD creation succeeds
	expectCRDCreateSuccess(apiExtClient)

	// During establishment waiting - polling returns NotFound -> Established
	expectCRDEstablishmentWithPolling(apiExtClient)

	// ApplySet CR creation
	expectApplySetNotFound(dynClient)
	expectApplySetCreateSuccess(dynClient)

	installer := createTestInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

// Test to cover the successful Create path in applyCRD (no AlreadyExists).
func TestKubectlInstaller_ApplyCRD_CreateSuccess_Direct(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially
	expectCRDNotFound(apiExtClient)

	// CRD Create succeeds immediately (no race condition)
	expectCRDCreateSuccess(apiExtClient)

	// Establishment check - CRD is already established
	expectCRDEstablishmentSuccess(apiExtClient)

	// ApplySet CR creation
	expectApplySetNotFound(dynClient)
	expectApplySetCreateSuccess(dynClient)

	installer := createTestInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}

// Test to cover the successful Create path in applyApplySetCR (no AlreadyExists).
func TestKubectlInstaller_ApplyApplySetCR_CreateSuccess_Direct(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	expectCRDExists(apiExtClient)

	// ApplySet CR not found initially
	expectApplySetNotFound(dynClient)

	// ApplySet Create succeeds immediately (no race condition)
	expectApplySetCreateSuccess(dynClient)

	installer := createTestInstaller(apiExtClient, dynClient)

	err := installer.Install(context.Background())

	require.NoError(t, err)
}
