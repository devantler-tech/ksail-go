// Package kubectl provides a kubectl client implementation.
package kubectl

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/create"
	"k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/edit"
	"k8s.io/kubectl/pkg/cmd/explain"
	"k8s.io/kubectl/pkg/cmd/rollout"
	"k8s.io/kubectl/pkg/cmd/scale"
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

// CreateEditCommand creates a kubectl edit command with all its flags and behavior.
func (c *Client) CreateEditCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the edit command using kubectl's NewCmdEdit
	editCmd := edit.NewCmdEdit(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	editCmd.Use = "edit"
	editCmd.Short = "Edit a resource"
	editCmd.Long = "Edit a Kubernetes resource from the default editor."

	return editCmd
}

// CreateDeleteCommand creates a kubectl delete command with all its flags and behavior.
func (c *Client) CreateDeleteCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the delete command using kubectl's NewCmdDelete
	deleteCmd := delete.NewCmdDelete(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	deleteCmd.Use = "delete"
	deleteCmd.Short = "Delete resources"
	deleteCmd.Long = "Delete Kubernetes resources by file names, stdin, resources and names, " +
		"or by resources and label selector."

	return deleteCmd
}

// CreateCreateCommand creates a kubectl create command with all its flags and behavior.
func (c *Client) CreateCreateCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the create command using kubectl's NewCmdCreate
	createCmd := create.NewCmdCreate(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	createCmd.Use = "create"
	createCmd.Short = "Create resources"
	createCmd.Long = "Create Kubernetes resources from files or stdin."

	return createCmd
}

// CreateExplainCommand creates a kubectl explain command with all its flags and behavior.
func (c *Client) CreateExplainCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the explain command using kubectl's NewCmdExplain
	explainCmd := explain.NewCmdExplain("ksail", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	explainCmd.Use = "explain"
	explainCmd.Short = "Get documentation for a resource"
	explainCmd.Long = "Get documentation for Kubernetes resources, including field descriptions and structure."

	return explainCmd
}

// CreateRolloutCommand creates a kubectl rollout command with all its flags and behavior.
func (c *Client) CreateRolloutCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the rollout command using kubectl's NewCmdRollout
	rolloutCmd := rollout.NewCmdRollout(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	rolloutCmd.Use = "rollout"
	rolloutCmd.Short = "Manage the rollout of a resource"
	rolloutCmd.Long = "Manage the rollout of one or many resources."

	return rolloutCmd
}

// CreateScaleCommand creates a kubectl scale command with all its flags and behavior.
func (c *Client) CreateScaleCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the scale command using kubectl's NewCmdScale
	scaleCmd := scale.NewCmdScale(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	scaleCmd.Use = "scale"
	scaleCmd.Short = "Scale resources"
	scaleCmd.Long = "Set a new size for a deployment, replica set, replication controller, or stateful set."

	return scaleCmd
}
