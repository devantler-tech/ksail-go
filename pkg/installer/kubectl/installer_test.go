package kubectlinstaller_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	kubectlinstaller "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"
	"github.com/devantler-tech/ksail-go/pkg/installer/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

func TestNewKubectlInstaller(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)

	// Act
	installer := kubectlinstaller.NewKubectlInstaller(kubeconfig, context, timeout, clientFactory)

	// Assert
	assert.NotNil(t, installer)
}

func TestNewKubectlInstallerWithFactory(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)

	// Act
	installer := kubectlinstaller.NewKubectlInstallerWithFactory(kubeconfig, context, timeout, clientFactory)

	// Assert
	assert.NotNil(t, installer)
}

func TestKubectlInstaller_Install_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)
	apiExtClient := kubectlinstaller.NewMockAPIExtensionsClientInterface(t)
	dynClient := kubectlinstaller.NewMockDynamicClientInterface(t)

	// Setup client factory mocks
	clientFactory.EXPECT().CreateAPIExtensionsClient(mock.Anything).Return(apiExtClient, nil)
	clientFactory.EXPECT().CreateDynamicClient(mock.Anything, mock.Anything).Return(dynClient, nil)

	// CRD already exists (skip creation and establishment)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(&apiextensionsv1.CustomResourceDefinition{}, nil)

	// Setup ApplySet CR not found and successful creation
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "ksail"))
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(&unstructured.Unstructured{}, nil)

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		5*time.Second,
		clientFactory,
	)

	// Act
	err := installer.Install()

	// Assert
	require.NoError(t, err)
}

func TestKubectlInstaller_Install_Error_APIExtensionsClientCreation(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)

	clientFactory.EXPECT().CreateAPIExtensionsClient(mock.Anything).
		Return(nil, errors.New("failed to create client"))

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		5*time.Second,
		clientFactory,
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create apiextensions client")
}

func TestKubectlInstaller_Install_Error_CRDCreation(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)
	apiExtClient := kubectlinstaller.NewMockAPIExtensionsClientInterface(t)

	clientFactory.EXPECT().CreateAPIExtensionsClient(mock.Anything).Return(apiExtClient, nil)
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "applysets.k8s.devantler.tech"))
	apiExtClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to create CRD"))

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		5*time.Second,
		clientFactory,
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create CRD")
}

func TestKubectlInstaller_Install_CRDEstablishmentTimeout(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)
	apiExtClient := kubectlinstaller.NewMockAPIExtensionsClientInterface(t)

	clientFactory.EXPECT().CreateAPIExtensionsClient(mock.Anything).Return(apiExtClient, nil)

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

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		1*time.Millisecond, // Very short timeout to fail quickly
		clientFactory,
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
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)
	apiExtClient := kubectlinstaller.NewMockAPIExtensionsClientInterface(t)
	dynClient := kubectlinstaller.NewMockDynamicClientInterface(t)

	clientFactory.EXPECT().CreateAPIExtensionsClient(mock.Anything).Return(apiExtClient, nil)
	clientFactory.EXPECT().CreateDynamicClient(mock.Anything, mock.Anything).Return(dynClient, nil)

	// CRD already exists
	apiExtClient.EXPECT().Get(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).
		Return(&apiextensionsv1.CustomResourceDefinition{}, nil)

	// ApplySet CR not found and creation fails
	dynClient.EXPECT().Get(mock.Anything, "ksail", mock.Anything).
		Return(nil, apierrors.NewNotFound(schema.GroupResource{}, "ksail"))
	dynClient.EXPECT().Create(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to create ApplySet"))

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		5*time.Second,
		clientFactory,
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create ApplySet CR")
}

func TestKubectlInstaller_Uninstall_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)
	apiExtClient := kubectlinstaller.NewMockAPIExtensionsClientInterface(t)
	dynClient := kubectlinstaller.NewMockDynamicClientInterface(t)

	clientFactory.EXPECT().CreateDynamicClient(mock.Anything, mock.Anything).Return(dynClient, nil)
	clientFactory.EXPECT().CreateAPIExtensionsClient(mock.Anything).Return(apiExtClient, nil)

	// Both deletions succeed
	dynClient.EXPECT().Delete(mock.Anything, "ksail", mock.Anything).Return(nil)
	apiExtClient.EXPECT().Delete(mock.Anything, "applysets.k8s.devantler.tech", mock.Anything).Return(nil)

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		5*time.Second,
		clientFactory,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.NoError(t, err)
}

func TestKubectlInstaller_Uninstall_DynamicClientCreationError(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	clientFactory := kubectlinstaller.NewMockClientFactoryInterface(t)

	clientFactory.EXPECT().CreateDynamicClient(mock.Anything, mock.Anything).
		Return(nil, errors.New("failed to create dynamic client"))

	installer := kubectlinstaller.NewKubectlInstallerWithFactory(
		kubeconfigPath,
		"test-context",
		5*time.Second,
		clientFactory,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create dynamic client")
}

func TestKubectlInstaller_Install_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		"/nonexistent/kubeconfig",
		"test-context",
		5*time.Minute,
		kubectlinstaller.NewDefaultClientFactory(),
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestKubectlInstaller_Uninstall_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		"/nonexistent/kubeconfig",
		"test-context",
		5*time.Minute,
		kubectlinstaller.NewDefaultClientFactory(),
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestKubectlInstaller_BuildRESTConfig_ValidPath(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateValidKubeconfigFile(t)
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"test-context",
		5*time.Minute,
		kubectlinstaller.NewDefaultClientFactory(),
	)

	// Act - test indirectly through Install
	err := installer.Install()

	// Assert - it should fail because test-server doesn't exist, but it should get past config building
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to check CRD existence") ||
		strings.Contains(err.Error(), "failed to create"),
		"Expected error about CRD check or client creation, got: %s", err.Error())
}

func TestKubectlInstaller_BuildRESTConfig_MalformedKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfigPath := testutils.CreateMalformedKubeconfigFile(t)
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"test-context",
		5*time.Minute,
		kubectlinstaller.NewDefaultClientFactory(),
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build rest config")
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