package fluxinstaller_test

import (
	"os"
	"strings"
	"testing"
	"time"

	fluxinstaller "github.com/devantler-tech/ksail-go/pkg/installer/flux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFluxInstaller(t *testing.T) {
	t.Parallel()

	// Arrange
	kubeconfig := "~/.kube/config"
	context := "test-context"
	timeout := 5 * time.Minute

	// Act
	installer := fluxinstaller.NewFluxInstaller(kubeconfig, context, timeout)

	// Assert
	assert.NotNil(t, installer)
}

func TestFluxInstaller_Install_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := fluxinstaller.NewFluxInstaller(
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

func TestFluxInstaller_Uninstall_Error_InvalidKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := fluxinstaller.NewFluxInstaller(
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

func TestFluxInstaller_Install_Error_EmptyKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := fluxinstaller.NewFluxInstaller(
		"", // empty kubeconfig path
		"test-context",
		5*time.Minute,
	)

	// Act
	err := installer.Install()

	// Assert
	require.Error(t, err)
}

func TestFluxInstaller_Uninstall_Error_EmptyKubeconfig(t *testing.T) {
	t.Parallel()

	// Arrange
	installer := fluxinstaller.NewFluxInstaller(
		"", // empty kubeconfig path
		"test-context",
		5*time.Minute,
	)

	// Act
	err := installer.Uninstall()

	// Assert
	require.Error(t, err)
}

func TestFluxInstaller_Install_Error_MalformedKubeconfig(t *testing.T) {
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
	installer := fluxinstaller.NewFluxInstaller(
		kubeconfigPath,
		"test-context",
		5*time.Minute,
	)

	// Act
	err = installer.Install()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to install Flux operator")
}

func TestFluxInstaller_Uninstall_Error_MalformedKubeconfig(t *testing.T) {
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
	installer := fluxinstaller.NewFluxInstaller(
		kubeconfigPath,
		"test-context",
		5*time.Minute,
	)

	// Act
	err = installer.Uninstall()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to uninstall flux-operator release")
}

func TestFluxInstaller_Install_ValidKubeconfig_ConnectError(t *testing.T) {
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
	installer := fluxinstaller.NewFluxInstaller(
		kubeconfigPath,
		"test-context",
		1*time.Second, // Short timeout for faster test
	)

	// Act
	err = installer.Install()

	// Assert
	require.Error(t, err)
	// Should fail when trying to connect to the Kubernetes API or install Helm chart
	assert.True(t, strings.Contains(err.Error(), "failed to install Flux operator") ||
		strings.Contains(err.Error(), "failed to install or upgrade chart"))
}

func TestFluxInstaller_Uninstall_ValidKubeconfig_ConnectError(t *testing.T) {
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
	installer := fluxinstaller.NewFluxInstaller(
		kubeconfigPath,
		"test-context",
		1*time.Second, // Short timeout for faster test
	)

	// Act
	err = installer.Uninstall()

	// Assert
	require.Error(t, err)
	// Should fail when trying to connect to the Kubernetes API or uninstall Helm release
	assert.True(t, strings.Contains(err.Error(), "failed to uninstall flux-operator release") ||
		strings.Contains(err.Error(), "failed to create Helm client"))
}

func TestFluxInstaller_EmptyContextName(t *testing.T) {
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
	installer := fluxinstaller.NewFluxInstaller(
		kubeconfigPath,
		"", // Empty context
		1*time.Second,
	)

	// Act
	err = installer.Install()

	// Assert
	require.Error(t, err)
	// Should fail when trying to connect to the server or install chart
	assert.True(t, strings.Contains(err.Error(), "failed to install Flux operator") ||
		strings.Contains(err.Error(), "failed to install or upgrade chart"))
}