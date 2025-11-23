package registry

import (
	"context"
	"errors"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

var (
	// ErrNameRequired indicates that an operation was attempted without providing a registry name.
	ErrNameRequired = errors.New("registry name is required")
	// ErrHostRequired indicates that no host or bind address was provided for the registry endpoint.
	ErrHostRequired = errors.New("registry host is required")
	// ErrInvalidPort indicates that a provided registry port is outside the valid TCP port range.
	ErrInvalidPort = errors.New("registry port must be between 1 and 65535")
)

// Service models the lifecycle management interface for localhost-scoped OCI registries.
type Service interface {
	// Create provisions (or updates) an OCI registry container definition using the supplied options.
	Create(ctx context.Context, opts CreateOptions) (v1alpha1.OCIRegistry, error)
	// Start ensures the registry container is running and optionally attached to the target network.
	Start(ctx context.Context, opts StartOptions) (v1alpha1.OCIRegistry, error)
	// Stop halts the running registry container and optionally removes persistent storage resources.
	Stop(ctx context.Context, opts StopOptions) error
	// Status inspects the registry container and returns its current lifecycle state and metadata.
	Status(ctx context.Context, opts StatusOptions) (v1alpha1.OCIRegistry, error)
}
