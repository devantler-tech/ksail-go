// Package kubectl provides a kubectl client implementation.
package kubectl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	// ErrResourceCommandNotFound is returned when a kubectl create subcommand is not found.
	ErrResourceCommandNotFound = errors.New("kubectl create command not found for resource type")
	// ErrNoRunFunction is returned when a kubectl command has neither RunE nor Run function.
	ErrNoRunFunction = errors.New("no run function found for kubectl create command")
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

// NewNamespaceCmd creates a command to generate a Namespace manifest.
func (c *Client) NewNamespaceCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("namespace")
	if err != nil {
		panic(fmt.Sprintf("failed to create namespace command: %v", err))
	}

	return cmd
}

// NewConfigMapCmd creates a command to generate a ConfigMap manifest.
func (c *Client) NewConfigMapCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("configmap")
	if err != nil {
		panic(fmt.Sprintf("failed to create configmap command: %v", err))
	}

	return cmd
}

// NewSecretCmd creates a command to generate a Secret manifest.
func (c *Client) NewSecretCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("secret")
	if err != nil {
		panic(fmt.Sprintf("failed to create secret command: %v", err))
	}

	return cmd
}

// NewServiceAccountCmd creates a command to generate a ServiceAccount manifest.
func (c *Client) NewServiceAccountCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("serviceaccount")
	if err != nil {
		panic(fmt.Sprintf("failed to create serviceaccount command: %v", err))
	}

	return cmd
}

// NewDeploymentCmd creates a command to generate a Deployment manifest.
func (c *Client) NewDeploymentCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("deployment")
	if err != nil {
		panic(fmt.Sprintf("failed to create deployment command: %v", err))
	}

	return cmd
}

// NewJobCmd creates a command to generate a Job manifest.
func (c *Client) NewJobCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("job")
	if err != nil {
		panic(fmt.Sprintf("failed to create job command: %v", err))
	}

	return cmd
}

// NewCronJobCmd creates a command to generate a CronJob manifest.
func (c *Client) NewCronJobCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("cronjob")
	if err != nil {
		panic(fmt.Sprintf("failed to create cronjob command: %v", err))
	}

	return cmd
}

// NewServiceCmd creates a command to generate a Service manifest.
func (c *Client) NewServiceCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("service")
	if err != nil {
		panic(fmt.Sprintf("failed to create service command: %v", err))
	}

	return cmd
}

// NewIngressCmd creates a command to generate an Ingress manifest.
func (c *Client) NewIngressCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("ingress")
	if err != nil {
		panic(fmt.Sprintf("failed to create ingress command: %v", err))
	}

	return cmd
}

// NewRoleCmd creates a command to generate a Role manifest.
func (c *Client) NewRoleCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("role")
	if err != nil {
		panic(fmt.Sprintf("failed to create role command: %v", err))
	}

	return cmd
}

// NewRoleBindingCmd creates a command to generate a RoleBinding manifest.
func (c *Client) NewRoleBindingCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("rolebinding")
	if err != nil {
		panic(fmt.Sprintf("failed to create rolebinding command: %v", err))
	}

	return cmd
}

// NewClusterRoleCmd creates a command to generate a ClusterRole manifest.
func (c *Client) NewClusterRoleCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("clusterrole")
	if err != nil {
		panic(fmt.Sprintf("failed to create clusterrole command: %v", err))
	}

	return cmd
}

// NewClusterRoleBindingCmd creates a command to generate a ClusterRoleBinding manifest.
func (c *Client) NewClusterRoleBindingCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("clusterrolebinding")
	if err != nil {
		panic(fmt.Sprintf("failed to create clusterrolebinding command: %v", err))
	}

	return cmd
}

// NewQuotaCmd creates a command to generate a ResourceQuota manifest.
func (c *Client) NewQuotaCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("quota")
	if err != nil {
		panic(fmt.Sprintf("failed to create quota command: %v", err))
	}

	return cmd
}

