package scaffolder_test

import (
	"io"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/scaffolder"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

type containerdPatchCase struct {
	name    string
	mirrors []string
	assert  func(t *testing.T, patches []string)
}

func containerdPatchCases() []containerdPatchCase {
	return []containerdPatchCase{
		containerdSingleMirrorCase(),
		containerdMultipleMirrorCase(),
		containerdNoMirrorCase(),
		containerdInvalidMirrorCase(),
		containerdCustomPortCase(),
	}
}

func containerdSingleMirrorCase() containerdPatchCase {
	return containerdPatchCase{
		name:    "single mirror registry",
		mirrors: []string{"docker.io=https://registry-1.docker.io"},
		assert: func(t *testing.T, patches []string) {
			t.Helper()
			require.Len(t, patches, 1)
			require.Contains(t, patches[0], "docker.io")
			require.Contains(t, patches[0], "http://docker.io:5000")
		},
	}
}

func containerdMultipleMirrorCase() containerdPatchCase {
	return containerdPatchCase{
		name: "multiple mirror registries",
		mirrors: []string{
			"docker.io=https://registry-1.docker.io",
			"ghcr.io=https://ghcr.io",
			"quay.io=https://quay.io",
		},
		assert: func(t *testing.T, patches []string) {
			t.Helper()
			require.Len(t, patches, 3)
			require.Contains(t, patches[0], "docker.io")
			require.Contains(t, patches[0], "http://docker.io:5000")
			require.Contains(t, patches[1], "ghcr.io")
			require.Contains(t, patches[1], "http://ghcr.io:5001")
			require.Contains(t, patches[2], "quay.io")
			require.Contains(t, patches[2], "http://quay.io:5002")
		},
	}
}

func containerdNoMirrorCase() containerdPatchCase {
	return containerdPatchCase{
		name:    "no mirror registries",
		mirrors: []string{},
		assert: func(t *testing.T, patches []string) {
			t.Helper()
			require.Empty(t, patches)
		},
	}
}

func containerdInvalidMirrorCase() containerdPatchCase {
	return containerdPatchCase{
		name: "invalid mirror spec skipped",
		mirrors: []string{
			"docker.io=https://registry-1.docker.io",
			"invalid-spec-no-equals",
			"ghcr.io=https://ghcr.io",
		},
		assert: func(t *testing.T, patches []string) {
			t.Helper()
			require.Len(t, patches, 2)
			require.Contains(t, patches[0], "docker.io")
			require.Contains(t, patches[0], "http://docker.io:5000")
			require.Contains(t, patches[1], "ghcr.io")
			require.Contains(t, patches[1], "http://ghcr.io:5001")
		},
	}
}

func containerdCustomPortCase() containerdPatchCase {
	return containerdPatchCase{
		name:    "custom port in upstream URL",
		mirrors: []string{"localhost=http://localhost:5001"},
		assert: func(t *testing.T, patches []string) {
			t.Helper()
			require.Len(t, patches, 1)
			require.Contains(t, patches[0], "localhost")
			require.Contains(t, patches[0], "http://localhost:5000")
		},
	}
}

type k3dRegistryConfigCase struct {
	name    string
	mirrors []string
	assert  func(t *testing.T, config k3dv1alpha5.SimpleConfigRegistries)
}

func k3dRegistryConfigCases() []k3dRegistryConfigCase {
	return []k3dRegistryConfigCase{
		{
			name:    "single mirror registry",
			mirrors: []string{"docker.io=https://registry-1.docker.io"},
			assert: func(t *testing.T, config k3dv1alpha5.SimpleConfigRegistries) {
				t.Helper()
				require.Nil(t, config.Create)
				require.Contains(t, config.Config, "\"docker.io\":")
				require.Contains(t, config.Config, "https://registry-1.docker.io")
				require.NotContains(t, config.Config, "http://docker.io:5000")
				require.Empty(t, config.Use)
			},
		},
		{
			name:    "no mirror registries",
			mirrors: []string{},
			assert: func(t *testing.T, config k3dv1alpha5.SimpleConfigRegistries) {
				t.Helper()
				require.Empty(t, config.Use)
				require.Nil(t, config.Create)
				require.Empty(t, config.Config)
			},
		},
		{
			name:    "invalid mirror spec",
			mirrors: []string{"invalid-no-equals"},
			assert: func(t *testing.T, config k3dv1alpha5.SimpleConfigRegistries) {
				t.Helper()
				require.Empty(t, config.Use)
				require.Nil(t, config.Create)
				require.Empty(t, config.Config)
			},
		},
		{
			name: "multiple mirror registries",
			mirrors: []string{
				"docker.io=https://registry-1.docker.io",
				"ghcr.io=https://ghcr.io",
			},
			assert: func(t *testing.T, config k3dv1alpha5.SimpleConfigRegistries) {
				t.Helper()
				require.Contains(t, config.Config, "\"docker.io\":")
				require.Contains(t, config.Config, "\"ghcr.io\":")
				require.Contains(t, config.Config, "https://registry-1.docker.io")
				require.Contains(t, config.Config, "https://ghcr.io")
				require.NotContains(t, config.Config, "http://docker.io:5000")
				require.NotContains(t, config.Config, "http://ghcr.io:5001")
				require.Nil(t, config.Create)
				require.Empty(t, config.Use)
			},
		},
	}
}

func TestGenerateContainerdPatches(t *testing.T) {
	t.Parallel()

	for _, testCase := range containerdPatchCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			scaf := createTestScaffolderForKind()
			scaf.MirrorRegistries = testCase.mirrors

			patches := scaf.GenerateContainerdPatches()
			testCase.assert(t, patches)
		})
	}
}

func TestGenerateK3dRegistryConfig(t *testing.T) {
	t.Parallel()

	for _, testCase := range k3dRegistryConfigCases() {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			scaf := createTestScaffolderForK3d()
			scaf.MirrorRegistries = testCase.mirrors

			registryConfig := scaf.GenerateK3dRegistryConfig()
			testCase.assert(t, registryConfig)
		})
	}
}
