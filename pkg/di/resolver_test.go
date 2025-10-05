package di_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/di"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResolver(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		clusterFactory func(*testing.T) *v1alpha1.Cluster
		wantErr        error
	}{
		"nil cluster returns error": {
			clusterFactory: func(*testing.T) *v1alpha1.Cluster {
				return nil
			},
			wantErr: di.ErrClusterConfigRequired,
		},
		"valid cluster returns resolver": {
			clusterFactory: func(t *testing.T) *v1alpha1.Cluster {
				t.Helper()

				return newK3dCluster(t)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cluster := tc.clusterFactory(t)

			resolver, err := di.NewResolver(cluster)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, resolver)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolver)
		})
	}
}

func TestResolverResolve(t *testing.T) {
	t.Parallel()

	t.Run("returns dependencies for k3d distribution", func(t *testing.T) {
		t.Parallel()

		resolver := newResolver(t, newK3dCluster(t))

		deps, err := resolver.Resolve()

		require.NoError(t, err)
		require.NotNil(t, deps)

		assert.IsType(t, &k3dprovisioner.K3dClusterProvisioner{}, deps.Provisioner)

		_, ok := deps.DistributionConfig.(*v1alpha5.SimpleConfig)
		assert.True(t, ok, "expected distribution config to be a *v1alpha5.SimpleConfig")
	})

	t.Run("returns error when distribution is unsupported", func(t *testing.T) {
		t.Parallel()

		resolver := newResolver(t, &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Distribution: v1alpha1.Distribution("Unsupported"),
			},
		})

		_, err := resolver.Resolve()

		require.Error(t, err)
		require.ErrorContains(t, err, "create cluster provisioner")
		require.ErrorContains(t, err, "unsupported distribution")
	})
}

func newResolver(t *testing.T, cluster *v1alpha1.Cluster) *di.Resolver {
	t.Helper()

	resolver, err := di.NewResolver(cluster)
	require.NoError(t, err)

	return resolver
}

func newK3dCluster(t *testing.T) *v1alpha1.Cluster {
	t.Helper()

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "k3d.yaml")

	return &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionK3d,
			DistributionConfig: configPath,
			Connection: v1alpha1.Connection{
				Kubeconfig: filepath.Join(configDir, "kubeconfig"),
			},
		},
	}
}
