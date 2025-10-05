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
