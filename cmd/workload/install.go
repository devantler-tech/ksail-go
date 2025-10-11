package workload

import (
	"os"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := shared.GetKubeconfigPathSilently()

	// Create helm client and get the install command directly
	client := helm.NewClient(os.Stdout, os.Stderr, kubeconfigPath)
	installCmd := client.CreateInstallCommand()

	return installCmd
}
