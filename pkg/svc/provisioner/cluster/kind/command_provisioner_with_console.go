// This file demonstrates Option 1 from docs/kind-console-logging-options.md
// It shows how to modify the POC to provide console logging like k3d.
//
//nolint:godoclint,revive,dupl // Demo file with expected duplication from POC
package kindprovisioner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	iopath "github.com/devantler-tech/ksail-go/pkg/io"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	createcluster "sigs.k8s.io/kind/pkg/cmd/kind/create/cluster"
	deletecluster "sigs.k8s.io/kind/pkg/cmd/kind/delete/cluster"
	getclusters "sigs.k8s.io/kind/pkg/cmd/kind/get/clusters"
	"sigs.k8s.io/kind/pkg/log"
)

// NOTE: This demonstrates Option 1: Using Kind's Cobra commands WITH console output.
// This provides the same real-time logging UX as k3d provisioner.
// See docs/kind-console-logging-options.md for details.

// ConsoleKindRunner executes kind commands and displays output to console.
type ConsoleKindRunner struct {
	stdout io.Writer
	stderr io.Writer
}

// NewConsoleKindRunner creates a command runner that displays output to console.
func NewConsoleKindRunner(stdout, stderr io.Writer) *ConsoleKindRunner {
	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	return &ConsoleKindRunner{
		stdout: stdout,
		stderr: stderr,
	}
}

// Run executes a kind Cobra command and displays output in real-time.
func (r *ConsoleKindRunner) Run(
	ctx context.Context,
	cmd *cobra.Command,
	args []string,
) (string, string, error) {
	var outBuf, errBuf bytes.Buffer

	// KEY DIFFERENCE: Use io.MultiWriter to display AND capture
	// This is how k3d shows console output to users
	cmd.SetOut(io.MultiWriter(&outBuf, r.stdout))
	cmd.SetErr(io.MultiWriter(&errBuf, r.stderr))

	cmd.SetContext(ctx)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	execErr := cmd.ExecuteContext(ctx)

	return outBuf.String(), errBuf.String(), execErr
}

// KindConsoleProvisioner is like KindCommandProvisionerPOC but displays console output.
// This demonstrates how to get the same user experience as k3d provisioner.
type KindConsoleProvisioner struct {
	kubeConfig string
	kindConfig *v1alpha4.Cluster
	client     client.ContainerAPIClient
	runner     KindCommandRunner
	builders   KindCommandBuilders
	provider   KindProvider
}

// NewKindConsoleProvisioner creates a provisioner that displays console output.
// This demonstrates Option 1 from docs/kind-console-logging-options.md.
func NewKindConsoleProvisioner(
	kindConfig *v1alpha4.Cluster,
	kubeConfig string,
	client client.ContainerAPIClient,
	provider KindProvider,
	stdout, stderr io.Writer,
	opts ...KindCommandProvisionerOption,
) *KindConsoleProvisioner {
	prov := &KindConsoleProvisioner{
		kubeConfig: kubeConfig,
		kindConfig: kindConfig,
		client:     client,
		provider:   provider,
		runner:     NewConsoleKindRunner(stdout, stderr),
		builders: KindCommandBuilders{
			Create: createcluster.NewCommand,
			Delete: deletecluster.NewCommand,
			List:   getclusters.NewCommand,
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt((*KindCommandProvisionerPOC)(prov))
		}
	}

	return prov
}

// Create creates a kind cluster and displays progress to console.
func (k *KindConsoleProvisioner) Create(ctx context.Context, name string) error {
	target := setName(name, k.kindConfig.Name)

	// Serialize config to temp file (limitation of Cobra commands)
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

	cmd := k.builders.Create(logger, streams)

	args := []string{"--name", target, "--config", tmpFile.Name()}

	// Console output will be displayed in real-time by ConsoleKindRunner
	_, stderr, err := k.runner.Run(ctx, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to create kind cluster: %w (stderr: %s)", err, stderr)
	}

	return nil
}

// Delete deletes a kind cluster and displays progress to console.
func (k *KindConsoleProvisioner) Delete(ctx context.Context, name string) error {
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

	// Console output will be displayed in real-time
	_, stderr, err := k.runner.Run(ctx, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w (stderr: %s)", err, stderr)
	}

	return nil
}

// Start starts a kind cluster (no Cobra command available).
func (k *KindConsoleProvisioner) Start(ctx context.Context, name string) error {
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

// Stop stops a kind cluster (no Cobra command available).
func (k *KindConsoleProvisioner) Stop(ctx context.Context, name string) error {
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

// List returns all kind clusters and displays output to console.
func (k *KindConsoleProvisioner) List(ctx context.Context) ([]string, error) {
	logger := log.NoopLogger{}

	var outBuf, errBuf bytes.Buffer

	streams := kindcmd.IOStreams{
		Out:    &outBuf,
		ErrOut: &errBuf,
	}

	cmd := k.builders.List(logger, streams)

	// Console output will be displayed in real-time
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
func (k *KindConsoleProvisioner) Exists(ctx context.Context, name string) (bool, error) {
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
