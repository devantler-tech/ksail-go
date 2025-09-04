// Package eksprovisioner provides implementations of the Provisioner interface
// for provisioning EKS clusters in AWS.
package eksprovisioner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
)

// ErrClusterNotFound is returned when a cluster is not found.
var ErrClusterNotFound = errors.New("cluster not found")

const (
	// DefaultWaitInterval is the default wait interval for cluster operations.
	DefaultWaitInterval = 30 * time.Second
	// DefaultPodEvictionWaitPeriod is the default wait period for pod eviction.
	DefaultPodEvictionWaitPeriod = 10 * time.Minute
	// DefaultChunkSize is the default chunk size for listing clusters.
	DefaultChunkSize = 100
	// DefaultParallelism is the default parallelism for cluster operations.
	DefaultParallelism = 4
)

// EKSClusterProvisioner is an implementation of the ClusterProvisioner interface for provisioning EKS clusters.
type EKSClusterProvisioner struct {
	clusterConfig           *v1alpha5.ClusterConfig
	providerConstructor     EKSProviderConstructor
	clusterActionsFactory   EKSClusterActionsFactory
	clusterLister           EKSClusterLister
	clusterCreator          EKSClusterCreator
	nodeGroupManagerFactory EKSNodeGroupManagerFactory
}

// NewEKSClusterProvisioner constructs an EKSClusterProvisioner with explicit dependencies
// for the eksctl provider and cluster actions. This supports both production wiring
// and unit testing via mocks.
func NewEKSClusterProvisioner(
	clusterConfig *v1alpha5.ClusterConfig,
	providerConstructor EKSProviderConstructor,
	clusterActionsFactory EKSClusterActionsFactory,
	clusterLister EKSClusterLister,
	clusterCreator EKSClusterCreator,
	nodeGroupManagerFactory EKSNodeGroupManagerFactory,
) *EKSClusterProvisioner {
	return &EKSClusterProvisioner{
		clusterConfig:           clusterConfig,
		providerConstructor:     providerConstructor,
		clusterActionsFactory:   clusterActionsFactory,
		clusterLister:           clusterLister,
		clusterCreator:          clusterCreator,
		nodeGroupManagerFactory: nodeGroupManagerFactory,
	}
}

// createProvider creates a provider with explicit config for the given cluster.
func (e *EKSClusterProvisioner) createProvider(ctx context.Context) (*eks.ClusterProvider, error) {
	// Create provider with explicit config
	providerConfig := &v1alpha5.ProviderConfig{
		Region: e.clusterConfig.Metadata.Region,
	}

	ctl, err := e.providerConstructor.NewClusterProvider(ctx, providerConfig, e.clusterConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create EKS provider: %w", err)
	}

	return ctl, nil
}

// ensureClusterExists checks if a cluster exists and returns ErrClusterNotFound if not.
func (e *EKSClusterProvisioner) ensureClusterExists(name string) error {
	exists, err := e.Exists(name)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if !exists {
		return ErrClusterNotFound
	}

	return nil
}

// setupClusterOperation sets up common cluster operation prerequisites.
func (e *EKSClusterProvisioner) setupClusterOperation(ctx context.Context, name string) (*eks.ClusterProvider, error) {
	target := setName(name, e.clusterConfig.Metadata.Name)
	e.clusterConfig.Metadata.Name = target

	return e.createProvider(ctx)
}

// setupNodeGroupManager sets up common node group management prerequisites.
func (e *EKSClusterProvisioner) setupNodeGroupManager(ctx context.Context, name string) (EKSNodeGroupManager, error) {
	if err := e.ensureClusterExists(name); err != nil {
		return nil, err
	}

	ctl, err := e.setupClusterOperation(ctx, name)
	if err != nil {
		return nil, err
	}

	// Create node group manager
	ngManager := e.nodeGroupManagerFactory.NewNodeGroupManager(e.clusterConfig, ctl, nil, nil)
	return ngManager, nil
}

