// Package kindprovisioner provides implementations of the Provisioner interface
// for provisioning clusters in different providers.
package kindprovisioner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"slices"
	"time"

	iopath "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	createcluster "sigs.k8s.io/kind/pkg/cmd/kind/create/cluster"
	deletecluster "sigs.k8s.io/kind/pkg/cmd/kind/delete/cluster"
	getclusters "sigs.k8s.io/kind/pkg/cmd/kind/get/clusters"
	"sigs.k8s.io/kind/pkg/log"
)

// NOTE: This is a PROOF-OF-CONCEPT implementation demonstrating that it's TECHNICALLY POSSIBLE
// to use kind's Cobra commands, but it is NOT RECOMMENDED for production use.
// See docs/kind-cobra-analysis.md for detailed reasoning.

// KindCommandRunner executes Cobra commands with kind-specific requirements.
type KindCommandRunner interface {
	Run(ctx context.Context, cmd *cobra.Command, args []string) (stdout, stderr string, err error)
}

// SimpleKindRunner is a basic command runner for kind commands.
type SimpleKindRunner struct{}

// NewSimpleKindRunner creates a new simple kind command runner.
func NewSimpleKindRunner() *SimpleKindRunner {
	return &SimpleKindRunner{}
}

// Run executes a kind Cobra command and returns stdout, stderr, and any error.
func (r *SimpleKindRunner) Run(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
) (stdout, stderr string, err error) {
	var outBuf, errBuf bytes.Buffer

	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	execErr := cmd.ExecuteContext(ctx)

	return outBuf.String(), errBuf.String(), execErr
}

// KindCommandBuilders supplies constructors for kind Cobra commands.
type KindCommandBuilders struct {
	Create func(logger log.Logger, streams kindcmd.IOStreams) *cobra.Command
	Delete func(logger log.Logger, streams kindcmd.IOStreams) *cobra.Command
	List   func(logger log.Logger, streams kindcmd.IOStreams) *cobra.Command
}

// KindCommandProvisionerOption configures the kind command provisioner.
type KindCommandProvisionerOption func(*KindCommandProvisionerPOC)

// KindCommandProvisionerPOC is a PROOF-OF-CONCEPT provisioner using kind's Cobra commands.
// This demonstrates it's technically possible but NOT RECOMMENDED.
// See docs/kind-cobra-analysis.md for why the current Provider-based approach is better.
type KindCommandProvisionerPOC struct {
	kubeConfig string
	kindConfig *v1alpha4.Cluster
	client     client.ContainerAPIClient
	runner     KindCommandRunner
	builders   KindCommandBuilders
	provider   KindProvider // Still needed for ListNodes in Start/Stop
}

// NewKindCommandProvisionerPOC creates a proof-of-concept command-based provisioner.
// WARNING: This is for demonstration only. Use NewKindClusterProvisioner for production.
func NewKindCommandProvisionerPOC(
	kindConfig *v1alpha4.Cluster,
	kubeConfig string,
	client client.ContainerAPIClient,
	provider KindProvider,
	opts ...KindCommandProvisionerOption,
) *KindCommandProvisionerPOC {
	prov := &KindCommandProvisionerPOC{
		kubeConfig: kubeConfig,
		kindConfig: kindConfig,
		client:     client,
		provider:   provider,
		runner:     NewSimpleKindRunner(),
		builders: KindCommandBuilders{
			Create: createcluster.NewCommand,
			Delete: deletecluster.NewCommand,
			List:   getclusters.NewCommand,
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(prov)
		}
	}

	return prov
}

// WithKindCommandRunner overrides the command runner.
func WithKindCommandRunner(runner KindCommandRunner) KindCommandProvisionerOption {
	return func(p *KindCommandProvisionerPOC) {
		if runner != nil {
			p.runner = runner
		}
	}
}

// WithKindCommandBuilders overrides command builders.
func WithKindCommandBuilders(builders KindCommandBuilders) KindCommandProvisionerOption {
	return func(p *KindCommandProvisionerPOC) {
		if builders.Create != nil {
			p.builders.Create = builders.Create
		}
		if builders.Delete != nil {
			p.builders.Delete = builders.Delete
		}
		if builders.List != nil {
			p.builders.List = builders.List
		}
	}
}

// Create creates a kind cluster using the Cobra command.
// NOTE: This requires serializing config to a temp file - a significant limitation.
func (k *KindCommandProvisionerPOC) Create(ctx context.Context, name string) error {
	target := setName(name, k.kindConfig.Name)

	// LIMITATION: Must serialize config to temp file for Cobra command
	tmpFile, err := os.CreateTemp("", "kind-config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp config file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	configYAML, err := yaml.Marshal(k.kindConfig)
	if err != nil {
		return fmt.Errorf("marshal kind config: %w", err)
	}

	if err := os.WriteFile(tmpFile.Name(), configYAML, 0600); err != nil {
		return fmt.Errorf("write temp config file: %w", err)
	}

	logger := log.NoopLogger{}
	var outBuf, errBuf bytes.Buffer
	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := k.builders.Create(logger, streams)

	args := []string{"--name", target, "--config", tmpFile.Name()}

	_, stderr, err := k.runner.Run(ctx, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to create kind cluster: %w (stderr: %s)", err, stderr)
	}

	return nil
}

// Delete deletes a kind cluster using the Cobra command.
func (k *KindCommandProvisionerPOC) Delete(ctx context.Context, name string) error {
	target := setName(name, k.kindConfig.Name)

	kubeconfigPath, _ := iopath.ExpandHomePath(k.kubeConfig)

	logger := log.NoopLogger{}
	var outBuf, errBuf bytes.Buffer
	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := k.builders.Delete(logger, streams)

	args := []string{"--name", target}
	if kubeconfigPath != "" {
		args = append(args, "--kubeconfig", kubeconfigPath)
	}

	_, stderr, err := k.runner.Run(ctx, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w (stderr: %s)", err, stderr)
	}

	return nil
}

// Start starts a kind cluster.
// NOTE: No Cobra command exists - must use Provider interface + Docker client.
func (k *KindCommandProvisionerPOC) Start(ctx context.Context, name string) error {
	const dockerStartTimeout = 30 * time.Second

	target := setName(name, k.kindConfig.Name)

	// LIMITATION: Must use Provider interface - no Cobra command for ListNodes
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
// NOTE: No Cobra command exists - must use Provider interface + Docker client.
func (k *KindCommandProvisionerPOC) Stop(ctx context.Context, name string) error {
	const dockerStopTimeout = 60 * time.Second

	target := setName(name, k.kindConfig.Name)

	// LIMITATION: Must use Provider interface - no Cobra command for ListNodes
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

// List returns all kind clusters using the Cobra command.
func (k *KindCommandProvisionerPOC) List(ctx context.Context) ([]string, error) {
	logger := log.NoopLogger{}
	var outBuf, errBuf bytes.Buffer
	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := k.builders.List(logger, streams)

	stdout, stderr, err := k.runner.Run(ctx, cmd, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w (stderr: %s)", err, stderr)
	}

	// Parse stdout - each line is a cluster name
	lines := bytes.Split([]byte(stdout), []byte("\n"))
	var clusters []string
	for _, line := range lines {
		name := string(bytes.TrimSpace(line))
		if name != "" && name != "No kind clusters found." {
			clusters = append(clusters, name)
		}
	}

	return clusters, nil
}

// Exists checks if a kind cluster exists.
func (k *KindCommandProvisionerPOC) Exists(ctx context.Context, name string) (bool, error) {
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
