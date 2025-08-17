// Package di provides dependency injection utilities for KSail.
package di

import (
	factory "github.com/devantler-tech/ksail-go/internal/factories"
	"github.com/devantler-tech/ksail-go/internal/validators"
	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	reconciliationtoolbootstrapper "github.com/devantler-tech/ksail-go/pkg/bootstrapper/reconciliation_tool"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	containerengineprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine"
)

// Services holds all the initialized services and configuration.
type Services struct {
	Config                         ksailcluster.Cluster
	ClusterProvisioner             clusterprovisioner.ClusterProvisioner
	ContainerEngineProvisioner     containerengineprovisioner.ContainerEngineProvisioner
	ReconciliationToolBootstrapper reconciliationtoolbootstrapper.Bootstrapper
	ConfigValidator                *validators.ConfigValidator
}

// InitServices initializes the services required by the CLI using the provided configuration.
func InitServices(ksailConfig *ksailcluster.Cluster) (*Services, error) {
	clusterProvisioner, err := factory.ClusterProvisioner(ksailConfig)
	if err != nil {
		return nil, err
	}

	containerEngineProvisioner, err := factory.ContainerEngineProvisioner(ksailConfig)
	if err != nil {
		return nil, err
	}

	reconciliationToolBootstrapper, err := factory.ReconciliationTool(ksailConfig)
	if err != nil {
		return nil, err
	}

	configValidator := validators.NewConfigValidator(ksailConfig)

	return &Services{
		Config:                         *ksailConfig,
		ClusterProvisioner:             clusterProvisioner,
		ContainerEngineProvisioner:     containerEngineProvisioner,
		ReconciliationToolBootstrapper: reconciliationToolBootstrapper,
		ConfigValidator:                configValidator,
	}, nil
}