// NewPodDisruptionBudgetCmd creates a command to generate a PodDisruptionBudget manifest.
func (c *Client) NewPodDisruptionBudgetCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("poddisruptionbudget")
	if err != nil {
		panic(fmt.Sprintf("failed to create poddisruptionbudget command: %v", err))
	}

	return cmd
}

// NewPriorityClassCmd creates a command to generate a PriorityClass manifest.
func (c *Client) NewPriorityClassCmd() *cobra.Command {
	cmd, err := c.newResourceCmd("priorityclass")
	if err != nil {
		panic(fmt.Sprintf("failed to create priorityclass command: %v", err))
	}

	return cmd
}

// newResourceCmd creates a gen command that wraps kubectl create with forced --dry-run=client -o yaml.
func (c *Client) newResourceCmd(resourceType string) (*cobra.Command, error) {
	// Use empty string for kubeconfig since --dry-run=client doesn't need cluster access
	tempCreateCmd := c.CreateCreateCommand("")

	// Find the subcommand for this resource type
	var resourceCmd *cobra.Command

	for _, subCmd := range tempCreateCmd.Commands() {
		if subCmd.Name() == resourceType {
			resourceCmd = subCmd

			break
		}
	}

	if resourceCmd == nil {
		return nil, fmt.Errorf("%w: %s", ErrResourceCommandNotFound, resourceType)
	}

	// Create a wrapper command
	wrapperCmd := &cobra.Command{
		Use:          resourceCmd.Use,
		Short:        resourceCmd.Short,
		Long:         resourceCmd.Long,
		Example:      resourceCmd.Example,
		Aliases:      resourceCmd.Aliases,
		SilenceUsage: true,
	}

	// If the resource has subcommands (like secret/service), recursively copy them
	if len(resourceCmd.Commands()) > 0 {
		for _, subCmd := range resourceCmd.Commands() {
			subWrapper := c.createSubcommandWrapper(resourceType, subCmd)
			wrapperCmd.AddCommand(subWrapper)
		}
	} else {
		// Create our custom RunE that calls kubectl with forced flags
		wrapperCmd.RunE = func(cmd *cobra.Command, args []string) error {
			return c.executeResourceGen(resourceType, cmd, args)
		}

		// Copy all flags from the resource command
		wrapperCmd.Flags().AddFlagSet(resourceCmd.Flags())
	}

	return wrapperCmd, nil
}

// createSubcommandWrapper creates a wrapper for a subcommand (e.g., secret generic).
func (c *Client) createSubcommandWrapper(parentType string, subCmd *cobra.Command) *cobra.Command {
	wrapper := &cobra.Command{
		Use:          subCmd.Use,
		Short:        subCmd.Short,
		Long:         subCmd.Long,
		Example:      subCmd.Example,
		Aliases:      subCmd.Aliases,
		SilenceUsage: true,
	}

	// Create RunE for the subcommand
	wrapper.RunE = func(cmd *cobra.Command, args []string) error {
		return c.executeSubcommandGen(parentType, subCmd.Name(), cmd, args)
	}

	// Copy all flags from the subcommand
	wrapper.Flags().AddFlagSet(subCmd.Flags())

	return wrapper
}

// executeSubcommandGen executes kubectl create with subcommand and forced flags.
func (c *Client) executeSubcommandGen(
	parentType, subType string,
	cmd *cobra.Command,
	args []string,
) error {
	// Create a fresh kubectl create command
	createCmd := c.CreateCreateCommand("")

	// Find the parent resource command
	var parentCmd *cobra.Command

	for _, subCmd := range createCmd.Commands() {
		if subCmd.Name() == parentType {
			parentCmd = subCmd

			break
		}
	}

	if parentCmd == nil {
		return fmt.Errorf("%w: %s", ErrResourceCommandNotFound, parentType)
	}

	// Find the subcommand
	var freshSubCmd *cobra.Command

	for _, sub := range parentCmd.Commands() {
		if sub.Name() == subType {
			freshSubCmd = sub

			break
		}
	}

	if freshSubCmd == nil {
		return fmt.Errorf("%w: %s %s", ErrResourceCommandNotFound, parentType, subType)
	}

	// Force --dry-run=client and -o yaml
	err := c.setForcedFlags(freshSubCmd)
	if err != nil {
		return err
	}

	// Copy user flags
	err = c.copyUserFlags(cmd, freshSubCmd)
	if err != nil {
		return err
	}

	// Execute
	return c.executeCommand(freshSubCmd, args)
}