// Create creates an EKS cluster.
func (e *EKSClusterProvisioner) Create(name string) error {
	ctx := context.Background()

	ctl, err := e.setupClusterOperation(ctx, name)
	if err != nil {
		return err
	}

	err = e.clusterCreator.CreateCluster(ctx, e.clusterConfig, ctl)
	if err != nil {
		return fmt.Errorf("failed to create EKS cluster: %w", err)
	}

	return nil
}

// Delete deletes an EKS cluster.
func (e *EKSClusterProvisioner) Delete(name string) error {
	ctx := context.Background()

	ctl, err := e.setupClusterOperation(ctx, name)
	if err != nil {
		return err
	}

	clusterActions, err := e.clusterActionsFactory.NewClusterActions(ctx, e.clusterConfig, ctl)
	if err != nil {
		return fmt.Errorf("failed to create cluster actions: %w", err)
	}

	// Use reasonable defaults for deletion parameters
	waitInterval := DefaultWaitInterval
	podEvictionWaitPeriod := DefaultPodEvictionWaitPeriod
	wait := true
	force := false
	disableNodegroupEviction := false
	parallel := DefaultParallelism

	err = clusterActions.Delete(ctx, waitInterval, podEvictionWaitPeriod, wait, force, disableNodegroupEviction, parallel)
	if err != nil {
		return fmt.Errorf("failed to delete EKS cluster: %w", err)
	}

	return nil
}

// Start starts an EKS cluster by scaling node groups from 0 to their desired capacity.
func (e *EKSClusterProvisioner) Start(name string) error {
	ctx := context.Background()

	ngManager, err := e.setupNodeGroupManager(ctx, name)
	if err != nil {
		return err
	}

	// Scale all node groups to their desired capacity
	for _, ng := range e.clusterConfig.NodeGroups {
		if ng.ScalingConfig != nil && ng.DesiredCapacity != nil {
			// Ensure min size allows for desired capacity
			if ng.MinSize != nil && *ng.MinSize == 0 {
				*ng.MinSize = *ng.DesiredCapacity
			}

			err = ngManager.Scale(ctx, ng.NodeGroupBase, true)
			if err != nil {
				return fmt.Errorf("failed to scale node group %s: %w", ng.Name, err)
			}
		}
	}

	return nil
}

// Stop stops an EKS cluster by scaling all node groups to 0.
func (e *EKSClusterProvisioner) Stop(name string) error {
	ctx := context.Background()

	ngManager, err := e.setupNodeGroupManager(ctx, name)
	if err != nil {
		return err
	}

	// Scale all node groups to 0
	for _, ng := range e.clusterConfig.NodeGroups {
		// Set desired capacity to 0 and min size to 0
		zeroSize := 0

		if ng.ScalingConfig == nil {
			ng.ScalingConfig = &v1alpha5.ScalingConfig{}
		}

		ng.DesiredCapacity = &zeroSize
		ng.MinSize = &zeroSize

		err = ngManager.Scale(ctx, ng.NodeGroupBase, true)
		if err != nil {
			return fmt.Errorf("failed to scale down node group %s: %w", ng.Name, err)
		}
	}

	return nil
}

// List lists all EKS clusters.
func (e *EKSClusterProvisioner) List() ([]string, error) {
	ctx := context.Background()

	ctl, err := e.createProvider(ctx)
	if err != nil {
		return nil, err
	}

	descriptions, err := e.clusterLister.GetClusters(ctx, ctl, false, DefaultChunkSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	clusterNames := make([]string, 0, len(descriptions))
	for _, desc := range descriptions {
		clusterNames = append(clusterNames, desc.Name)
	}

	return clusterNames, nil
}

// Exists checks if an EKS cluster exists.
func (e *EKSClusterProvisioner) Exists(name string) (bool, error) {
	clusters, err := e.List()
	if err != nil {
		return false, fmt.Errorf("failed to list clusters: %w", err)
	}

	target := setName(name, e.clusterConfig.Metadata.Name)

	return slices.Contains(clusters, target), nil
}

// setName returns name if non-empty, otherwise returns defaultName.
func setName(name, defaultName string) string {
	if name == "" {
		return defaultName
	}

	return name
}
