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

// TestHandleUpRunE exercises success and validation error paths.
//
//nolint:dupl // Intentional duplication with down_test - similar test structure for lifecycle operations
func TestHandleUpRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "up")
		testutils.SeedValidClusterConfig(manager)

		// Use mock provisioner that doesn't require Docker
		mockProvisioner := &mockClusterProvisioner{}
		mockProvisioner.On("Create", mock.Anything, "kind").Return(nil)

		factory := func(
			_ context.Context,
			_ v1alpha1.Distribution,
			_ string,
			_ string,
		) (clusterprovisioner.ClusterProvisioner, string, error) {
			return mockProvisioner, "kind", nil
		}

		err := handleUpRunEWithProvisioner(cmd, manager, factory)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Strip timing information from output before snapshot comparison
		// Replace timing values with * to preserve structure: [stage: *] or [stage: *|total: *]
		outputStr := output.String()
		timingRegex := regexp.MustCompile(
			`\[(?:stage:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h)(?:\s*\|\s*total:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h))?)\]`,
		)
		sanitizedOutput := timingRegex.ReplaceAllStringFunc(outputStr, func(match string) string {
			// Check if it's a multi-stage timing (contains |total:)
			if regexp.MustCompile(`\|`).MatchString(match) {
				return "[stage: *|total: *]"
			}

			return "[stage: *]"
		})

		// Capture the output as a snapshot
		snaps.MatchSnapshot(t, sanitizedOutput)

		mockProvisioner.AssertExpectations(t)
	})

	t.Run("validation error", func(t *testing.T) { //nolint:paralleltest // uses t.Chdir
		testutils.RunValidationErrorTest(t, "up", HandleUpRunE)
	})
}

// mockClusterProvisioner is a test mock using testify/mock.
type mockClusterProvisioner struct {
	mock.Mock
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisioner) Create(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisioner) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisioner) Start(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisioner) Stop(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisioner) List(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]string), args.Error(1) //nolint:forcetypeassert // Test mock
}

func (m *mockClusterProvisioner) Exists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)

	return args.Bool(0), args.Error(1)
}
