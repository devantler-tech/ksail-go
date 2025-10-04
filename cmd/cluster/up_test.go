package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/mock"
)

// TestHandleUpRunE exercises success and validation error paths.
func TestHandleUpRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "up")
		testutils.SeedValidClusterConfig(manager)

		// Use mock provisioner that doesn't require Docker
		mockProvisioner := &mockClusterProvisioner{}
		mockProvisioner.On("Create", mock.Anything, "kind").Return(nil)

		factory := createProvisionerFactory(mockProvisioner)

		err := handleUpRunEWithProvisioner(cmd, manager, factory)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Strip timing information from output before snapshot comparison
		sanitizedOutput := sanitizeTimingOutput(output)

		// Capture the output as a snapshot
		snaps.MatchSnapshot(t, sanitizedOutput)

		mockProvisioner.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		testutils.RunValidationErrorTest(t, "up", HandleUpRunE)
	})
}
