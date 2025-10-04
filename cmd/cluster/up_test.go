package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cluster/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestHandleUpRunE exercises success and error paths for the up command handler.
func TestHandleUpRunE(t *testing.T) { //nolint:paralleltest
	t.Run("success", testHandleUpRunESuccess)             //nolint:paralleltest
	t.Run("validation error", testHandleUpRunEValidation) //nolint:paralleltest // uses t.Chdir
	t.Run("provisioner creation failure", testHandleUpRunEProvisionerCreationFailure)
	t.Run("provision failure", testHandleUpRunEProvisionFailure)
}

func testHandleUpRunESuccess(t *testing.T) {
	t.Helper()

	cmd, manager, output := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

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
	require.NoError(t, err)

	sanitizedOutput := sanitizeTimingOutput(output.String())
	snaps.MatchSnapshot(t, sanitizedOutput)

	mockProvisioner.AssertExpectations(t)
}

func testHandleUpRunEValidation(t *testing.T) {
	t.Helper()

	testutils.RunValidationErrorTest(t, "up", HandleUpRunE)
}

func testHandleUpRunEProvisionerCreationFailure(t *testing.T) {
	t.Helper()

	cmd, manager, _ := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

	expectedErr := errors.New("failed factory")

	factory := func(
		_ context.Context,
		_ v1alpha1.Distribution,
		_ string,
		_ string,
	) (clusterprovisioner.ClusterProvisioner, string, error) {
		return nil, "", expectedErr
	}

	err := handleUpRunEWithProvisioner(cmd, manager, factory)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to create provisioner")
	require.ErrorIs(t, err, expectedErr)
}

func testHandleUpRunEProvisionFailure(t *testing.T) {
	t.Helper()

	cmd, manager, _ := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

	provisionErr := errors.New("provision failed")

	mockProvisioner := &mockClusterProvisioner{}
	mockProvisioner.On("Create", mock.Anything, "kind").Return(provisionErr)

	factory := func(
		_ context.Context,
		_ v1alpha1.Distribution,
		_ string,
		_ string,
	) (clusterprovisioner.ClusterProvisioner, string, error) {
		return mockProvisioner, "kind", nil
	}

	err := handleUpRunEWithProvisioner(cmd, manager, factory)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to provision cluster")
	require.ErrorIs(t, err, provisionErr)

	mockProvisioner.AssertExpectations(t)
}

var timingRegex = regexp.MustCompile(
	`\[(?:stage:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h)(?:\s*\|\s*total:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h))?)\]`,
)

func sanitizeTimingOutput(output string) string {
	return timingRegex.ReplaceAllStringFunc(output, func(match string) string {
		if strings.Contains(match, "|") {
			return "[stage: *|total: *]"
		}

		return "[stage: *]"
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
