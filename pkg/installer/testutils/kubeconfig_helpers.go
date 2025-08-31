// Package testutils provides common test utilities for installer packages
package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// DefaultFilePermissions defines the default permissions for temporary test files.
	DefaultFilePermissions = 0600
)

// createKubeconfigFileWithContent creates a temporary kubeconfig file with the given content.
func createKubeconfigFileWithContent(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()

	kubeconfigPath := tmpDir + "/kubeconfig"
	err := os.WriteFile(kubeconfigPath, []byte(content), DefaultFilePermissions)
	require.NoError(t, err)

	return kubeconfigPath
}

// CreateMalformedKubeconfigFile creates a temporary file with malformed YAML for testing.
func CreateMalformedKubeconfigFile(t *testing.T) string {
	t.Helper()

	malformedKubeconfig := `
this is not valid yaml: [
`

	return createKubeconfigFileWithContent(t, malformedKubeconfig)
}

// CreateEmptyKubeconfigFile creates a temporary empty kubeconfig file for testing.
func CreateEmptyKubeconfigFile(t *testing.T) string {
	t.Helper()

	return createKubeconfigFileWithContent(t, "")
}

// CreateValidKubeconfigFile creates a temporary valid kubeconfig file for testing.
func CreateValidKubeconfigFile(t *testing.T) string {
	t.Helper()

	validKubeconfig := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://nonexistent-server:6443
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

	return createKubeconfigFileWithContent(t, validKubeconfig)
}
