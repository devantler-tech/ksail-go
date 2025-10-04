package cluster //nolint:testpackage // Access internal helpers without exporting them.

import (
	"bytes"
	"context"
	"regexp"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/stretchr/testify/mock"
)

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

// sanitizeTimingOutput replaces timing values with placeholders for snapshot testing.
func sanitizeTimingOutput(output *bytes.Buffer) string {
	outputStr := output.String()
	timingRegex := regexp.MustCompile(
		`\[(?:stage:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h)(?:\s*\|\s*total:\s*\d+(?:\.\d+)?(?:µs|ms|s|m|h))?)\]`,
	)

	return timingRegex.ReplaceAllStringFunc(outputStr, func(match string) string {
		// Check if it's a multi-stage timing (contains |total:)
		if regexp.MustCompile(`\|`).MatchString(match) {
			return "[stage: *|total: *]"
		}

		return "[stage: *]"
	})
}

// createProvisionerFactory creates a provisioner factory function for testing.
// This helper reduces duplication when creating factory functions in tests.
func createProvisionerFactory(
	provisioner clusterprovisioner.ClusterProvisioner,
) provisionerFactory {
	return func(
		_ context.Context,
		_ v1alpha1.Distribution,
		_ string,
		_ string,
	) (clusterprovisioner.ClusterProvisioner, string, error) {
		return provisioner, "kind", nil
	}
}
