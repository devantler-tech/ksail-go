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

var (
	errProvisionerFactory = errors.New("failed factory")
	errProvisionFailed    = errors.New("provision failed")
)

func newProvisionerFactory(
	provisioner clusterprovisioner.ClusterProvisioner,
	distribution string,
	factoryErr error,
) func(context.Context, v1alpha1.Distribution, string, string) (
	clusterprovisioner.ClusterProvisioner,
	string,
	error,
) {
	return func(
		_ context.Context,
		_ v1alpha1.Distribution,
		_ string,
		_ string,
	) (clusterprovisioner.ClusterProvisioner, string, error) {
		return provisioner, distribution, factoryErr
	}
}

// TestHandleUpRunE exercises success and error paths for the up command handler.
//
//nolint:paralleltest,tparallel // validation subtest uses t.Chdir
func TestHandleUpRunE(t *testing.T) {
	t.Run("success", testHandleUpRunESuccess)
	t.Run("validation error", testHandleUpRunEValidation) //nolint:paralleltest // uses t.Chdir
	t.Run("provisioner creation failure", testHandleUpRunEProvisionerCreationFailure)
	t.Run("provision failure", testHandleUpRunEProvisionFailure)
	t.Run(
		"default factory unsupported distribution",
		testHandleUpRunEDefaultFactoryUnsupportedDistribution,
	)
}

func testHandleUpRunESuccess(t *testing.T) {
	t.Helper()
	t.Parallel()

	cmd, manager, output := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

	mockProvisioner := clusterprovisioner.NewMockClusterProvisioner(t)
	mockProvisioner.On("Create", mock.Anything, "kind").Return(nil)

	factory := newProvisionerFactory(mockProvisioner, "kind", nil)

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
	t.Parallel()

	cmd, manager, _ := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

	factory := newProvisionerFactory(nil, "", errProvisionerFactory)

	err := handleUpRunEWithProvisioner(cmd, manager, factory)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to create provisioner")
	require.ErrorIs(t, err, errProvisionerFactory)
}

func testHandleUpRunEProvisionFailure(t *testing.T) {
	t.Helper()
	t.Parallel()

	cmd, manager, _ := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

	mockProvisioner := clusterprovisioner.NewMockClusterProvisioner(t)
	mockProvisioner.On("Create", mock.Anything, "kind").Return(errProvisionFailed)

	factory := newProvisionerFactory(mockProvisioner, "kind", nil)

	err := handleUpRunEWithProvisioner(cmd, manager, factory)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to provision cluster")
	require.ErrorIs(t, err, errProvisionFailed)

	mockProvisioner.AssertExpectations(t)
}

func testHandleUpRunEDefaultFactoryUnsupportedDistribution(t *testing.T) {
	t.Helper()
	t.Parallel()

	cmd, manager, _ := testutils.NewCommandAndManager(t, "up")
	testutils.SeedValidClusterConfig(manager)

	_, err := manager.LoadConfig(nil)
	require.NoError(t, err)

	manager.Config.Spec.Distribution = v1alpha1.Distribution("unsupported")

	err = handleUpRunEWithProvisioner(cmd, manager, nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to create provisioner")
	require.ErrorContains(t, err, "unsupported distribution")
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
