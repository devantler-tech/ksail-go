package k3dprovisioner

import (
	"context"
	"fmt"
	"io"
	"strings"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	k3dtypes "github.com/k3d-io/k3d/v5/pkg/types"
	"sigs.k8s.io/yaml"
)

// SetupRegistries creates mirror registries based on the K3d simple configuration.
func SetupRegistries(
	ctx context.Context,
	simpleCfg *k3dv1alpha5.SimpleConfig,
	clusterName string,
	dockerClient client.APIClient,
	writer io.Writer,
) error {
	registryMgr, registryInfos, err := setupRegistryManager(simpleCfg, dockerClient)
	if err != nil {
		return err
	}

	if registryMgr == nil {
		return nil
	}

	errRegistry := registries.SetupRegistries(ctx, registryMgr, registryInfos, clusterName, writer)
	if errRegistry != nil {
		return fmt.Errorf("failed to setup k3d registries: %w", errRegistry)
	}

	return nil
}

// ConnectRegistriesToNetwork attaches registry containers to the K3d cluster network.
func ConnectRegistriesToNetwork(
	ctx context.Context,
	simpleCfg *k3dv1alpha5.SimpleConfig,
	clusterName string,
	dockerClient client.APIClient,
	writer io.Writer,
) error {
	if simpleCfg == nil {
		return nil
	}

	registryInfos := extractRegistriesFromConfig(simpleCfg)
	if len(registryInfos) == 0 {
		return nil
	}

	networkName := "k3d-" + strings.TrimSpace(clusterName)
	if strings.TrimSpace(networkName) == "k3d-" {
		networkName = "k3d"
	}

	errConnect := registries.ConnectRegistriesToNetwork(
		ctx,
		dockerClient,
		registryInfos,
		networkName,
		writer,
	)
	if errConnect != nil {
		return fmt.Errorf("failed to connect k3d registries to network: %w", errConnect)
	}

	return nil
}

// CleanupRegistries removes registry containers associated with the cluster.
func CleanupRegistries(
	ctx context.Context,
	simpleCfg *k3dv1alpha5.SimpleConfig,
	clusterName string,
	dockerClient client.APIClient,
	deleteVolumes bool,
	writer io.Writer,
) error {
	registryMgr, registryInfos, err := setupRegistryManager(simpleCfg, dockerClient)
	if err != nil {
		return err
	}

	if registryMgr == nil {
		return nil
	}

	errCleanup := registries.CleanupRegistries(
		ctx,
		registryMgr,
		registryInfos,
		clusterName,
		deleteVolumes,
		writer,
	)
	if errCleanup != nil {
		return fmt.Errorf("failed to cleanup k3d registries: %w", errCleanup)
	}

	return nil
}

func setupRegistryManager(
	simpleCfg *k3dv1alpha5.SimpleConfig,
	dockerClient client.APIClient,
) (*dockerclient.RegistryManager, []registries.Info, error) {
	if simpleCfg == nil {
		return nil, nil, nil
	}

	registryInfos := extractRegistriesFromConfig(simpleCfg)
	if len(registryInfos) == 0 {
		return nil, nil, nil
	}

	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create registry manager: %w", err)
	}

	return registryMgr, registryInfos, nil
}

type mirrorConfig struct {
	Endpoint []string `yaml:"endpoint"`
}

type k3dMirrorConfig struct {
	Mirrors map[string]mirrorConfig `yaml:"mirrors"`
}

func extractRegistriesFromConfig(simpleCfg *k3dv1alpha5.SimpleConfig) []registries.Info {
	if simpleCfg == nil {
		return nil
	}

	configStr := strings.TrimSpace(simpleCfg.Registries.Config)
	if configStr == "" {
		return nil
	}

	var mirrorCfg k3dMirrorConfig

	err := yaml.Unmarshal([]byte(configStr), &mirrorCfg)
	if err != nil {
		return nil
	}

	if len(mirrorCfg.Mirrors) == 0 {
		return nil
	}

	hosts := make([]string, 0, len(mirrorCfg.Mirrors))
	for host := range mirrorCfg.Mirrors {
		hosts = append(hosts, host)
	}

	registries.SortHosts(hosts)

	usedPorts := make(map[int]struct{})
	nextPort := registries.DefaultRegistryPort

	registryInfos := make([]registries.Info, 0, len(hosts))
	prefix := k3dtypes.DefaultObjectNamePrefix + "-"

	for _, host := range hosts {
		endpoints := mirrorCfg.Mirrors[host].Endpoint
		port := registries.ExtractRegistryPort(endpoints, usedPorts, &nextPort)
		info := registries.BuildRegistryInfo(host, endpoints, port, prefix, "")
		registryInfos = append(registryInfos, info)
	}

	return registryInfos
}

// ExtractRegistriesFromConfigForTesting exposes registry extraction for testing and callers that need inspection.
func ExtractRegistriesFromConfigForTesting(simpleCfg *k3dv1alpha5.SimpleConfig) []registries.Info {
	return extractRegistriesFromConfig(simpleCfg)
}
