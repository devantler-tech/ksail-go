package cluster

import (
	"errors"
	"fmt"
	"strings"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

const allFlag = "all"

var errAllFlagMissing = errors.New("all flag not found")

// NewListCmd creates the list command for clusters.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clusters",
		Long:  `List all Kubernetes clusters managed by KSail.`,
		RunE:  HandleListRunE,
	}

	cmd.Flags().
		BoolP(allFlag, "a", false, "List all clusters, including those not defined in the configuration")

	return cmd
}

// HandleListRunE handles the list command.
// Exported for testing purposes.
func HandleListRunE(
	cmd *cobra.Command,
	_ []string,
) error {
	// Create command utils
	utils, err := utils.NewCommandUtils(cmd)
	if err != nil {
		return fmt.Errorf("failed to create command utils: %w", err)
	}

	// Bind CLI only flags
	err = bindAllFlag(cmd, utils)
	if err != nil {
		return fmt.Errorf("failed to bind all flag: %w", err)
	}

	// Load cluster configuration
	err = utils.ConfigManager.LoadConfigSilent()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Resolve dependencies
	deps, err := utils.Resolver.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// List clusters
	err = listClusters(deps, cmd)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	return nil
}

func listClusters(deps *di.ResolvedDependencies, cmd *cobra.Command) error {
	clusters, err := deps.Provisioner.List(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	displayClusterList(clusters, cmd)
	return nil
}

func displayClusterList(clusters []string, cmd *cobra.Command) {
	if len(clusters) == 0 {
		notify.WriteMessage(notify.Message{
			Type:    notify.ActivityType,
			Content: "no clusters found",
			Writer:  cmd.OutOrStdout(),
		})
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), strings.Join(clusters, ", "))
	}
}

func bindAllFlag(cmd *cobra.Command, utils *utils.CommandUtils) error {
	flag := cmd.Flags().Lookup(allFlag)
	err := utils.ConfigManager.Viper.BindPFlag(allFlag, flag)
	if err != nil {
		return fmt.Errorf("failed to bind all flag: %w", err)
	}
	return nil
}
