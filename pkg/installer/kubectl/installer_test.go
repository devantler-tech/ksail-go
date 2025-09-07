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

func TestNewKubectlInstaller(t *testing.T) {
	t.Parallel()

	// Arrange
	timeout := 5 * time.Minute
	apiExtClient := kubectlinstaller.NewMockCustomResourceDefinitionInterface(t)
	dynClient := kubectlinstaller.NewMockResourceInterface(t)

	// Act
	installer := kubectlinstaller.NewKubectlInstaller(timeout, apiExtClient, dynClient)

	// Assert
	assert.NotNil(t, installer)
}


func TestKubectlInstaller_Install_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip creation and establishment)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil)

	// Setup ApplySet CR not found and successful creation
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail"))
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{
			Object: map[string]interface{}{},
		}, nil)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.NoError(t, err)
}

func TestKubectlInstaller_Install_Error_CRDCreation(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)
	
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech"))
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errCRDCreationFailed)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create CRD")
}

func TestKubectlInstaller_Install_CRDEstablishmentTimeout(t *testing.T) {
	t.Parallel()

	// Arrange
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

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to wait for CRD to be established")
}

func TestKubectlInstaller_Install_ApplySetCRCreateError(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)

	// CRD already exists
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil)

	// ApplySet CR not found and creation fails
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail"))
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errApplySetCreationFailed)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create ApplySet CR")
}

func TestKubectlInstaller_Uninstall_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)

	// Both deletions succeed
	dynClient.EXPECT().Delete(mock.Anything, "ksail", mock.Anything).Return(nil)
	apiExtClient.EXPECT().Delete(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).Return(nil)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Uninstall(context.Background())

	// Assert
	require.NoError(t, err)
}

func TestKubectlInstaller_Install_CRDGetError(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)

	// CRD Get operation fails with non-NotFound error
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, errAPIServerError)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check CRD existence")
}

func TestKubectlInstaller_Install_ApplySetGetError(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)

	// CRD already exists
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil)

	// ApplySet Get operation fails with non-NotFound error
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, errAPIServerError)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get ApplySet CR")
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
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ObjectMeta: k8sutils.NewEmptyObjectMeta(),
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group:                 "",
			Names:                 createDefaultCRDNames(),
			Scope:                 "",
			Versions:              nil,
			Conversion:            nil,
			PreserveUnknownFields: false,
		},
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions:     nil,
			AcceptedNames:  createDefaultCRDNames(),
			StoredVersions: nil,
		},
	}
}

// createDefaultCRDNames creates a default CustomResourceDefinitionNames for testing.
func createDefaultCRDNames() apiextensionsv1.CustomResourceDefinitionNames {
	return apiextensionsv1.CustomResourceDefinitionNames{
		Plural:     "",
		Singular:   "",
		ShortNames: nil,
		Kind:       "",
		ListKind:   "",
		Categories: nil,
	}
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
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD creation succeeds
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// During establishment waiting, Get always returns a server error
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, errServerError).
		Maybe() // Allow multiple calls during polling

	// Use a very short timeout to make the test fast
	installer := kubectlinstaller.NewKubectlInstaller(
		10*time.Millisecond, // Even shorter
		apiExtClient,
		dynClient,
	)

	// Act
	err := installer.Install(context.Background())

	// Assert 
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get CRD")
}

// Test to cover the NamesAccepted false condition in waitForCRDEstablished.
func TestKubectlInstaller_WaitForCRDEstablished_NamesNotAccepted_Direct(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD creation succeeds
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// During establishment waiting, return CRD with NamesAccepted = false
	crdWithNamesNotAccepted := createDefaultCRD()
	crdWithNamesNotAccepted.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:               apiextensionsv1.NamesAccepted,
			Status:             apiextensionsv1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MultipleNamesNotAllowed",
			Message:            "names conflict with existing CRD",
		},
	}
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(crdWithNamesNotAccepted, nil).
		Maybe()

	installer := kubectlinstaller.NewKubectlInstaller(
		10*time.Millisecond,
		apiExtClient,
		dynClient,
	)

	// Act
	err := installer.Install(context.Background())

	// Assert 
	require.Error(t, err)
	assert.Contains(t, err.Error(), "crd names not accepted")
	assert.Contains(t, err.Error(), "names conflict with existing CRD")
}

// Test to cover the CRD update path (AlreadyExists -> Update) which triggers createDefaultUpdateOptions.
func TestKubectlInstaller_ApplyCRD_UpdatePath_Success(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD Create returns AlreadyExists (race condition)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// Get existing CRD for update
	existingCRD := createDefaultCRD()
	existingCRD.ResourceVersion = "test-version-123"
	existingCRD.Name = "applysets.k8s.devantler.tech"
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(existingCRD, nil).
		Times(1)
	
	// Update succeeds (this triggers createDefaultUpdateOptions)
	apiExtClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(existingCRD, nil).
		Times(1)
	
	// Establishment check - CRD is already established
	establishedCRD := createDefaultCRD()
	establishedCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:               apiextensionsv1.Established,
			Status:             apiextensionsv1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "EstablishedSuccessfully", 
			Message:            "CRD is established",
		},
	}
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(establishedCRD, nil).
		Maybe()
	
	// ApplySet CR creation
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{Object: map[string]interface{}{}}, nil).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.NoError(t, err)
}

