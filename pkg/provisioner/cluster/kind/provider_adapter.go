package kindprovisioner

import (
	"fmt"

	"sigs.k8s.io/kind/pkg/cluster"
)

// KindProviderAdapter wraps sigs.k8s.io/kind/pkg/cluster.Provider to implement KindProvider interface.
type KindProviderAdapter struct {
	provider *cluster.Provider
}

// NewKindProviderAdapter creates a new adapter wrapping the real Kind provider.
func NewKindProviderAdapter(provider *cluster.Provider) *KindProviderAdapter {
	return &KindProviderAdapter{
		provider: provider,
	}
}

// Create creates a new kind cluster.
func (a *KindProviderAdapter) Create(name string, opts ...cluster.CreateOption) error {
	err := a.provider.Create(name, opts...)
	if err != nil {
		return fmt.Errorf("failed to create kind cluster %s: %w", name, err)
	}

	return nil
}

// Delete deletes a kind cluster.
func (a *KindProviderAdapter) Delete(name, kubeconfigPath string) error {
	err := a.provider.Delete(name, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster %s: %w", name, err)
	}

	return nil
}

// List lists kind clusters.
func (a *KindProviderAdapter) List() ([]string, error) {
	res, err := a.provider.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	return res, nil
}

// ListNodes lists nodes in a kind cluster, converting nodes.Node to string names.
func (a *KindProviderAdapter) ListNodes(name string) ([]string, error) {
	nodesList, err := a.provider.ListNodes(name)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes for cluster %s: %w", name, err)
	}

	// Convert []nodes.Node to []string by extracting node names
	nodeNames := make([]string, len(nodesList))
	for i, node := range nodesList {
		nodeNames[i] = node.String()
	}

	return nodeNames, nil
}
