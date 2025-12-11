package testutils

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

// Suppress unused warnings for shared utilities that may not be used in all test files.
var (
	_ = context.Background
	_ = time.Second
)

const (
	// FilePermUserReadWrite is the permission for user read/write only (0o600).
	FilePermUserReadWrite fs.FileMode = 0o600
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
	return strings.Contains(s, substr)
}

// UpdateDeploymentStatusToUnready modifies a deployment payload to mark it as unready.
func UpdateDeploymentStatusToUnready(t *testing.T, payload map[string]any) {
	t.Helper()

	status := getStatusMap(t, payload)
	status["updatedReplicas"] = 0
	status["availableReplicas"] = 0
}

// UpdateDaemonSetStatusToUnready modifies a daemonset payload to mark it as unready.
func UpdateDaemonSetStatusToUnready(t *testing.T, payload map[string]any) {
	t.Helper()

	status := getStatusMap(t, payload)
	status["numberUnavailable"] = 1
	status["updatedNumberScheduled"] = 0
}

func getStatusMap(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()

	status, ok := payload["status"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected payload status type %T", payload["status"])
	}

	return status
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

	err := os.WriteFile(path, []byte(content), FilePermUserReadWrite)
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

	err := os.WriteFile(path, []byte(contents), FilePermUserReadWrite)
	if err != nil {
		t.Fatalf("write kubeconfig file: %v", err)
	}

	return path
}

// CreateReadyDaemonSetClient creates a fake clientset with a ready DaemonSet.
//

func CreateReadyDaemonSetClient(namespace, name string) kubernetes.Interface {
	return fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberUnavailable:      0,
			UpdatedNumberScheduled: 1,
		},
	})
}

// CreateDaemonSetClientWithAPIError creates a fake clientset that returns an error on DaemonSet get.
//

func CreateDaemonSetClientWithAPIError(err error) kubernetes.Interface {
	client := fake.NewSimpleClientset()
	client.PrependReactor(
		"get",
		"daemonsets",
		func(_ k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, err
		},
	)

	return client
}

// CreateUnreadyDaemonSetClient creates a fake clientset with an unready DaemonSet.
//

func CreateUnreadyDaemonSetClient(namespace, name string) kubernetes.Interface {
	return fake.NewSimpleClientset(&appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DaemonSetStatus{
			DesiredNumberScheduled: 1,
			NumberUnavailable:      1,
			UpdatedNumberScheduled: 0,
		},
	})
}

// CreateReadyDeploymentClient creates a fake clientset with a ready Deployment.
//

func CreateReadyDeploymentClient(namespace, name string) kubernetes.Interface {
	return fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DeploymentStatus{
			Replicas:          1,
			UpdatedReplicas:   1,
			AvailableReplicas: 1,
		},
	})
}

// CreateDeploymentClientWithAPIError creates a fake clientset that returns an error on Deployment get.
//

func CreateDeploymentClientWithAPIError(err error) kubernetes.Interface {
	client := fake.NewSimpleClientset()
	client.PrependReactor(
		"get",
		"deployments",
		func(_ k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, err
		},
	)

	return client
}

// CreateUnreadyDeploymentClient creates a fake clientset with an unready Deployment.
//

func CreateUnreadyDeploymentClient(namespace, name string) kubernetes.Interface {
	return fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DeploymentStatus{
			Replicas:          1,
			UpdatedReplicas:   0,
			AvailableReplicas: 0,
		},
	})
}
