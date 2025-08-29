package kubectlinstaller_test

import (
	"testing"
	"time"

	kubectlinstaller "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"
	"github.com/stretchr/testify/assert"
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