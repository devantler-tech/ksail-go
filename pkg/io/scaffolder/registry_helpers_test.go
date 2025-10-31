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

	cases := []struct {
		name    string
		mirrors []string
		assert  func(t *testing.T, patches []string)
	}{
		{
			name:    "single mirror registry",
			mirrors: []string{"docker.io=https://registry-1.docker.io"},
			assert: func(t *testing.T, patches []string) {
				require.Len(t, patches, 1)
				require.Contains(t, patches[0], "docker.io")
				require.Contains(t, patches[0], "http://docker.io:5000")
			},
		},
		{
			name: "multiple mirror registries",
			mirrors: []string{
				"docker.io=https://registry-1.docker.io",
				"ghcr.io=https://ghcr.io",
				"quay.io=https://quay.io",
			},
			assert: func(t *testing.T, patches []string) {
				require.Len(t, patches, 3)
				require.Contains(t, patches[0], "docker.io")
				require.Contains(t, patches[0], "http://docker.io:5000")
				require.Contains(t, patches[1], "ghcr.io")
				require.Contains(t, patches[1], "http://ghcr.io:5001")
				require.Contains(t, patches[2], "quay.io")
				require.Contains(t, patches[2], "http://quay.io:5002")
			},
		},
		{
			name:    "no mirror registries",
			mirrors: []string{},
			assert: func(t *testing.T, patches []string) {
				require.Empty(t, patches)
			},
		},
		{
			name: "invalid mirror spec skipped",
			mirrors: []string{
				"docker.io=https://registry-1.docker.io",
				"invalid-spec-no-equals",
				"ghcr.io=https://ghcr.io",
			},
			assert: func(t *testing.T, patches []string) {
				require.Len(t, patches, 2)
				require.Contains(t, patches[0], "docker.io")
				require.Contains(t, patches[0], "http://docker.io:5000")
				require.Contains(t, patches[1], "ghcr.io")
				require.Contains(t, patches[1], "http://ghcr.io:5001")
			},
		},
		{
			name:    "custom port in upstream URL",
			mirrors: []string{"localhost=http://localhost:5001"},
			assert: func(t *testing.T, patches []string) {
				require.Len(t, patches, 1)
				require.Contains(t, patches[0], "localhost")
				require.Contains(t, patches[0], "http://localhost:5000")
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			scaf := createTestScaffolderForKind()
			scaf.MirrorRegistries = testCase.mirrors

			patches := scaf.GenerateContainerdPatches()
			testCase.assert(t, patches)
		})
	}
}

// TestGenerateK3dRegistryConfig tests the generation of K3d registry configuration.
func TestGenerateK3dRegistryConfig(t *testing.T) {
	t.Parallel()

	t.Run("single mirror registry", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForK3d()
		scaf.MirrorRegistries = []string{"docker.io=https://registry-1.docker.io"}

		registryConfig := scaf.GenerateK3dRegistryConfig()

		require.Nil(t, registryConfig.Create)
		require.Contains(t, registryConfig.Config, "\"docker.io\":")
		require.Contains(t, registryConfig.Config, "https://registry-1.docker.io")
		require.NotContains(t, registryConfig.Config, "http://docker.io:5000")
		require.Empty(t, registryConfig.Use)
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

	t.Run("multiple mirror registries", func(t *testing.T) {
		t.Parallel()

		scaf := createTestScaffolderForK3d()
		scaf.MirrorRegistries = []string{
			"docker.io=https://registry-1.docker.io",
			"ghcr.io=https://ghcr.io",
		}

		registryConfig := scaf.GenerateK3dRegistryConfig()

		require.Contains(t, registryConfig.Config, "\"docker.io\":")
		require.Contains(t, registryConfig.Config, "\"ghcr.io\":")
		require.Contains(t, registryConfig.Config, "https://registry-1.docker.io")
		require.Contains(t, registryConfig.Config, "https://ghcr.io")
		require.NotContains(t, registryConfig.Config, "http://docker.io:5000")
		require.NotContains(t, registryConfig.Config, "http://ghcr.io:5001")
		require.Nil(t, registryConfig.Create)
		require.Empty(t, registryConfig.Use)
	})
}
