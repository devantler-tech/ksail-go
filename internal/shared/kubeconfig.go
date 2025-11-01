package shared

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	iopath "github.com/devantler-tech/ksail-go/pkg/io"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

// GetDefaultKubeconfigPath returns the default kubeconfig path.
func GetDefaultKubeconfigPath() string {
	homeDir, _ := os.UserHomeDir()

	return filepath.Join(homeDir, ".kube", "config")
}

// GetKubeconfigPathSilently tries to load config and get kubeconfig path without any output.
func GetKubeconfigPathSilently() string {
	// Use io.Discard to suppress all output
	cfgManager := ksailconfigmanager.NewConfigManager(io.Discard)

	kubeconfigPath, err := getKubeconfigPath(cfgManager)
	if err != nil {
		// If we can't load config, use default kubeconfig
		return GetDefaultKubeconfigPath()
	}

	return kubeconfigPath
}

// getKubeconfigPath loads the ksail config and extracts the kubeconfig path.
func getKubeconfigPath(cfgManager *ksailconfigmanager.ConfigManager) (string, error) {
	// Create a minimal timer for config loading
	tmr := timer.New()
	tmr.Start()

	clusterCfg, err := cfgManager.LoadConfig(tmr)
	if err != nil {
		return "", fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	kubeconfigPath := clusterCfg.Spec.Connection.Kubeconfig
	if kubeconfigPath == "" {
		kubeconfigPath = GetDefaultKubeconfigPath()
	}

	// Expand home path
	expandedPath, err := iopath.ExpandHomePath(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to expand home path: %w", err)
	}

	return expandedPath, nil
}
