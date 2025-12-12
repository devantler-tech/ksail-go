package testutils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// HTTP mock response helpers.

// ServeDeployment serves a mock Kubernetes Deployment resource response.
// This is used in test API servers to simulate deployment readiness checks.
func ServeDeployment(t *testing.T, writer http.ResponseWriter, ready bool) {
	t.Helper()

	payload := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"status": map[string]any{
			"replicas":          1,
			"updatedReplicas":   1,
			"availableReplicas": 1,
		},
	}

	if !ready {
		UpdateDeploymentStatusToUnready(t, payload)
	}

	EncodeJSON(t, writer, payload)
}

// ServeDaemonSet serves a mock Kubernetes DaemonSet resource response.
// This is used in test API servers to simulate daemonset readiness checks.
func ServeDaemonSet(t *testing.T, writer http.ResponseWriter, ready bool) {
	t.Helper()

	payload := map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "DaemonSet",
		"status": map[string]any{
			"desiredNumberScheduled": 1,
			"numberUnavailable":      0,
			"updatedNumberScheduled": 1,
		},
	}

	if !ready {
		UpdateDaemonSetStatusToUnready(t, payload)
	}

	EncodeJSON(t, writer, payload)
}

// Test server helpers.

// NewTestAPIServer creates a test HTTP server with a custom handler function.
// This eliminates boilerplate for creating httptest servers in CNI installer tests.
func NewTestAPIServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	return httptest.NewServer(handler)
}