// executeResourceGen executes kubectl create with forced --dry-run=client -o yaml flags.
func (c *Client) executeResourceGen(resourceType string, cmd *cobra.Command, args []string) error {
	// Create a fresh kubectl create command
	createCmd := c.CreateCreateCommand("")

	// Find the resource command
	freshResourceCmd := c.findResourceCommand(createCmd, resourceType)
	if freshResourceCmd == nil {
		return fmt.Errorf("%w: %s", ErrResourceCommandNotFound, resourceType)
	}

	// Force --dry-run=client and -o yaml FIRST
	err := c.setForcedFlags(freshResourceCmd)
	if err != nil {
		return err
	}

	// Copy user flags
	err = c.copyUserFlags(cmd, freshResourceCmd)
	if err != nil {
		return err
	}

	// Execute
	return c.executeCommand(freshResourceCmd, args)
}

// findResourceCommand finds a kubectl create subcommand by resource type name.
func (c *Client) findResourceCommand(createCmd *cobra.Command, resourceType string) *cobra.Command {
	for _, subCmd := range createCmd.Commands() {
		if subCmd.Name() == resourceType {
			return subCmd
		}
	}

	return nil
}

// setForcedFlags sets the --dry-run=client and -o yaml flags.
func (c *Client) setForcedFlags(cmd *cobra.Command) error {
	err := cmd.Flags().Set("dry-run", "client")
	if err != nil {
		return fmt.Errorf("failed to set dry-run flag: %w", err)
	}

	err = cmd.Flags().Set("output", "yaml")
	if err != nil {
		return fmt.Errorf("failed to set output flag: %w", err)
	}

	return nil
}

// copyUserFlags copies user-provided flags from wrapper command to kubectl command.
func (c *Client) copyUserFlags(wrapperCmd, targetCmd *cobra.Command) error {
	var errs []error

	wrapperCmd.Flags().Visit(func(flag *pflag.Flag) {
		if flag.Name == "dry-run" || flag.Name == "output" {
			return
		}

		targetFlag := targetCmd.Flags().Lookup(flag.Name)
		if targetFlag != nil {
			err := c.copyFlagValue(flag, targetCmd)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to copy flag %s: %w", flag.Name, err))
			}
		}
	})

	if len(errs) > 0 {
		return fmt.Errorf("failed to copy flags: %w", errors.Join(errs...))
	}

	return nil
}

// copyFlagValue copies a flag value, handling slice flags specially.
func (c *Client) copyFlagValue(flag *pflag.Flag, targetCmd *cobra.Command) error {
	// For slice flags, we need to get the actual slice values
	if sliceVal, ok := flag.Value.(pflag.SliceValue); ok {
		strSlice := sliceVal.GetSlice()
		for _, v := range strSlice {
			err := targetCmd.Flags().Set(flag.Name, v)
			if err != nil {
				return fmt.Errorf("failed to set flag %s: %w", flag.Name, err)
			}
		}
	} else {
		// For non-slice flags, just copy the string value
		err := targetCmd.Flags().Set(flag.Name, flag.Value.String())
		if err != nil {
			return fmt.Errorf("failed to set flag %s: %w", flag.Name, err)
		}
	}

	return nil
}

// executeCommand executes the kubectl command's Run or RunE function.
func (c *Client) executeCommand(cmd *cobra.Command, args []string) error {
	if cmd.RunE != nil {
		err := cmd.RunE(cmd, args)
		if err != nil {
			return fmt.Errorf("kubectl command execution failed: %w", err)
		}

		return nil
	}

	if cmd.Run != nil {
		cmd.Run(cmd, args)

		return nil
	}

	return ErrNoRunFunction
}
