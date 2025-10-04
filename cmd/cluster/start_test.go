//nolint:testpackage,dupl // Access internal helpers. Test structure similar to up_test.go
package cluster

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

// TestHandleStartRunE exercises the success and validation error paths for the start command.

func TestHandleStartRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", func(t *testing.T) { //nolint:paralleltest
		cmd, manager, output := testutils.NewCommandAndManager(t, "start")
		testutils.SeedValidClusterConfig(manager)

		// Use mock provisioner that doesn't require Docker
		mockProvisioner := &mockClusterProvisionerForStart{}
		mockProvisioner.On("Start", mock.Anything, "kind").Return(nil)

		factory := func(
			_ context.Context,
			_ v1alpha1.Distribution,
			_ string,
			_ string,
		) (clusterprovisioner.ClusterProvisioner, string, error) {
			return mockProvisioner, "kind", nil
		}

		err := handleStartRunEWithProvisioner(cmd, manager, factory)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// Strip timing information from output before snapshot comparison
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
		testutils.RunValidationErrorTest(t, "start", HandleStartRunE)
	})
}

// mockClusterProvisionerForStart is a test mock using testify/mock.
type mockClusterProvisionerForStart struct {
	mock.Mock
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisionerForStart) Start(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisionerForStart) Create(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisionerForStart) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisionerForStart) Stop(ctx context.Context, name string) error {
	args := m.Called(ctx, name)

	return args.Error(0)
}

//nolint:wrapcheck // Test mock returning error from mock framework
func (m *mockClusterProvisionerForStart) List(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]string), args.Error(1) //nolint:forcetypeassert // Test mock
}

func (m *mockClusterProvisionerForStart) Exists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)

	return args.Bool(0), args.Error(1)
}
