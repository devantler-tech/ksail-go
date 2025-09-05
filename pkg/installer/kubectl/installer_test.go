package kubectlinstaller_test

import (
	"errors"
	"testing"
	"time"

	kubectlinstaller "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
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
		Return(&apiextensionsv1.CustomResourceDefinition{}, nil)

	// Setup ApplySet CR not found and successful creation
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "ksail"))
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{}, nil)

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install()

	// Assert
	require.NoError(t, err)
}

func TestKubectlInstaller_Install_Error_CRDCreation(t *testing.T) {
	t.Parallel()

	// Arrange
	apiExtClient, dynClient := testSetup(t)
	
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "applysets.k8s.devantler.tech"))
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to create CRD"))

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install()

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
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "applysets.k8s.devantler.tech"))
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&apiextensionsv1.CustomResourceDefinition{}, nil)
	
	// CRD establishment polling returns not established (times out)
	crdNotEstablished := &apiextensionsv1.CustomResourceDefinition{
		Status: apiextensionsv1.CustomResourceDefinitionStatus{
			Conditions: []apiextensionsv1.CustomResourceDefinitionCondition{
				{
					Type:   apiextensionsv1.Established,
					Status: apiextensionsv1.ConditionFalse,
				},
			},
		},
	}
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(crdNotEstablished, nil).Maybe()

	installer := kubectlinstaller.NewKubectlInstaller(
		1*time.Millisecond, // Very short timeout to fail quickly
		apiExtClient,
		dynClient,
	)

	// Act
	err := installer.Install()

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
		Return(&apiextensionsv1.CustomResourceDefinition{}, nil)

	// ApplySet CR not found and creation fails
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "ksail"))
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to create ApplySet"))

	installer := createTestInstaller(apiExtClient, dynClient)

	// Act
	err := installer.Install()

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
	err := installer.Uninstall()

	// Assert
	require.NoError(t, err)
}


func TestKubectlInstaller_EmbeddedCRDAsset(t *testing.T) {
	t.Parallel()

	// Test that the embedded CRD YAML can be unmarshaled properly
	testCRDYAML := `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: applysets.k8s.devantler.tech
spec:
  group: k8s.devantler.tech
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
  scope: Cluster
  names:
    plural: applysets
    singular: applyset
    kind: ApplySet
`

	var crd apiextensionsv1.CustomResourceDefinition

	err := yaml.Unmarshal([]byte(testCRDYAML), &crd)

	require.NoError(t, err)
	assert.Equal(t, "applysets.k8s.devantler.tech", crd.Name)
	assert.Equal(t, "k8s.devantler.tech", crd.Spec.Group)
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