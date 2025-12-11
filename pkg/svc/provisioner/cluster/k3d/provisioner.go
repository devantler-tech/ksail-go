package k3dprovisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	clustercommand "github.com/k3d-io/k3d/v5/cmd/cluster"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	runner "github.com/devantler-tech/ksail-go/pkg/cmd/runner"
)

// CommandBuilders supplies constructors for k3d Cobra commands, allowing test injection.
type CommandBuilders struct {
	Create func() *cobra.Command
	Delete func() *cobra.Command
	Start  func() *cobra.Command
	Stop   func() *cobra.Command
	List   func() *cobra.Command
}

// Option configures the k3d command provisioner.
type Option func(*K3dClusterProvisioner)

// K3dClusterProvisioner executes k3d lifecycle commands via Cobra.
type K3dClusterProvisioner struct {
	simpleCfg  *v1alpha5.SimpleConfig
	configPath string
	runner     runner.CommandRunner
	builders   CommandBuilders
}

// NewK3dClusterProvisioner constructs a new command-backed provisioner.
func NewK3dClusterProvisioner(
	simpleCfg *v1alpha5.SimpleConfig,
	configPath string,
	opts ...Option,
) *K3dClusterProvisioner {
	// Configure logrus for k3d's console output
	// k3d uses logrus for logging, so we need to set it up properly
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:      true,
		DisableTimestamp: false,
		FullTimestamp:    false,
		TimestampFormat:  "2006-01-02T15:04:05Z",
	})
	logrus.SetLevel(logrus.InfoLevel)

	prov := &K3dClusterProvisioner{
		simpleCfg:  simpleCfg,
		configPath: configPath,
		runner:     runner.NewCobraCommandRunner(nil, nil),
		builders: CommandBuilders{
			Create: clustercommand.NewCmdClusterCreate,
			Delete: clustercommand.NewCmdClusterDelete,
			Start:  clustercommand.NewCmdClusterStart,
			Stop:   clustercommand.NewCmdClusterStop,
			List:   clustercommand.NewCmdClusterList,
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt(prov)
		}
	}

	return prov
}

// WithCommandRunner overrides the command runner (primarily for tests).
func WithCommandRunner(runner runner.CommandRunner) Option {
	return func(provisioner *K3dClusterProvisioner) {
		if runner != nil {
			provisioner.runner = runner
		}
	}
}

// WithCommandBuilders overrides specific Cobra command builders.
func WithCommandBuilders(builders CommandBuilders) Option {
	return func(provisioner *K3dClusterProvisioner) {
		if builders.Create != nil {
			provisioner.builders.Create = builders.Create
		}

		if builders.Delete != nil {
			provisioner.builders.Delete = builders.Delete
		}

		if builders.Start != nil {
			provisioner.builders.Start = builders.Start
		}

		if builders.Stop != nil {
			provisioner.builders.Stop = builders.Stop
		}

		if builders.List != nil {
			provisioner.builders.List = builders.List
		}
	}
}

// Create provisions a k3d cluster using the native Cobra command.
func (k *K3dClusterProvisioner) Create(ctx context.Context, name string) error {
	args := k.appendConfigFlag(nil)

	return k.runLifecycleCommand(
		ctx,
		k.builders.Create,
		args,
		name,
		"cluster create",
		func(target string) {
			if k.simpleCfg != nil {
				k.simpleCfg.Name = target
			}
		},
	)
}

// Delete removes a k3d cluster via the Cobra command.
func (k *K3dClusterProvisioner) Delete(ctx context.Context, name string) error {
	args := k.appendConfigFlag(nil)

	return k.runLifecycleCommand(ctx, k.builders.Delete, args, name, "cluster delete", nil)
}

// Start resumes a stopped k3d cluster via Cobra.
func (k *K3dClusterProvisioner) Start(ctx context.Context, name string) error {
	return k.runLifecycleCommand(ctx, k.builders.Start, nil, name, "cluster start", nil)
}

// Stop halts a running k3d cluster via Cobra.
func (k *K3dClusterProvisioner) Stop(ctx context.Context, name string) error {
	return k.runLifecycleCommand(ctx, k.builders.Stop, nil, name, "cluster stop", nil)
}

// List returns cluster names reported by the Cobra command.
func (k *K3dClusterProvisioner) List(ctx context.Context) ([]string, error) {
	cmd := k.builders.List()
	args := []string{"--output", "json"}

	res, err := k.runner.Run(ctx, cmd, args)
	if err != nil {
		return nil, fmt.Errorf("cluster list: %w", err)
	}

	output := strings.TrimSpace(res.Stdout)
	if output == "" {
		return nil, nil
	}

	var entries []struct {
		Name string `json:"name"`
	}

	decodeErr := json.Unmarshal([]byte(output), &entries)
	if decodeErr != nil {
		return nil, fmt.Errorf("cluster list: parse output: %w", decodeErr)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Name != "" {
			names = append(names, entry.Name)
		}
	}

	return names, nil
}

// Exists returns whether the target cluster is present.
func (k *K3dClusterProvisioner) Exists(ctx context.Context, name string) (bool, error) {
	clusters, err := k.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list: %w", err)
	}

	target := k.resolveName(name)
	if target == "" {
		return false, nil
	}

	return slices.Contains(clusters, target), nil
}

func (k *K3dClusterProvisioner) appendConfigFlag(args []string) []string {
	if k.configPath == "" {
		return args
	}

	return append(args, "--config", k.configPath)
}

func (k *K3dClusterProvisioner) resolveName(name string) string {
	if strings.TrimSpace(name) != "" {
		return name
	}

	if k.simpleCfg != nil && strings.TrimSpace(k.simpleCfg.Name) != "" {
		return k.simpleCfg.Name
	}

	return ""
}

func (k *K3dClusterProvisioner) runLifecycleCommand(
	ctx context.Context,
	builder func() *cobra.Command,
	args []string,
	name string,
	errorPrefix string,
	onTarget func(string),
) error {
	cmd := builder()

	target := k.resolveName(name)
	if target != "" {
		args = append(args, target)
		if onTarget != nil {
			onTarget(target)
		}
	}

	_, runErr := k.runner.Run(ctx, cmd, args)
	if runErr != nil {
		return fmt.Errorf("%s: %w", errorPrefix, runErr)
	}

	return nil
}
