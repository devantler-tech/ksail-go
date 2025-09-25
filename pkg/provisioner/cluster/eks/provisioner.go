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

// ErrInvalidClusterConfig is returned when cluster configuration is invalid.
var ErrInvalidClusterConfig = errors.New("cluster configuration or metadata is nil")

// ErrEmptyClusterName is returned when cluster name is empty.
var ErrEmptyClusterName = errors.New("cluster name cannot be empty")

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
	clusterConfig    *v1alpha5.ClusterConfig
	clusterProvider  *eks.ClusterProvider
	clusterActions   EKSClusterActions
	clusterLister    EKSClusterLister
	clusterCreator   EKSClusterCreator
	nodeGroupManager EKSNodeGroupManager
}

// NewEKSClusterProvisioner constructs an EKSClusterProvisioner with explicit dependencies
// for the eksctl provider and cluster actions. This supports both production wiring
// and unit testing via mocks.
func NewEKSClusterProvisioner(
	clusterConfig *v1alpha5.ClusterConfig,
	clusterProvider *eks.ClusterProvider,
	clusterActions EKSClusterActions,
	clusterLister EKSClusterLister,
	clusterCreator EKSClusterCreator,
	nodeGroupManager EKSNodeGroupManager,
) *EKSClusterProvisioner {
	return &EKSClusterProvisioner{
		clusterConfig:    clusterConfig,
		clusterProvider:  clusterProvider,
		clusterActions:   clusterActions,
		clusterLister:    clusterLister,
		clusterCreator:   clusterCreator,
		nodeGroupManager: nodeGroupManager,
	}
}

// Create creates an EKS cluster.
func (e *EKSClusterProvisioner) Create(ctx context.Context, name string) error {
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
func (e *EKSClusterProvisioner) Delete(ctx context.Context, name string) error {
	_, err := e.setupClusterOperation(ctx, name)
	if err != nil {
		return err
	}

	// Use reasonable defaults for deletion parameters
	waitInterval := DefaultWaitInterval
	podEvictionWaitPeriod := DefaultPodEvictionWaitPeriod
	wait := true
	force := false
	disableNodegroupEviction := false
	parallel := DefaultParallelism

	err = e.clusterActions.Delete(
		ctx, waitInterval, podEvictionWaitPeriod, wait, force, disableNodegroupEviction, parallel,
	)
	if err != nil {
		return fmt.Errorf("failed to delete EKS cluster: %w", err)
	}

	return nil
}

// Start starts an EKS cluster by scaling node groups from 0 to their desired capacity.
func (e *EKSClusterProvisioner) Start(ctx context.Context, name string) error {
	ngManager, err := e.setupNodeGroupManager(ctx, name)
	if err != nil {
		return err
	}

	// Scale all node groups to their desired capacity
	for _, nodeGroup := range e.clusterConfig.NodeGroups {
		if nodeGroup.ScalingConfig != nil && nodeGroup.DesiredCapacity != nil {
			// Ensure min size allows for desired capacity
			if nodeGroup.MinSize != nil && *nodeGroup.MinSize == 0 {
				*nodeGroup.MinSize = *nodeGroup.DesiredCapacity
			}

			err = ngManager.Scale(ctx, nodeGroup.NodeGroupBase, true)
			if err != nil {
				return fmt.Errorf("failed to scale node group %s: %w", nodeGroup.Name, err)
			}
		}
	}

	return nil
}

// Stop stops an EKS cluster by scaling all node groups to 0.
func (e *EKSClusterProvisioner) Stop(ctx context.Context, name string) error {
	ngManager, err := e.setupNodeGroupManager(ctx, name)
	if err != nil {
		return err
	}

	// Scale all node groups to 0
	for _, nodeGroup := range e.clusterConfig.NodeGroups {
		// Set desired capacity to 0 and min size to 0
		zeroSize := 0

		if nodeGroup.ScalingConfig == nil {
			nodeGroup.ScalingConfig = &v1alpha5.ScalingConfig{
				DesiredCapacity: nil,
				MinSize:         nil,
				MaxSize:         nil,
			}
		}

		nodeGroup.DesiredCapacity = &zeroSize
		nodeGroup.MinSize = &zeroSize

		err = ngManager.Scale(ctx, nodeGroup.NodeGroupBase, true)
		if err != nil {
			return fmt.Errorf("failed to scale down node group %s: %w", nodeGroup.Name, err)
		}
	}

	return nil
}

// List lists all EKS clusters.
func (e *EKSClusterProvisioner) List(ctx context.Context) ([]string, error) {
	descriptions, err := e.clusterLister.GetClusters(
		ctx,
		e.clusterProvider,
		false,
		DefaultChunkSize,
	)
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
func (e *EKSClusterProvisioner) Exists(ctx context.Context, name string) (bool, error) {
	clusters, err := e.List(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list clusters: %w", err)
	}

	target := e.getEffectiveClusterName(name)

	return slices.Contains(clusters, target), nil
}

// getEffectiveClusterName determines the effective cluster name to use.
// Returns the provided name if not empty, otherwise returns the name from cluster config.
// Returns empty string if no name is available from either source.
func (e *EKSClusterProvisioner) getEffectiveClusterName(name string) string {
	if name != "" {
		return name
	}

	if e.clusterConfig != nil && e.clusterConfig.Metadata != nil && e.clusterConfig.Metadata.Name != "" {
		return e.clusterConfig.Metadata.Name
	}

	// Return default cluster name if none is available
	return "ksail-default"
}

// setupClusterOperation sets up common cluster operation prerequisites.
func (e *EKSClusterProvisioner) setupClusterOperation(
	_ context.Context,
	name string,
) (*eks.ClusterProvider, error) {
	if e.clusterConfig == nil || e.clusterConfig.Metadata == nil {
		return nil, ErrInvalidClusterConfig
	}

	effectiveName := e.getEffectiveClusterName(name)
	if effectiveName == "" {
		return nil, ErrEmptyClusterName
	}

	// Update the cluster config with the effective name
	e.clusterConfig.Metadata.Name = effectiveName

	return e.clusterProvider, nil
}

// ensureClusterExists checks if a cluster exists and returns ErrClusterNotFound if not.
func (e *EKSClusterProvisioner) ensureClusterExists(ctx context.Context, name string) error {
	exists, err := e.Exists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check if cluster exists: %w", err)
	}

	if !exists {
		return ErrClusterNotFound
	}

	return nil
}

// setupNodeGroupManager sets up common node group management prerequisites.
//
//nolint:ireturn // Returning interface is intended design for dependency injection
func (e *EKSClusterProvisioner) setupNodeGroupManager(
	ctx context.Context,
	name string,
) (EKSNodeGroupManager, error) {
	err := e.ensureClusterExists(ctx, name)
	if err != nil {
		return nil, err
	}

	_, err = e.setupClusterOperation(ctx, name)
	if err != nil {
		return nil, err
	}

	return e.nodeGroupManager, nil
}
