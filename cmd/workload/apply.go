package workload

import (
	"fmt"
	"os"
	"path/filepath"

	kubectlapplier "github.com/devantler-tech/ksail-go/pkg/applier/kubectl"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	iopath "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewApplyCmd creates the workload apply command.
func NewApplyCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a wrapper command that will lazily create the kubectl apply command
	cmd := &cobra.Command{
		Use:                   "apply (-f FILENAME | -k DIRECTORY)",
		DisableFlagsInUseLine: true,
		Short:                 "Apply manifests",
		Long:                  "Apply local Kubernetes manifests to your cluster.",
		SilenceUsage:          true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config and get kubeconfig path
			cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout())

			kubeconfigPath, err := getKubeconfigPath(cfgManager)
			if err != nil {
				// If we can't load config, use default kubeconfig
				homeDir, _ := os.UserHomeDir()
				kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
			}

			// Create IO streams for kubectl
			ioStreams := genericiooptions.IOStreams{
				In:     cmd.InOrStdin(),
				Out:    cmd.OutOrStdout(),
				ErrOut: cmd.ErrOrStderr(),
			}

			// Create applier and get the kubectl apply command
			applier := kubectlapplier.NewApplier(ioStreams)
			applyCmd := applier.CreateApplyCommand(kubeconfigPath)

			// Transfer flags from parent command to kubectl apply command
			applyCmd.SetArgs(args)
			applyCmd.SetIn(cmd.InOrStdin())
			applyCmd.SetOut(cmd.OutOrStdout())
			applyCmd.SetErr(cmd.ErrOrStderr())

			// Execute the kubectl apply command
			return applyCmd.Execute()
		},
	}

	// Add kubectl apply flags by creating a temporary apply command
	// This ensures help shows correct flags even before execution
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	applier := kubectlapplier.NewApplier(ioStreams)
	tempApplyCmd := applier.CreateApplyCommand("")

	// Copy flags from temporary kubectl apply command
	cmd.Flags().AddFlagSet(tempApplyCmd.Flags())

	return cmd
}

// getKubeconfigPath loads the ksail config and extracts the kubeconfig path.
func getKubeconfigPath(cfgManager *ksailconfigmanager.ConfigManager) (string, error) {
	// Create a minimal timer for config loading
	tmr := timer.New()
	tmr.Start()

	err := cfgManager.LoadConfig(tmr)
	if err != nil {
		return "", fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	clusterCfg := cfgManager.GetConfig()

	kubeconfigPath := clusterCfg.Spec.Connection.Kubeconfig
	if kubeconfigPath == "" {
		homeDir, _ := os.UserHomeDir()
		kubeconfigPath = filepath.Join(homeDir, ".kube", "config")
	}

	// Expand home path
	expandedPath, err := iopath.ExpandHomePath(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to expand home path: %w", err)
	}

	return expandedPath, nil
}
