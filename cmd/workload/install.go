package workload

import (
	"os"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/helm"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := getKubeconfigPathSilently()

	// Create helm client and get the install command directly
	client := helm.NewClient(os.Stdout, os.Stderr, kubeconfigPath)
	installCmd := client.CreateInstallCommand()

	return installCmd
}
