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

// CreateMalformedKubeconfigFile creates a temporary file with malformed YAML for testing.
func CreateMalformedKubeconfigFile(t *testing.T) string {
	t.Helper()

	malformedKubeconfig := `
this is not valid yaml: [
`

	tmpDir, err := os.MkdirTemp(os.TempDir(), "ksail-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	
	kubeconfigPath := tmpDir + "/kubeconfig"
	err = os.WriteFile(kubeconfigPath, []byte(malformedKubeconfig), DefaultFilePermissions)
	require.NoError(t, err)

	return kubeconfigPath
}

// CreateEmptyKubeconfigFile creates a temporary empty kubeconfig file for testing.
func CreateEmptyKubeconfigFile(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "ksail-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	
	kubeconfigPath := tmpDir + "/kubeconfig"
	err = os.WriteFile(kubeconfigPath, []byte(""), DefaultFilePermissions)
	require.NoError(t, err)

	return kubeconfigPath
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

	tmpDir, err := os.MkdirTemp(os.TempDir(), "ksail-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) })
	
	kubeconfigPath := tmpDir + "/kubeconfig"
	err = os.WriteFile(kubeconfigPath, []byte(validKubeconfig), DefaultFilePermissions)
	require.NoError(t, err)

	return kubeconfigPath
}
