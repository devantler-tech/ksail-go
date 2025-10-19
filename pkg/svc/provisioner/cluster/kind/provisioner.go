// Package kindprovisioner provides implementations of the Provisioner interface
// for provisioning clusters in different providers.
package kindprovisioner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	iopath "github.com/devantler-tech/ksail-go/pkg/io"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	commandrunner "github.com/devantler-tech/ksail-go/pkg/svc/commandrunner"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	createcluster "sigs.k8s.io/kind/pkg/cmd/kind/create/cluster"
	deletecluster "sigs.k8s.io/kind/pkg/cmd/kind/delete/cluster"
	getclusters "sigs.k8s.io/kind/pkg/cmd/kind/get/clusters"
	"sigs.k8s.io/kind/pkg/log"
)

// ErrClusterNotFound is returned when a cluster is not found.
var ErrClusterNotFound = errors.New("cluster not found")

// KindProvider describes the subset of methods from kind's Provider used here.
type KindProvider interface {
	Create(name string, opts ...cluster.CreateOption) error
	Delete(name, kubeconfigPath string) error
	List() ([]string, error)
	ListNodes(name string) ([]string, error)
}

// KindClusterProvisioner is an implementation of the ClusterProvisioner interface for provisioning kind clusters.
// It uses kind's Cobra commands where available (create, delete, list) and falls back to
// Docker client for operations not available as Cobra commands (start, stop).
type KindClusterProvisioner struct {
	kubeConfig string
	kindConfig *v1alpha4.Cluster
	provider   KindProvider
	client     client.ContainerAPIClient
	runner     commandrunner.CommandRunner
}

// NewKindClusterProvisioner constructs a KindClusterProvisioner with explicit dependencies
// for the kind provider and docker client. This supports both production wiring
// and unit testing via mocks.
func NewKindClusterProvisioner(
	kindConfig *v1alpha4.Cluster,
	kubeConfig string,
	provider KindProvider,
	client client.ContainerAPIClient,
) *KindClusterProvisioner {
	return NewKindClusterProvisionerWithRunner(
		kindConfig,
		kubeConfig,
		provider,
		client,
		commandrunner.NewGenericCobraCommandRunner(os.Stdout, os.Stderr),
	)
}

// NewKindClusterProvisionerWithRunner constructs a KindClusterProvisioner with
// an explicit command runner for testing purposes.
func NewKindClusterProvisionerWithRunner(
	kindConfig *v1alpha4.Cluster,
	kubeConfig string,
	provider KindProvider,
	client client.ContainerAPIClient,
	runner commandrunner.CommandRunner,
) *KindClusterProvisioner {
	return &KindClusterProvisioner{
		kubeConfig: kubeConfig,
		kindConfig: kindConfig,
		provider:   provider,
		client:     client,
		runner:     runner,
	}
}

// Create creates a kind cluster using kind's Cobra command.
func (k *KindClusterProvisioner) Create(ctx context.Context, name string) error {
	target := setName(name, k.kindConfig.Name)

	// Serialize config to temp file (required by kind's Cobra command)
	tmpFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp config file: %w", err)
	}

	defer func() { _ = os.Remove(tmpFile.Name()) }()

	marshaller := yamlmarshaller.NewMarshaller[*v1alpha4.Cluster]()

	configYAML, err := marshaller.Marshal(k.kindConfig)
	if err != nil {
		return fmt.Errorf("marshal kind config: %w", err)
	}

	const configFilePerms = 0o600

	err = os.WriteFile(tmpFile.Name(), []byte(configYAML), configFilePerms)
	if err != nil {
		return fmt.Errorf("write temp config file: %w", err)
	}

	logger := log.NoopLogger{}

	var outBuf, errBuf bytes.Buffer

	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := createcluster.NewCommand(logger, streams)

	args := []string{"--name", target, "--config", tmpFile.Name()}

	_, err = k.runner.Run(ctx, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to create kind cluster: %w", err)
	}

	return nil
}

// Delete deletes a kind cluster using kind's Cobra command.
func (k *KindClusterProvisioner) Delete(ctx context.Context, name string) error {
	target := setName(name, k.kindConfig.Name)

	kubeconfigPath, _ := iopath.ExpandHomePath(k.kubeConfig)

	logger := log.NoopLogger{}

	var outBuf, errBuf bytes.Buffer

	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := deletecluster.NewCommand(logger, streams)

	args := []string{"--name", target}
	if kubeconfigPath != "" {
		args = append(args, "--kubeconfig", kubeconfigPath)
	}

	_, err := k.runner.Run(ctx, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w", err)
	}

	return nil
}

// Start starts a kind cluster.
// Note: kind does not provide a Cobra command for start, so we use Docker client directly.
func (k *KindClusterProvisioner) Start(ctx context.Context, name string) error {
	const dockerStartTimeout = 30 * time.Second

	target := setName(name, k.kindConfig.Name)

	nodes, err := k.provider.ListNodes(target)
	if err != nil {
		return fmt.Errorf("cluster '%s': %w", target, err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("%w", ErrClusterNotFound)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, dockerStartTimeout)
	defer cancel()

	for _, name := range nodes {
		// Start each node container by name using Docker SDK
		err := k.client.ContainerStart(timeoutCtx, name, container.StartOptions{
			CheckpointID:  "",
			CheckpointDir: "",
		})
		if err != nil {
			return fmt.Errorf("docker start failed for %s: %w", name, err)
		}
	}

	return nil
}

// Stop stops a kind cluster.
// Note: kind does not provide a Cobra command for stop, so we use Docker client directly.
func (k *KindClusterProvisioner) Stop(ctx context.Context, name string) error {
	const dockerStopTimeout = 60 * time.Second

	target := setName(name, k.kindConfig.Name)

	nodes, err := k.provider.ListNodes(target)
	if err != nil {
		return fmt.Errorf("failed to list nodes for cluster '%s': %w", target, err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("%w", ErrClusterNotFound)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, dockerStopTimeout)
	defer cancel()

	for _, name := range nodes {
		// Stop each node container by name using Docker SDK
		// Graceful stop with default timeout
		err := k.client.ContainerStop(timeoutCtx, name, container.StopOptions{
			Signal:  "",
			Timeout: nil,
		})
		if err != nil {
			return fmt.Errorf("docker stop failed for %s: %w", name, err)
		}
	}

	return nil
}

// List returns all kind clusters using kind's Cobra command.
func (k *KindClusterProvisioner) List(ctx context.Context) ([]string, error) {
	logger := log.NoopLogger{}

	var outBuf, errBuf bytes.Buffer

	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := getclusters.NewCommand(logger, streams)

	result, err := k.runner.Run(ctx, cmd, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	const noKindClustersMsg = "No kind clusters found."

	// Parse stdout - each line is a cluster name
	lines := bytes.Split([]byte(result.Stdout), []byte("\n"))

	var clusters []string

	for _, line := range lines {
		name := string(bytes.TrimSpace(line))
		if name != "" && name != noKindClustersMsg {
			clusters = append(clusters, name)
		}
	}

	return clusters, nil
}

// Exists checks if a kind cluster exists.
func (k *KindClusterProvisioner) Exists(ctx context.Context, name string) (bool, error) {
	clusters, err := k.List(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	target := setName(name, k.kindConfig.Name)

	if slices.Contains(clusters, target) {
		return true, nil
	}

	return false, nil
}

// --- internals ---

func setName(name string, kindConfigName string) string {
	target := name
	if target == "" {
		target = kindConfigName
	}

	return target
}
