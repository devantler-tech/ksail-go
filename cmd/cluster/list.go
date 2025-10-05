package cluster

import (
	"fmt"
	"strings"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

const allFlag = "all"

// NewListCmd creates the list command for clusters.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List clusters",
		Long:         `List all Kubernetes clusters managed by KSail.`,
		SilenceUsage: true,
	}

	utils, _ := utils.NewCommandUtils(
		cmd,
		configmanager.DefaultDistributionFieldSelector(),
		configmanager.DefaultDistributionConfigFieldSelector(),
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return HandleListRunE(cmd, utils, args)
	}

	bindAllFlag(cmd, utils)

	return cmd
}

// HandleListRunE handles the list command.
// Exported for testing purposes.
func HandleListRunE(
	cmd *cobra.Command,
	utils *utils.CommandUtils,
	_ []string,
) error {
	// Load cluster configuration
	err := utils.ConfigManager.LoadConfigSilent()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// List clusters
	err = listClusters(utils, cmd)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	return nil
}

func listClusters(utils *utils.CommandUtils, cmd *cobra.Command) error {
	deps, err := utils.Resolver.Resolve()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	clusters, err := deps.Provisioner.List(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}

	displayClusterList(utils.ConfigManager.Config.Spec.Distribution, clusters, cmd)

	if utils.ConfigManager.Viper.GetBool(allFlag) {
		// for each distribution that is not utils.ConfigManager.Config.Cluster.Distribution
		// create a new provisioner for that distribution and list clusters
		// You are not able to use the resolver for this.
		distributions := []v1alpha1.Distribution{
			v1alpha1.DistributionKind,
			v1alpha1.DistributionK3d,
		}
		for _, distribution := range distributions {
			if distribution == utils.ConfigManager.Config.Spec.Distribution {
				continue
			}
			otherProv, _, err := clusterprovisioner.CreateClusterProvisioner(
				cmd.Context(),
				distribution,
				utils.ConfigManager.Config.Spec.DistributionConfig,
				utils.ConfigManager.Config.Spec.Connection.Kubeconfig,
			)
			if err != nil {
				return fmt.Errorf(
					"failed to create provisioner for distribution %s: %w",
					distribution,
					err,
				)
			}
			otherClusters, err := otherProv.List(cmd.Context())
			if err != nil {
				return fmt.Errorf(
					"failed to list clusters for distribution %s: %w",
					distribution,
					err,
				)
			}
			displayClusterList(distribution, otherClusters, cmd)
		}
	}

	return nil
}

func displayClusterList(distribution v1alpha1.Distribution, clusters []string, cmd *cobra.Command) {
	if len(clusters) == 0 {
		notify.WriteMessage(notify.Message{
			Type:    notify.ActivityType,
			Content: "no clusters found",
			Writer:  cmd.OutOrStdout(),
		})
	} else {
		fmt.Fprint(cmd.OutOrStdout(), string(distribution)+": ")
		fmt.Fprintln(cmd.OutOrStdout(), strings.Join(clusters, ", "))
	}
}

func bindAllFlag(cmd *cobra.Command, utils *utils.CommandUtils) {
	cmd.Flags().
		BoolP(allFlag, "a", false, "List all clusters, including those not defined in the configuration")
	flag := cmd.Flags().Lookup(allFlag)
	_ = utils.ConfigManager.Viper.BindPFlag(allFlag, flag)
}
