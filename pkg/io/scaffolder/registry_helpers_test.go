package scaffolder_test

import (
	"io"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/scaffolder"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createTestScaffolderForKind creates a test scaffolder for Kind distribution.
func createTestScaffolderForKind() *scaffolder.Scaffolder {
	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksail.io/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionKind,
		},
	}

	return scaffolder.NewScaffolder(*cluster, io.Discard)
}

// createTestScaffolderForK3d creates a test scaffolder for K3d distribution.
func createTestScaffolderForK3d() *scaffolder.Scaffolder {
	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "ksail.io/v1alpha1",
			Kind:       "Cluster",
		},
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionK3d,
		},
	}

	return scaffolder.NewScaffolder(*cluster, io.Discard)
}

// TestGenerateContainerdPatches tests the generation of containerd patches for Kind.
func TestGenerateContainerdPatches(t *testing.T) {
	t.Parallel()

	t.Run("single mirror registry", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForKind()
		scaf.MirrorRegistries = []string{"docker-io=https://registry-1.docker.io"}

		patches := scaf.GenerateContainerdPatches()
		require.Len(t, patches, 1)
		require.Contains(t, patches[0], "docker.io")
		require.Contains(t, patches[0], "kind-docker-io")
		require.Contains(t, patches[0], "plugins.\"io.containerd.grpc.v1.cri\".registry.mirrors")
	})

	t.Run("multiple mirror registries", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForKind()
		scaf.MirrorRegistries = []string{
			"docker-io=https://registry-1.docker.io",
			"ghcr-io=https://ghcr.io",
			"quay-io=https://quay.io",
		}

		patches := scaf.GenerateContainerdPatches()
		require.Len(t, patches, 3)
		require.Contains(t, patches[0], "docker.io")
		require.Contains(t, patches[1], "ghcr.io")
		require.Contains(t, patches[2], "quay.io")
	})

	t.Run("no mirror registries", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForKind()
		scaf.MirrorRegistries = []string{}

		patches := scaf.GenerateContainerdPatches()
		require.Empty(t, patches)
	})

	testContainerdPatchesInvalidAndCustomPort(t)
}

func testContainerdPatchesInvalidAndCustomPort(t *testing.T) {
	t.Helper()

	t.Run("invalid mirror spec skipped", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForKind()
		scaf.MirrorRegistries = []string{
			"docker-io=https://registry-1.docker.io",
			"invalid-spec-no-equals",
			"ghcr-io=https://ghcr.io",
		}

		patches := scaf.GenerateContainerdPatches()
		require.Len(t, patches, 2)
		require.Contains(t, patches[0], "docker.io")
		require.Contains(t, patches[1], "ghcr.io")
	})

	t.Run("custom port in upstream URL", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForKind()
		scaf.MirrorRegistries = []string{"localhost=http://localhost:5001"}

		patches := scaf.GenerateContainerdPatches()
		require.Len(t, patches, 1)
		require.Contains(t, patches[0], "localhost")
		require.Contains(t, patches[0], "kind-localhost:5001")
	})
}

// TestGenerateK3dRegistryConfig tests the generation of K3d registry configuration.
func TestGenerateK3dRegistryConfig(t *testing.T) {
	t.Parallel()

	t.Run("single mirror registry", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForK3d()
		scaf.MirrorRegistries = []string{"docker-io=https://registry-1.docker.io"}

		registryConfig := scaf.GenerateK3dRegistryConfig()
		require.NotNil(t, registryConfig)
		require.NotNil(t, registryConfig.Create)
		require.Equal(t, "k3d-docker-io", registryConfig.Create.Name)
		require.NotEmpty(t, registryConfig.Config)
		require.Contains(t, registryConfig.Config, "mirrors:")
		require.Contains(t, registryConfig.Config, "docker.io")
		require.Contains(t, registryConfig.Config, "k3d-docker-io:5000")
	})

	t.Run("no mirror registries", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForK3d()
		scaf.MirrorRegistries = []string{}

		registryConfig := scaf.GenerateK3dRegistryConfig()
		require.Empty(t, registryConfig.Use)
		require.Nil(t, registryConfig.Create)
		require.Empty(t, registryConfig.Config)
	})

	t.Run("invalid mirror spec", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForK3d()
		scaf.MirrorRegistries = []string{"invalid-no-equals"}

		registryConfig := scaf.GenerateK3dRegistryConfig()
		require.Empty(t, registryConfig.Use)
		require.Nil(t, registryConfig.Create)
		require.Empty(t, registryConfig.Config)
	})

	t.Run("first mirror used when multiple", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForK3d()
		scaf.MirrorRegistries = []string{
			"docker-io=https://registry-1.docker.io",
			"ghcr-io=https://ghcr.io",
		}

		registryConfig := scaf.GenerateK3dRegistryConfig()
		require.NotNil(t, registryConfig.Create)
		require.Equal(t, "k3d-docker-io", registryConfig.Create.Name)
		require.Contains(t, registryConfig.Config, "docker.io")
		// Currently only first mirror is used in K3d config
		require.NotContains(t, registryConfig.Config, "ghcr.io")
	})
}
