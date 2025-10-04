package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"context"
	"regexp"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/mock"
)

// TestHandleStopRunE exercises the success and validation error paths for the stop command.

func TestHandleStopRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "stop")
		testutils.SeedValidClusterConfig(manager)

		// Use mock provisioner that doesn't require Docker
		mockProvisioner := &mockClusterProvisioner{}
		mockProvisioner.On("Stop", mock.Anything, "kind").Return(nil)

		factory := func(
			_ context.Context,
			_ v1alpha1.Distribution,
			_ string,
			_ string,
		) (clusterprovisioner.ClusterProvisioner, string, error) {
			return mockProvisioner, "kind", nil
		}

		err := handleStopRunEWithProvisioner(cmd, manager, factory)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Strip timing information from output before snapshot comparison
		// Replace timing values with * to preserve structure
		outputStr := output.String()
		timingRegex := regexp.MustCompile(
			`\[(?:stage:\s*)?\d+(?:\.\d+)?(?:µs|ms|s|m|h)(?:\s*\|\s*total:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h))?\]`,
		)
		sanitizedOutput := timingRegex.ReplaceAllString(outputStr, "[*]")

		// Capture the output as a snapshot
		snaps.MatchSnapshot(t, sanitizedOutput)

		mockProvisioner.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		testutils.RunValidationErrorTest(t, "stop", HandleStopRunE)
	})
}
