package workload

import (
	"os"

	"github.com/devantler-tech/ksail-go/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// NewRolloutCmd creates the workload rollout command.
// The runtime parameter is kept for consistency with other workload command constructors,
// though it's currently unused as this command wraps kubectl directly.
func NewRolloutCmd(_ *runtime.Runtime) *cobra.Command {
	// Try to load config silently to get kubeconfig path
	kubeconfigPath := shared.GetKubeconfigPathSilently()

	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create kubectl client and get the rollout command directly
	client := kubectl.NewClient(ioStreams)
	rolloutCmd := client.CreateRolloutCommand(kubeconfigPath)

	return rolloutCmd
}
