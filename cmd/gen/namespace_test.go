package gen //nolint:testpackage // Tests need access to unexported helpers

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) {
	setupTestKubeconfig()
	exitCode := m.Run()
	cleanupTestKubeconfig()

	// Clean snapshots
	_, err := snaps.Clean(m, snaps.CleanOpts{Sort: true})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to clean snapshots: " + err.Error() + "\n")
		os.Exit(1)
	}

	os.Exit(exitCode)
}

func setupTestKubeconfig() {
	homeDir, _ := os.UserHomeDir()
	kubeDir := filepath.Join(homeDir, ".kube")
	_ = os.MkdirAll(kubeDir, 0o750)

	kubeconfigContent := `apiVersion: v1
kind: Config
current-context: test
contexts:
- name: test
  context:
    cluster: test
    user: test
clusters:
- name: test
  cluster:
    server: https://127.0.0.1:6443
users:
- name: test
  user: {}
`
	_ = os.WriteFile(filepath.Join(kubeDir, "config"), []byte(kubeconfigContent), 0o600)
}

func cleanupTestKubeconfig() {
	homeDir, _ := os.UserHomeDir()
	kubeconfig := filepath.Join(homeDir, ".kube", "config")
	_ = os.Remove(kubeconfig)
}

func newTestRuntime() *runtime.Runtime {
	return runtime.NewRuntime()
}

// TestGenNamespace tests generating a namespace manifest.
//
//nolint:paralleltest // Snapshot tests should not run in parallel
func TestGenNamespace(t *testing.T) {
	rt := newTestRuntime()
	cmd := NewNamespaceCmd(rt)
	buffer := &bytes.Buffer{}
	cmd.SetOut(buffer)
	cmd.SetErr(buffer)
	cmd.SetArgs([]string{"test-namespace"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected gen namespace to succeed, got %v", err)
	}

	snaps.MatchSnapshot(t, buffer.String())
}
