// Package k3dgenerator provides utilities for generating k3d cluster configurations.
package k3dgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// K3dGenerator generates a k3d SimpleConfig YAML.
type K3dGenerator struct {
	Marshaller marshaller.Marshaller[*v1alpha5.SimpleConfig]
}

// NewK3dGenerator creates and returns a new K3dGenerator instance.
func NewK3dGenerator() *K3dGenerator {
	m := yamlmarshaller.NewMarshaller[*v1alpha5.SimpleConfig]()

	return &K3dGenerator{
		Marshaller: m,
	}
}

// Generate creates a k3d cluster YAML configuration and writes it to the specified output.
func (g *K3dGenerator) Generate(cluster *v1alpha1.Cluster, opts yamlgenerator.Options) (string, error) {
	cfg := g.buildSimpleConfig(cluster)

	out, err := g.Marshaller.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal k3d config: %w", err)
	}

	// write to file if output path is specified
	if opts.Output != "" {
		result, err := io.TryWriteFile(out, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("write k3d config: %w", err)
		}

		return result, nil
	}

	return out, nil
}

func (g *K3dGenerator) buildSimpleConfig(cluster *v1alpha1.Cluster) *v1alpha5.SimpleConfig {
	return &v1alpha5.SimpleConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: types.ObjectMeta{
			Name: cluster.Metadata.Name,
		},
		Servers:      0,
		Agents:       0,
		ExposeAPI:    g.buildExposureOpts(),
		Image:        "",
		Network:      "",
		Subnet:       "",
		ClusterToken: "",
		Volumes:      nil,
		Ports:        nil,
		Options:      g.buildConfigOptions(),
		Env:          nil,
		Registries:   g.buildRegistries(),
		HostAliases:  nil,
		Files:        nil,
	}
}

func (g *K3dGenerator) buildExposureOpts() v1alpha5.SimpleExposureOpts {
	return v1alpha5.SimpleExposureOpts{
		Host:     "",
		HostIP:   "",
		HostPort: "",
	}
}

func (g *K3dGenerator) buildConfigOptions() v1alpha5.SimpleConfigOptions {
	return v1alpha5.SimpleConfigOptions{
		K3dOptions:        BuildDefaultK3dOptions(),
		K3sOptions:        BuildDefaultK3sOptions(),
		KubeconfigOptions: BuildDefaultKubeconfigOptions(),
		Runtime:           BuildDefaultRuntimeOptions(),
	}
}

// BuildDefaultK3dOptions creates default K3d options configuration.
func BuildDefaultK3dOptions() v1alpha5.SimpleConfigOptionsK3d {
	return v1alpha5.SimpleConfigOptionsK3d{
		Wait:                false,
		Timeout:             0,
		DisableLoadbalancer: false,
		DisableImageVolume:  false,
		NoRollback:          false,
		NodeHookActions:     nil,
		Loadbalancer: v1alpha5.SimpleConfigOptionsK3dLoadbalancer{
			ConfigOverrides: nil,
		},
	}
}

// BuildDefaultK3sOptions creates default K3s options configuration.
func BuildDefaultK3sOptions() v1alpha5.SimpleConfigOptionsK3s {
	return v1alpha5.SimpleConfigOptionsK3s{
		ExtraArgs:  nil,
		NodeLabels: nil,
	}
}

// BuildDefaultKubeconfigOptions creates default kubeconfig options configuration.
func BuildDefaultKubeconfigOptions() v1alpha5.SimpleConfigOptionsKubeconfig {
	return v1alpha5.SimpleConfigOptionsKubeconfig{
		UpdateDefaultKubeconfig: false,
		SwitchCurrentContext:    false,
	}
}

// BuildDefaultRuntimeOptions creates default runtime options configuration.
func BuildDefaultRuntimeOptions() v1alpha5.SimpleConfigOptionsRuntime {
	return v1alpha5.SimpleConfigOptionsRuntime{
		GPURequest:    "",
		ServersMemory: "",
		AgentsMemory:  "",
		HostPidMode:   false,
		Labels:        nil,
		Ulimits:       nil,
	}
}

func (g *K3dGenerator) buildRegistries() v1alpha5.SimpleConfigRegistries {
	return v1alpha5.SimpleConfigRegistries{
		Use:    nil,
		Create: nil,
		Config: "",
	}
}
