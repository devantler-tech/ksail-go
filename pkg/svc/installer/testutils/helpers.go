// Package testutils provides shared testing utilities for CNI installer tests.
package testutils

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// Common test errors used across installer tests.
var (
	ErrInstallFailed   = errors.New("install failed")
	ErrAddRepoFailed   = errors.New("add repo failed")
	ErrUninstallFailed = errors.New("uninstall failed")
	ErrDaemonSetBoom   = errors.New("boom")
	ErrDeploymentFail  = errors.New("fail")
	ErrPollBoom        = errors.New("boom")
)

// ExpectEqual is a generic test helper that compares two comparable values.
func ExpectEqual[T comparable](t *testing.T, got, want T, description string) {
	t.Helper()

	if got != want {
		t.Fatalf("unexpected %s: got %v want %v", description, got, want)
	}
}

// ExpectInstallerResult checks if an error matches the expected result.
func ExpectInstallerResult(t *testing.T, err error, wantErr, operation string) {
	t.Helper()

	if wantErr == "" {
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", operation, err)
		}

		return
	}

	if err == nil {
		t.Fatalf("%s: expected error containing %q, got nil", operation, wantErr)
	}

	if !contains(err.Error(), wantErr) {
		t.Fatalf("%s: expected error containing %q, got %q", operation, wantErr, err.Error())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// UpdateDeploymentStatusToUnready modifies a deployment payload to mark it as unready.
func UpdateDeploymentStatusToUnready(t *testing.T, payload map[string]any) {
	t.Helper()

	status, ok := payload["status"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected payload status type %T", payload["status"])
	}

	status["updatedReplicas"] = 0
	status["availableReplicas"] = 0
}

// UpdateDaemonSetStatusToUnready modifies a daemonset payload to mark it as unready.
func UpdateDaemonSetStatusToUnready(t *testing.T, payload map[string]any) {
	t.Helper()

	status, ok := payload["status"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected payload status type %T", payload["status"])
	}

	status["numberUnavailable"] = 1
	status["updatedNumberScheduled"] = 0
}

// EncodeJSON encodes a payload as JSON and writes it to an HTTP response.
func EncodeJSON(t *testing.T, writer http.ResponseWriter, payload any) {
	t.Helper()

	writer.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(writer)

	err := encoder.Encode(payload)
	if err != nil {
		t.Fatalf("failed to encode response: %v", err)
	}
}

// WriteServerBackedKubeconfig creates a minimal kubeconfig file for testing.
func WriteServerBackedKubeconfig(t *testing.T, serverURL string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "kubeconfig.yaml")

	content := "apiVersion: v1\n" +
		"clusters:\n" +
		"- cluster:\n" +
		"    server: " + serverURL + "\n" +
		"    insecure-skip-tls-verify: true\n" +
		"  name: local\n" +
		"contexts:\n" +
		"- context:\n" +
		"    cluster: local\n" +
		"    user: default\n" +
		"  name: default\n" +
		"current-context: default\n" +
		"kind: Config\n" +
		"preferences: {}\n" +
		"users:\n" +
		"- name: default\n" +
		"  user: {}\n"

	err := os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("failed to write kubeconfig: %v", err)
	}

	return path
}

// WriteKubeconfig creates a test kubeconfig with multiple contexts.
func WriteKubeconfig(t *testing.T, dir string) string {
	t.Helper()

	contents := `apiVersion: v1
kind: Config
clusters:
- name: cluster-one
  cluster:
    server: https://cluster-one.example.com
- name: cluster-two
  cluster:
    server: https://cluster-two.example.com
contexts:
- name: primary
  context:
    cluster: cluster-one
    user: user-one
- name: alt
  context:
    cluster: cluster-two
    user: user-two
current-context: primary
users:
- name: user-one
  user:
    token: token-one
- name: user-two
  user:
    token: token-two
`

	path := filepath.Join(dir, "config")

	err := os.WriteFile(path, []byte(contents), 0o600)
	if err != nil {
		t.Fatalf("write kubeconfig file: %v", err)
	}

	return path
}
