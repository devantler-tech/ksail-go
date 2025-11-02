package gen

import (
	"errors"
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

var (
	// ErrResourceCommandNotFound is returned when a kubectl create subcommand is not found.
	ErrResourceCommandNotFound = errors.New("kubectl create command not found for resource type")
	// ErrNoRunFunction is returned when a kubectl command has neither RunE nor Run function.
	ErrNoRunFunction = errors.New("no run function found for kubectl create command")
)

// KubectlGenerator generates Kubernetes resource manifests using kubectl create with --dry-run=client -o yaml.
type KubectlGenerator struct {
	kubeconfigPath string
}

// NewKubectlGenerator creates a new kubectl-based generator.
func NewKubectlGenerator(kubeconfigPath string) *KubectlGenerator {
	return &KubectlGenerator{
		kubeconfigPath: kubeconfigPath,
	}
}

// GenerateCommand creates a gen subcommand that wraps kubectl create with forced --dry-run=client -o yaml.
func (g *KubectlGenerator) GenerateCommand(resourceType string) *cobra.Command {
	// Create temporary kubectl client to get the resource command
	tempIOStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
	tempClient := kubectl.NewClient(tempIOStreams)
	tempCreateCmd := tempClient.CreateCreateCommand(g.kubeconfigPath)

	// Find the subcommand for this resource type
	var resourceCmd *cobra.Command

	for _, subCmd := range tempCreateCmd.Commands() {
		if subCmd.Name() == resourceType {
			resourceCmd = subCmd

			break
		}
	}

	if resourceCmd == nil {
		panic(fmt.Sprintf("kubectl create %s command not found", resourceType))
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

	// Create our custom RunE that calls kubectl with forced flags
	wrapperCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return g.executeKubectlCreate(cmd, args, resourceType)
	}

	// Copy all flags from the resource command
	wrapperCmd.Flags().AddFlagSet(resourceCmd.Flags())

	return wrapperCmd
}

// executeKubectlCreate executes the kubectl create command with forced --dry-run=client -o yaml flags.
func (g *KubectlGenerator) executeKubectlCreate(
	cmd *cobra.Command,
	args []string,
	resourceType string,
) error {
	// Create IO streams for kubectl
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// Create a fresh kubectl client and command
	client := kubectl.NewClient(ioStreams)
	createCmd := client.CreateCreateCommand(g.kubeconfigPath)

	// Find the resource command again with the new IO streams
	freshResourceCmd := findResourceCommand(createCmd, resourceType)
	if freshResourceCmd == nil {
		return fmt.Errorf("%w: %s", ErrResourceCommandNotFound, resourceType)
	}

	// Force --dry-run=client and -o yaml FIRST before copying user flags
	err := freshResourceCmd.Flags().Set("dry-run", "client")
	if err != nil {
		return fmt.Errorf("failed to set dry-run flag: %w", err)
	}

	err = freshResourceCmd.Flags().Set("output", "yaml")
	if err != nil {
		return fmt.Errorf("failed to set output flag: %w", err)
	}

	// Copy all flags from wrapper to resource command (these will override defaults but not our forced flags above)
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		// Skip the flags we want to force
		if flag.Name == "dry-run" || flag.Name == "output" {
			return
		}
		// Set the flag on the resource command
		targetFlag := freshResourceCmd.Flags().Lookup(flag.Name)
		if targetFlag != nil {
			// For slice flags, we need to get the actual slice values
			// pflag has StringSliceVar which stores []string internally
			if sliceVal, ok := flag.Value.(pflag.SliceValue); ok {
				// Get the slice as []string
				strSlice := sliceVal.GetSlice()
				// Set each value separately
				for _, v := range strSlice {
					_ = freshResourceCmd.Flags().Set(flag.Name, v)
				}
			} else {
				// For non-slice flags, just copy the string value
				_ = freshResourceCmd.Flags().Set(flag.Name, flag.Value.String())
			}
		}
	})

	// Call the fresh kubectl command's RunE or Run
	if freshResourceCmd.RunE != nil {
		err = freshResourceCmd.RunE(freshResourceCmd, args)
		if err != nil {
			return fmt.Errorf("kubectl command execution failed: %w", err)
		}

		return nil
	}

	if freshResourceCmd.Run != nil {
		freshResourceCmd.Run(freshResourceCmd, args)

		return nil
	}

	return fmt.Errorf("%w: %s", ErrNoRunFunction, resourceType)
}

// findResourceCommand finds a kubectl create subcommand by resource type name.
func findResourceCommand(createCmd *cobra.Command, resourceType string) *cobra.Command {
	for _, subCmd := range createCmd.Commands() {
		if subCmd.Name() == resourceType {
			return subCmd
		}
	}

	return nil
}
