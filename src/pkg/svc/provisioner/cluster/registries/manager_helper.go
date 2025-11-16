package registries

import (
	"context"
	"fmt"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/client"
)

// PrepareRegistryManager builds a registry manager and extracts registry info via the provided extractor.
// The extractor receives the set of already used ports (if any) and should return the registries that need work.
func PrepareRegistryManager(
	ctx context.Context,
	dockerClient client.APIClient,
	extractor func(baseUsedPorts map[int]struct{}) []Info,
) (*dockerclient.RegistryManager, []Info, error) {
	regManager, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create registry manager: %w", err)
	}

	registryInfos := extractor(nil)
	if len(registryInfos) == 0 {
		return nil, nil, nil
	}

	existingPorts, err := CollectExistingRegistryPorts(ctx, regManager)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to collect existing registry ports: %w", err)
	}

	if len(existingPorts) != 0 {
		registryInfos = extractor(existingPorts)
	}

	return regManager, registryInfos, nil
}
