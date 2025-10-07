// Package kubectlapplier provides a kubectl applier implementation.
package kubectlapplier

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Applier wraps kubectl apply command functionality.
type Applier struct {
	ioStreams genericiooptions.IOStreams
}

// NewApplier creates a new kubectl applier instance.
func NewApplier(ioStreams genericiooptions.IOStreams) *Applier {
	return &Applier{
		ioStreams: ioStreams,
	}
}

// CreateApplyCommand creates a kubectl apply command with all its flags and behavior.
func (a *Applier) CreateApplyCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the apply command using kubectl's NewCmdApply
	applyCmd := apply.NewCmdApply("ksail", factory, a.ioStreams)

	// Customize command metadata to fit ksail context
	applyCmd.Use = "apply"
	applyCmd.Short = "Apply manifests"
	applyCmd.Long = "Apply local Kubernetes manifests to your cluster."

	return applyCmd
}
