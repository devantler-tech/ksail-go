package di

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"go.uber.org/dig"
)

var ErrClusterConfigRequired = errors.New(
	"cluster configuration is required for dependency resolution",
)

// ResolvedDependencies contains the concrete services required by cluster commands.
type ResolvedDependencies struct {
	Provisioner        clusterprovisioner.ClusterProvisioner
	DistributionConfig any
}

// Resolver provides a minimal dig-backed dependency resolver.
type Resolver struct {
	cluster *v1alpha1.Cluster
	writer  io.Writer
	timer   timer.Timer
}

// NewResolver creates a new dependency resolver instance.
func NewResolver(cluster *v1alpha1.Cluster, writer io.Writer, tmr timer.Timer) (*Resolver, error) {
	if cluster == nil {
		return nil, ErrClusterConfigRequired
	}

	if writer == nil {
		writer = io.Discard
	}

	if tmr == nil {
		tmr = timer.New()
	}

	return &Resolver{
		cluster: cluster,
		writer:  writer,
		timer:   tmr,
	}, nil
}

// Resolve executes the dependency resolution flow.
func (r *Resolver) Resolve() (*ResolvedDependencies, error) {
	container := dig.New()

	if err := container.Provide(r.provide); err != nil {
		return nil, fmt.Errorf("provide resolved dependencies: %w", err)
	}

	var deps *ResolvedDependencies
	if err := container.Invoke(func(d *ResolvedDependencies) {
		deps = d
	}); err != nil {
		return nil, fmt.Errorf("resolve dependencies: %w", err)
	}

	if deps == nil {
		return nil, errors.New("resolved dependencies not found in container")
	}

	return deps, nil
}

func (r *Resolver) provide() (*ResolvedDependencies, error) {
	provisioner, distributionConfig, err := clusterprovisioner.CreateClusterProvisioner(
		context.Background(),
		r.cluster.Spec.Distribution,
		r.cluster.Spec.DistributionConfig,
		r.cluster.Spec.Connection.Kubeconfig,
	)
	if err != nil {
		return nil, fmt.Errorf("create cluster provisioner: %w", err)
	}

	return &ResolvedDependencies{
		Provisioner:        provisioner,
		DistributionConfig: distributionConfig,
	}, nil
}
