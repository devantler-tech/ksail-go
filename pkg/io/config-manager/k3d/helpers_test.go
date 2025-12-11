package k3d_test

import (
	"testing"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
)

//nolint:funlen // Table-driven test with many scenarios
func TestParseRegistryConfig(t *testing.T) {
	t.Parallel()

	t.Run("returns_empty_map_for_empty_string", func(t *testing.T) {
		t.Parallel()

		result := k3d.ParseRegistryConfig("")

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns_empty_map_for_whitespace_only", func(t *testing.T) {
		t.Parallel()

		result := k3d.ParseRegistryConfig("   \n\t  ")

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("returns_empty_map_for_invalid_yaml", func(t *testing.T) {
		t.Parallel()

		invalidYAML := `
		invalid: [
		  - this is not valid
		`

		result := k3d.ParseRegistryConfig(invalidYAML)

		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("parses_single_mirror_with_one_endpoint", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint:
      - http://localhost:5000
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "docker.io")
		assert.Equal(t, []string{"http://localhost:5000"}, result["docker.io"])
	})

	t.Run("parses_single_mirror_with_multiple_endpoints", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint:
      - http://localhost:5000
      - http://localhost:5001
      - http://localhost:5002
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "docker.io")
		assert.Equal(t, []string{
			"http://localhost:5000",
			"http://localhost:5001",
			"http://localhost:5002",
		}, result["docker.io"])
	})

	t.Run("parses_multiple_mirrors", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint:
      - http://localhost:5000
  ghcr.io:
    endpoint:
      - http://localhost:5001
  registry.k8s.io:
    endpoint:
      - http://localhost:5002
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 3)
		assert.Contains(t, result, "docker.io")
		assert.Contains(t, result, "ghcr.io")
		assert.Contains(t, result, "registry.k8s.io")
		assert.Equal(t, []string{"http://localhost:5000"}, result["docker.io"])
		assert.Equal(t, []string{"http://localhost:5001"}, result["ghcr.io"])
		assert.Equal(t, []string{"http://localhost:5002"}, result["registry.k8s.io"])
	})

	t.Run("filters_out_empty_endpoints", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint:
      - http://localhost:5000
      - ""
      - http://localhost:5001
      - "  "
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "docker.io")
		assert.Equal(t, []string{
			"http://localhost:5000",
			"http://localhost:5001",
		}, result["docker.io"])
	})

	t.Run("trims_whitespace_from_endpoints", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint:
      - "  http://localhost:5000  "
      - "	http://localhost:5001	"
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "docker.io")
		assert.Equal(t, []string{
			"http://localhost:5000",
			"http://localhost:5001",
		}, result["docker.io"])
	})

	t.Run("skips_mirrors_with_no_endpoints", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint: []
  ghcr.io:
    endpoint:
      - http://localhost:5000
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "ghcr.io")
		assert.NotContains(t, result, "docker.io")
	})

	t.Run("skips_mirrors_with_only_empty_endpoints", func(t *testing.T) {
		t.Parallel()

		yaml := `
mirrors:
  docker.io:
    endpoint:
      - ""
      - "  "
  ghcr.io:
    endpoint:
      - http://localhost:5000
`

		result := k3d.ParseRegistryConfig(yaml)

		assert.Len(t, result, 1)
		assert.Contains(t, result, "ghcr.io")
		assert.NotContains(t, result, "docker.io")
	})
}

//nolint:funlen // Table-driven test with many scenarios
func TestResolveClusterName(t *testing.T) {
	t.Parallel()

	t.Run("returns_default_when_both_configs_are_nil", func(t *testing.T) {
		t.Parallel()

		name := k3d.ResolveClusterName(nil, nil)

		assert.Equal(t, "k3d", name)
	})

	t.Run("returns_k3d_config_name_when_present", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &v1alpha5.SimpleConfig{
			ObjectMeta: types.ObjectMeta{
				Name: "test-k3d-cluster",
			},
		}
		clusterCfg := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Context: "should-not-use-this",
				},
			},
		}

		name := k3d.ResolveClusterName(clusterCfg, k3dConfig)

		assert.Equal(t, "test-k3d-cluster", name)
	})

	t.Run("returns_cluster_context_when_k3d_name_is_empty", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &v1alpha5.SimpleConfig{
			ObjectMeta: types.ObjectMeta{
				Name: "",
			},
		}
		clusterCfg := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Context: "test-context",
				},
			},
		}

		name := k3d.ResolveClusterName(clusterCfg, k3dConfig)

		assert.Equal(t, "test-context", name)
	})

	t.Run("returns_cluster_context_when_k3d_config_is_nil", func(t *testing.T) {
		t.Parallel()

		clusterCfg := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Context: "test-context",
				},
			},
		}

		name := k3d.ResolveClusterName(clusterCfg, nil)

		assert.Equal(t, "test-context", name)
	})

	t.Run("returns_default_when_both_names_are_empty", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &v1alpha5.SimpleConfig{
			ObjectMeta: types.ObjectMeta{
				Name: "",
			},
		}
		clusterCfg := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Context: "",
				},
			},
		}

		name := k3d.ResolveClusterName(clusterCfg, k3dConfig)

		assert.Equal(t, "k3d", name)
	})

	t.Run("trims_whitespace_from_k3d_name", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &v1alpha5.SimpleConfig{
			ObjectMeta: types.ObjectMeta{
				Name: "  test-cluster  ",
			},
		}

		name := k3d.ResolveClusterName(nil, k3dConfig)

		assert.Equal(t, "test-cluster", name)
	})

	t.Run("trims_whitespace_from_cluster_context", func(t *testing.T) {
		t.Parallel()

		clusterCfg := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Context: "  test-context  ",
				},
			},
		}

		name := k3d.ResolveClusterName(clusterCfg, nil)

		assert.Equal(t, "test-context", name)
	})

	t.Run("returns_default_when_k3d_name_is_whitespace_only", func(t *testing.T) {
		t.Parallel()

		k3dConfig := &v1alpha5.SimpleConfig{
			ObjectMeta: types.ObjectMeta{
				Name: "   ",
			},
		}
		clusterCfg := &v1alpha1.Cluster{
			Spec: v1alpha1.Spec{
				Connection: v1alpha1.Connection{
					Context: "",
				},
			},
		}

		name := k3d.ResolveClusterName(clusterCfg, k3dConfig)

		assert.Equal(t, "k3d", name)
	})

	t.Run("returns_cluster_context_when_cluster_cfg_is_nil", func(t *testing.T) {
		t.Parallel()

		// When clusterCfg is nil, should fall through to default
		name := k3d.ResolveClusterName(nil, nil)

		assert.Equal(t, "k3d", name)
	})
}
