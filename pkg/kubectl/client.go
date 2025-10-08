// Package kubectl provides a kubectl client implementation.
package kubectl

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/get"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

// Client wraps kubectl command functionality.
type Client struct {
	ioStreams genericiooptions.IOStreams
}

// NewClient creates a new kubectl client instance.
func NewClient(ioStreams genericiooptions.IOStreams) *Client {
	return &Client{
		ioStreams: ioStreams,
	}
}

// CreateApplyCommand creates a kubectl apply command with all its flags and behavior.
func (c *Client) CreateApplyCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the apply command using kubectl's NewCmdApply
	applyCmd := apply.NewCmdApply("ksail", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	applyCmd.Use = "apply"
	applyCmd.Short = "Apply manifests"
	applyCmd.Long = "Apply local Kubernetes manifests to your cluster."

	return applyCmd
}

// CreateGetCommand creates a kubectl get command with all its flags and behavior.
func (c *Client) CreateGetCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the get command using kubectl's NewCmdGet
	getCmd := get.NewCmdGet("ksail", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	getCmd.Use = "get"
	getCmd.Short = "Get resources"
	getCmd.Long = "Display one or many Kubernetes resources from your cluster."

	return getCmd
}
