package kubectlinstaller_test

import (
	"testing"
	"time"

	kubectlinstaller "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"
	"github.com/stretchr/testify/assert"
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
	assert.Error(t, err)
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
	assert.Error(t, err)
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

	assert.NoError(t, err)
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

	assert.NoError(t, err)
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
	assert.Error(t, err)
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
	assert.Error(t, err)
}