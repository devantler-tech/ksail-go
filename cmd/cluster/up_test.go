package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
)

// TestHandleUpRunE exercises success and validation error paths.

func TestHandleUpRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "up")
		testutils.SeedValidClusterConfig(manager)

		// Use mock provisioner factory that doesn't require Docker
		mockProvisioner := &mockClusterProvisioner{}
		factory := func(_ context.Context, _ *v1alpha1.Cluster) (clusterprovisioner.ClusterProvisioner, error) {
			return mockProvisioner, nil
		}

		err := handleUpRunEWithProvisioner(cmd, manager, factory)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Verify the output contains expected messages
		outputStr := output.String()
		assertOutputContains(t, outputStr, "🚀 Provisioning cluster...")
		assertOutputContains(t, outputStr, "provisioned cluster successfully")
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		testutils.RunValidationErrorTest(t, "up", HandleUpRunE)
	})
}

// mockClusterProvisioner is a test mock that doesn't require Docker.
type mockClusterProvisioner struct{}

func (m *mockClusterProvisioner) Create(_ context.Context, _ string) error {
	return nil
}

func (m *mockClusterProvisioner) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockClusterProvisioner) Start(_ context.Context, _ string) error {
	return nil
}

func (m *mockClusterProvisioner) Stop(_ context.Context, _ string) error {
	return nil
}

func (m *mockClusterProvisioner) List(_ context.Context) ([]string, error) {
	return []string{}, nil
}

func (m *mockClusterProvisioner) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
