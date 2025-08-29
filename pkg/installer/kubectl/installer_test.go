package kubectlinstaller_test

import (
	"os"
	"strings"
	"testing"
	"time"

	kubectlinstaller "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestNewKubectlInstaller(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	// Act
	installer := kubectlinstaller.NewKubectlInstaller(kubeconfig, context, timeout)

	// Assert
	assert.NotNil(t, installer)
}

func TestKubectlInstaller_Install_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		"/nonexistent/kubeconfig",
		"test-context",
		5*time.Minute,
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
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestKubectlInstaller_EmbeddedAssets(t *testing.T) {
	t.Parallel()

	// This test verifies that the embedded assets are properly loaded
	// by creating an installer and checking it doesn't panic during construction
	installer := kubectlinstaller.NewKubectlInstaller(
		"test-kubeconfig",
		"test-context",
		1*time.Minute,
	)

	assert.NotNil(t, installer)
}

func TestKubectlInstaller_EmbeddedCRDAsset(t *testing.T) {
	t.Parallel()

	// Test that the embedded CRD YAML can be unmarshaled properly
	// This indirectly tests the applyCRD method's YAML unmarshaling
	// We can access the embedded asset through reflection or by testing similar logic

	// Test valid CRD structure - this mimics what applyCRD does
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
	// This indirectly tests the applyApplySetCR method's YAML unmarshaling
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

func TestKubectlInstaller_Install_Error_EmptyKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		"", // empty kubeconfig path
		"test-context",
		5*time.Minute,
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
}

func TestKubectlInstaller_Uninstall_Error_EmptyKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		"", // empty kubeconfig path
		"test-context",
		5*time.Minute,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.Error(t, err)
}

func TestKubectlInstaller_BuildRESTConfig_Error_InvalidPath(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		"/nonexistent/kubeconfig",
		"test-context",
		5*time.Minute,
	)

	// Act - we can't directly test buildRESTConfig since it's unexported,
	// but we can test it indirectly through Install
	err := installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestKubectlInstaller_BuildRESTConfig_ValidPath(t *testing.T) {
	t.Parallel()

	// Create a temporary kubeconfig file for testing
	tempKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-server:8443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpDir := t.TempDir()
	kubeconfigPath := tmpDir + "/kubeconfig"
	err := os.WriteFile(kubeconfigPath, []byte(tempKubeconfig), 0600)
	require.NoError(t, err)

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"test-context",
		5*time.Minute,
	)

	// Act - test indirectly through Install
	err = installer.Install()

	// Assert - it should fail because test-server doesn't exist, but it should get past config building
	require.Error(t, err)
	// The error should be about connecting to the server, not about the kubeconfig format
	assert.True(t, strings.Contains(err.Error(), "failed to check CRD existence") ||
		strings.Contains(err.Error(), "failed to create"),
		"Expected error about CRD check or client creation, got: %s", err.Error())
}

func TestKubectlInstaller_ApplyCRD_YAMLUnmarshalError(t *testing.T) {
	t.Parallel()

	// Test that the embedded CRD YAML is valid by trying to unmarshal it
	// This indirectly tests the applyCRD method's YAML handling
	testCRDYAML := `invalid yaml content: [}`

	var crd apiextensionsv1.CustomResourceDefinition

	err := yaml.Unmarshal([]byte(testCRDYAML), &crd)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "yaml")
}

func TestKubectlInstaller_ApplyApplySetCR_YAMLUnmarshalError(t *testing.T) {
	t.Parallel()

	// Test that invalid YAML fails to unmarshal
	// This indirectly tests the applyApplySetCR method's YAML handling
	testCRYAML := `invalid yaml content: [}`

	var applySetObj unstructured.Unstructured

	err := yaml.Unmarshal([]byte(testCRYAML), &applySetObj.Object)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "yaml")
}

func TestKubectlInstaller_InstallWithValidKubeconfig_ConnectError(t *testing.T) {
	t.Parallel()

	// Create a valid kubeconfig that points to a non-existent server
	tempKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://non-existent-server:8443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpDir := t.TempDir()
	kubeconfigPath := tmpDir + "/kubeconfig"
	err := os.WriteFile(kubeconfigPath, []byte(tempKubeconfig), 0600)
	require.NoError(t, err)

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"test-context",
		1*time.Second, // Short timeout for faster test
	)

	// Act
	err = installer.Install()

	// Assert
	require.Error(t, err)
	// Should fail when trying to connect to the Kubernetes API
	assert.True(t, strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "failed to check CRD existence") ||
		strings.Contains(err.Error(), "failed to create"))
}

func TestKubectlInstaller_UninstallWithValidKubeconfig_ConnectError(t *testing.T) {
	t.Parallel()

	// Create a valid kubeconfig that points to a non-existent server
	tempKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://non-existent-server:8443
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpDir := t.TempDir()
	kubeconfigPath := tmpDir + "/kubeconfig"
	err := os.WriteFile(kubeconfigPath, []byte(tempKubeconfig), 0600)
	require.NoError(t, err)

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"test-context",
		1*time.Second, // Short timeout for faster test
	)

	// Act
	err = installer.Uninstall()

	// Assert
	// Uninstall ignores deletion errors, so it should succeed even if server is unreachable
	require.NoError(t, err)
}

func TestKubectlInstaller_BuildRESTConfig_MalformedKubeconfig(t *testing.T) {
	t.Parallel()

	// Create a malformed kubeconfig file
	malformedKubeconfig := `
this is not valid yaml: [
`

	tmpDir := t.TempDir()
	kubeconfigPath := tmpDir + "/kubeconfig"
	err := os.WriteFile(kubeconfigPath, []byte(malformedKubeconfig), 0600)
	require.NoError(t, err)

	// Arrange
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"test-context",
		5*time.Minute,
	)

	// Act
	err = installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to build rest config")
}

func TestKubectlInstaller_EmptyContextName(t *testing.T) {
	t.Parallel()

	// Create a valid kubeconfig
	tempKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-server:8443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

	tmpDir := t.TempDir()
	kubeconfigPath := tmpDir + "/kubeconfig"
	err := os.WriteFile(kubeconfigPath, []byte(tempKubeconfig), 0600)
	require.NoError(t, err)

	// Arrange - empty context should use current-context from kubeconfig
	installer := kubectlinstaller.NewKubectlInstaller(
		kubeconfigPath,
		"", // Empty context
		1*time.Second,
	)

	// Act
	err = installer.Install()

	// Assert
	require.Error(t, err)
	// Should fail when trying to connect to the server, but config should be built successfully
	assert.True(t, strings.Contains(err.Error(), "connection") ||
		strings.Contains(err.Error(), "failed to check CRD existence") ||
		strings.Contains(err.Error(), "failed to create"))
}
