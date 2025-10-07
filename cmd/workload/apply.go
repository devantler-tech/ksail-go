package workload

import (
	"fmt"
	"io"
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
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := getKubeconfigPathSilently()

	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create applier and get the kubectl apply command directly
	applier := kubectlapplier.NewApplier(ioStreams)
	applyCmd := applier.CreateApplyCommand(kubeconfigPath)

	return applyCmd
}

// getKubeconfigPathSilently tries to load config and get kubeconfig path without any output.
func getKubeconfigPathSilently() string {
	// Use io.Discard to suppress all output
	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)

	kubeconfigPath, err := getKubeconfigPath(cfgManager)
	if err != nil {
		// If we can't load config, use default kubeconfig
		homeDir, _ := os.UserHomeDir()

		return filepath.Join(homeDir, ".kube", "config")
	}

	return kubeconfigPath
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
