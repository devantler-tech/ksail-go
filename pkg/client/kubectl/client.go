// Package kubectl provides a kubectl client implementation.
package kubectl

import (
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd/apply"
	"k8s.io/kubectl/pkg/cmd/clusterinfo"
	"k8s.io/kubectl/pkg/cmd/create"
	"k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/describe"
	"k8s.io/kubectl/pkg/cmd/edit"
	"k8s.io/kubectl/pkg/cmd/exec"
	"k8s.io/kubectl/pkg/cmd/explain"
	"k8s.io/kubectl/pkg/cmd/expose"
	"k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/cmd/logs"
	"k8s.io/kubectl/pkg/cmd/rollout"
	"k8s.io/kubectl/pkg/cmd/scale"
	"k8s.io/kubectl/pkg/cmd/wait"
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

// replaceKubectlInExamples replaces "kubectl" with "ksail workload" in command examples.
func replaceKubectlInExamples(cmd *cobra.Command) {
	if cmd.Example != "" {
		cmd.Example = strings.ReplaceAll(cmd.Example, "kubectl", "ksail workload")
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
	applyCmd := apply.NewCmdApply("ksail workload", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	applyCmd.Use = "apply"
	applyCmd.Short = "Apply manifests"
	applyCmd.Long = "Apply local Kubernetes manifests to your cluster."
	replaceKubectlInExamples(applyCmd)

	return applyCmd
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
	replaceKubectlInExamples(createCmd)

	return createCmd
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
	replaceKubectlInExamples(editCmd)

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
	replaceKubectlInExamples(deleteCmd)

	return deleteCmd
}

// CreateDescribeCommand creates a kubectl describe command with all its flags and behavior.
func (c *Client) CreateDescribeCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	// Create the describe command using kubectl's NewCmdDescribe
	describeCmd := describe.NewCmdDescribe("ksail workload", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	describeCmd.Use = "describe"
	describeCmd.Short = "Describe resources"
	describeCmd.Long = "Show details of a specific resource or group of resources."
	replaceKubectlInExamples(describeCmd)

	return describeCmd
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
	explainCmd := explain.NewCmdExplain("ksail workload", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	explainCmd.Use = "explain"
	explainCmd.Short = "Get documentation for a resource"
	explainCmd.Long = "Get documentation for Kubernetes resources, including field descriptions and structure."
	replaceKubectlInExamples(explainCmd)

	return explainCmd
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
	getCmd := get.NewCmdGet("ksail workload", factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	getCmd.Use = "get"
	getCmd.Short = "Get resources"
	getCmd.Long = "Display one or many Kubernetes resources from your cluster."
	replaceKubectlInExamples(getCmd)

	return getCmd
}

// CreateLogsCommand creates a kubectl logs command with all its flags and behavior.
func (c *Client) CreateLogsCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the logs command using kubectl's NewCmdLogs
	logsCmd := logs.NewCmdLogs(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	logsCmd.Use = "logs"
	logsCmd.Short = "Print container logs"
	logsCmd.Long = "Print the logs for a container in a pod or specified resource. " +
		"If the pod has only one container, the container name is optional."
	replaceKubectlInExamples(logsCmd)

	return logsCmd
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
	replaceKubectlInExamples(rolloutCmd)

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
	replaceKubectlInExamples(scaleCmd)

	return scaleCmd
}

// CreateExposeCommand creates a kubectl expose command with all its flags and behavior.
func (c *Client) CreateExposeCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the expose command using kubectl's NewCmdExposeService
	exposeCmd := expose.NewCmdExposeService(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	exposeCmd.Use = "expose"
	exposeCmd.Short = "Expose a resource as a service"
	exposeCmd.Long = "Expose a resource as a new Kubernetes service."
	replaceKubectlInExamples(exposeCmd)

	return exposeCmd
}

// CreateClusterInfoCommand creates a kubectl cluster-info command with all its flags and behavior.
func (c *Client) CreateClusterInfoCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create the cluster-info command using kubectl's NewCmdClusterInfo
	clusterInfoCmd := clusterinfo.NewCmdClusterInfo(configFlags, c.ioStreams)

	// Customize command metadata to fit ksail context
	clusterInfoCmd.Use = "info"
	clusterInfoCmd.Short = "Display cluster information"
	clusterInfoCmd.Long = "Display addresses of the control plane and services with label " +
		"kubernetes.io/cluster-service=true."

	return clusterInfoCmd
}

// CreateExecCommand creates a kubectl exec command with all its flags and behavior.
func (c *Client) CreateExecCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create factory for kubectl command
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	// Create the exec command using kubectl's NewCmdExec
	execCmd := exec.NewCmdExec(factory, c.ioStreams)

	// Customize command metadata to fit ksail context
	execCmd.Use = "exec"
	execCmd.Short = "Execute a command in a container"
	execCmd.Long = "Execute a command in a container in a pod."
	replaceKubectlInExamples(execCmd)

	return execCmd
}

// CreateWaitCommand creates a kubectl wait command with all its flags and behavior.
func (c *Client) CreateWaitCommand(kubeConfigPath string) *cobra.Command {
	// Create config flags with kubeconfig path
	configFlags := genericclioptions.NewConfigFlags(true)
	if kubeConfigPath != "" {
		configFlags.KubeConfig = &kubeConfigPath
	}

	// Create the wait command using kubectl's NewCmdWait
	waitCmd := wait.NewCmdWait(configFlags, c.ioStreams)

	// Customize command metadata to fit ksail context
	waitCmd.Use = "wait"
	waitCmd.Short = "Wait for a specific condition on one or many resources"
	waitCmd.Long = "Wait for a specific condition on one or many resources. " +
		"The command takes multiple resources and waits until the specified condition " +
		"is seen in the Status field of every given resource."
	replaceKubectlInExamples(waitCmd)

	return waitCmd
}