// Test to cover the ApplySet CR update path.
func TestKubectlInstaller_ApplyApplySetCR_UpdatePath_Success(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// ApplySet CR not found initially
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// ApplySet Create returns AlreadyExists (race condition)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// Get existing ApplySet for update
	existingCR := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":            "ksail",
				"resourceVersion": "applyset-version-456",
			},
		},
	}
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(existingCR, nil).
		Times(1)
	
	// Update succeeds (this also triggers createDefaultUpdateOptions)
	dynClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(existingCR, nil).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.NoError(t, err)
}

// Test to cover Get error in CRD update path.
func TestKubectlInstaller_ApplyCRD_GetErrorInUpdate(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD Create returns AlreadyExists (race condition)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// Get existing CRD fails
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, errGetError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get existing CRD for update")
}

// Test to cover Update error in CRD update path.
func TestKubectlInstaller_ApplyCRD_UpdateError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD Create returns AlreadyExists (race condition)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// Get existing CRD succeeds
	existingCRD := createDefaultCRD()
	existingCRD.ResourceVersion = "test-version-123"
	existingCRD.Name = "applysets.k8s.devantler.tech"
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(existingCRD, nil).
		Times(1)
	
	// Update fails
	apiExtClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errUpdateError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update CRD")
}

// Test to cover Get error in ApplySet CR update path.
func TestKubectlInstaller_ApplyApplySetCR_GetErrorInUpdate(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// ApplySet CR not found initially
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// ApplySet Create returns AlreadyExists (race condition)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// Get existing ApplySet fails
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, errGetError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get existing ApplySet")
}

// Test to cover Update error in ApplySet CR update path.
func TestKubectlInstaller_ApplyApplySetCR_UpdateError(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// ApplySet CR not found initially
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// ApplySet Create returns AlreadyExists (race condition)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, apierrors.NewAlreadyExists(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// Get existing ApplySet succeeds
	existingCR := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":            "ksail",
				"resourceVersion": "applyset-version-456",
			},
		},
	}
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(existingCR, nil).
		Times(1)
	
	// Update fails
	dynClient.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errUpdateError).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update ApplySet")
}

// Test to cover the Create failure path that's not AlreadyExists.
func TestKubectlInstaller_ApplyCRD_CreateFailure(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD Create fails with some other error (not AlreadyExists)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errCreateFailed).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create CRD")
}

// Test to cover the Create failure path in ApplySet CR that's not AlreadyExists.
func TestKubectlInstaller_ApplyApplySetCR_CreateFailure(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// ApplySet CR not found initially
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// ApplySet Create fails with some other error (not AlreadyExists)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errCreateFailed).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create ApplySet CR")
}

// Test to cover NotFound during CRD establishment polling.
func TestKubectlInstaller_WaitForCRDEstablished_NotFoundDuringPolling(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD creation succeeds
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// During establishment waiting - first call returns NotFound (should continue polling)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// Second call returns established CRD 
	establishedCRD := createDefaultCRD()
	establishedCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:               apiextensionsv1.Established,
			Status:             apiextensionsv1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "EstablishedSuccessfully",
			Message:            "CRD is established",
		},
	}
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(establishedCRD, nil).
		Maybe()
	
	// ApplySet CR creation
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{Object: map[string]interface{}{}}, nil).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.NoError(t, err)
}

// Test to cover the successful Create path in applyCRD (no AlreadyExists).
func TestKubectlInstaller_ApplyCRD_CreateSuccess_Direct(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD not found initially 
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "applysets.k8s.devantler.tech")).
		Times(1)
	
	// CRD Create succeeds immediately (no race condition)
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// Establishment check - CRD is already established
	establishedCRD := createDefaultCRD()
	establishedCRD.Status.Conditions = []apiextensionsv1.CustomResourceDefinitionCondition{
		{
			Type:               apiextensionsv1.Established,
			Status:             apiextensionsv1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "EstablishedSuccessfully",
			Message:            "CRD is established",
		},
	}
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(establishedCRD, nil).
		Maybe()
	
	// ApplySet CR creation
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{Object: map[string]interface{}{}}, nil).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.NoError(t, err)
}

// Test to cover the successful Create path in applyApplySetCR (no AlreadyExists).
func TestKubectlInstaller_ApplyApplySetCR_CreateSuccess_Direct(t *testing.T) {
	t.Parallel()

	apiExtClient, dynClient := testSetup(t)

	// CRD already exists (skip CRD logic)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(createDefaultCRD(), nil).
		Times(1)
	
	// ApplySet CR not found initially
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(createDefaultGroupResource(), "ksail")).
		Times(1)
	
	// ApplySet Create succeeds immediately (no race condition)
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{Object: map[string]interface{}{}}, nil).
		Times(1)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install(context.Background())

	// Assert
	require.NoError(t, err)
}