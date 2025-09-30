package stubs

import (
	"errors"

	kindcluster "sigs.k8s.io/kind/pkg/cluster"
)

// KindProviderStub is a stub implementation of KindProvider interface.
type KindProviderStub struct {
	CreateError     error
	DeleteError     error
	ListResult      []string
	ListError       error
	ListNodesResult []string
	ListNodesError  error

	CreateCalls    []string
	DeleteCalls    []string
	ListCalls      int
	ListNodesCalls []string
}

// NewKindProviderStub creates a new KindProviderStub with default behavior.
func NewKindProviderStub() *KindProviderStub {
	return &KindProviderStub{
		ListResult:      []string{"kind-cluster"},
		ListNodesResult: []string{"kind-cluster-control-plane"},
	}
}

// Create simulates cluster creation.
func (k *KindProviderStub) Create(name string, opts ...kindcluster.CreateOption) error {
	k.CreateCalls = append(k.CreateCalls, name)
	return k.CreateError
}

// Delete simulates cluster deletion.
func (k *KindProviderStub) Delete(name, kubeconfigPath string) error {
	k.DeleteCalls = append(k.DeleteCalls, name)
	return k.DeleteError
}

// List simulates cluster listing.
func (k *KindProviderStub) List() ([]string, error) {
	k.ListCalls++
	return k.ListResult, k.ListError
}

// ListNodes simulates node listing for a cluster.
func (k *KindProviderStub) ListNodes(name string) ([]string, error) {
	k.ListNodesCalls = append(k.ListNodesCalls, name)
	return k.ListNodesResult, k.ListNodesError
}

// WithCreateError configures the stub to return an error on Create.
func (k *KindProviderStub) WithCreateError(message string) *KindProviderStub {
	k.CreateError = errors.New(message)
	return k
}

// WithListResult configures the stub to return specific clusters on List.
func (k *KindProviderStub) WithListResult(clusters []string) *KindProviderStub {
	k.ListResult = clusters
	return k
}
