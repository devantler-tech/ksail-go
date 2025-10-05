package clusterprovisioner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectation func(*testing.T, clusterprovisioner.ClusterProvisioner, string, error)

type pathProvider func(*testing.T) string

type createClusterProvisionerCase struct {
	name           string
	distribution   v1alpha1.Distribution
	configProvider pathProvider
	assertion      expectation
}

func TestCreateClusterProvisioner(t *testing.T) {
	t.Parallel()

	for _, testCase := range buildCreateClusterProvisionerCases(t) {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			configPath := testCase.configProvider(t)

			provisioner, distributionConfig, err := clusterprovisioner.CreateClusterProvisioner(
				context.Background(),
				testCase.distribution,
				configPath,
				"",
			)

			var clusterName string
			if err == nil {
				clusterName, err = configmanager.GetClusterName(distributionConfig)
				if err != nil {
					t.Fatalf("failed to get cluster name from config: %v", err)
				}
			}

			testCase.assertion(t, provisioner, clusterName, err)
		})
	}
}

func buildCreateClusterProvisionerCases(t *testing.T) []createClusterProvisionerCase {
	t.Helper()

	kindConfig := "kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\nname: custom-kind\n"
	invalidKindConfig := ": invalid\n"
	k3dConfig := "apiVersion: k3d.io/v1alpha5\nkind: Simple\nmetadata:\n  name: custom-k3d\n"

	return []createClusterProvisionerCase{
		{
			name:           "kind default config uses fallback name",
			distribution:   v1alpha1.DistributionKind,
			configProvider: staticPath("non-existent-kind.yaml"),
			assertion:      expectSuccess("kind", &kindprovisioner.KindClusterProvisioner{}),
		},
		{
			name:           "kind config file returns custom name",
			distribution:   v1alpha1.DistributionKind,
			configProvider: tempConfig("kind.yaml", kindConfig),
			assertion:      expectSuccess("custom-kind", &kindprovisioner.KindClusterProvisioner{}),
		},
		{
			name:           "kind invalid config returns load error",
			distribution:   v1alpha1.DistributionKind,
			configProvider: tempConfig("kind-invalid.yaml", invalidKindConfig),
			assertion:      expectErrorContains("failed to load Kind configuration"),
		},
		{
			name:           "k3d default config uses fallback name",
			distribution:   v1alpha1.DistributionK3d,
			configProvider: staticPath("non-existent-k3d.yaml"),
			assertion:      expectSuccess("k3d-default", &k3dprovisioner.K3dClusterProvisioner{}),
		},
		{
			name:           "k3d config file returns custom name",
			distribution:   v1alpha1.DistributionK3d,
			configProvider: tempConfig("k3d.yaml", k3dConfig),
			assertion:      expectSuccess("custom-k3d", &k3dprovisioner.K3dClusterProvisioner{}),
		},
		{
			name:           "unsupported distribution returns error",
			distribution:   v1alpha1.Distribution("unknown"),
			configProvider: staticPath("ignored.yaml"),
			assertion:      expectErrorIs(clusterprovisioner.ErrUnsupportedDistribution),
		},
	}
}

func tempConfig(filename, content string) pathProvider {
	return func(t *testing.T) string {
		t.Helper()

		return createConfigFile(t, filename, content)
	}
}

func staticPath(path string) pathProvider {
	return func(t *testing.T) string {
		t.Helper()

		return path
	}
}

func expectSuccess(expectedName string, expectedType any) expectation {
	return func(t *testing.T, provisioner clusterprovisioner.ClusterProvisioner, clusterName string, err error) {
		t.Helper()

		require.NoError(t, err)
		require.NotNil(t, provisioner)
		assert.Equal(t, expectedName, clusterName)
		assert.IsType(t, expectedType, provisioner)
	}
}

func expectFailure(assertion func(*testing.T, error)) expectation {
	return func(t *testing.T, provisioner clusterprovisioner.ClusterProvisioner, clusterName string, err error) {
		t.Helper()

		require.Error(t, err)
		assert.Nil(t, provisioner)
		assert.Empty(t, clusterName)

		assertion(t, err)
	}
}

func expectErrorContains(message string) expectation {
	return expectFailure(func(t *testing.T, err error) {
		t.Helper()

		assert.ErrorContains(t, err, message)
	})
}

func expectErrorIs(target error) expectation {
	return expectFailure(func(t *testing.T, err error) {
		t.Helper()

		assert.ErrorIs(t, err, target)
	})
}

func createConfigFile(t *testing.T, filename, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, filename)

	err := os.WriteFile(path, []byte(content), 0o600)
	require.NoError(t, err, "writing config fixture should succeed")

	return path
}

func TestCreateKindProvisionerDockerClientError(t *testing.T) {
	t.Helper()

	t.Setenv("DOCKER_HOST", "://")
	t.Setenv("DOCKER_TLS_VERIFY", "")
	t.Setenv("DOCKER_CERT_PATH", "")

	configPath := createConfigFile(
		t,
		"kind.yaml",
		"kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\nname: custom-kind\n",
	)

	provisioner, clusterName, err := clusterprovisioner.CreateClusterProvisioner(
		context.Background(),
		v1alpha1.DistributionKind,
		configPath,
		"",
	)

	require.Error(t, err)
	assert.Nil(t, provisioner)
	assert.Empty(t, clusterName)
	assert.Contains(t, err.Error(), "failed to create Docker client")
}

func TestCreateK3dProvisionerInvalidConfig(t *testing.T) {
	t.Helper()
	t.Parallel()

	configPath := createConfigFile(t, "k3d-invalid.yaml", ": invalid\n")

	provisioner, clusterName, err := clusterprovisioner.CreateClusterProvisioner(
		context.Background(),
		v1alpha1.DistributionK3d,
		configPath,
		"",
	)

	require.Error(t, err)
	assert.Nil(t, provisioner)
	assert.Empty(t, clusterName)
	assert.Contains(t, err.Error(), "failed to load K3d configuration")
}
