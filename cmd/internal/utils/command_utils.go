package utils

import (
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

type CommandUtils struct {
	ConfigManager *configmanager.ConfigManager
	Timer         timer.Timer
	Resolver      *di.Resolver
}

func NewCommandUtils(
	cmd *cobra.Command,
	fieldSelectors ...configmanager.FieldSelector[v1alpha1.Cluster],
) (*CommandUtils, error) {
	configManager := configmanager.NewConfigManager(
		cmd.OutOrStdout(),
		fieldSelectors...,
	)
	configManager.AddFlagsFromFields(cmd)

	resolver, _ := di.NewResolver(configManager.Config)

	return &CommandUtils{
		ConfigManager: configManager,
		Timer:         timer.New(),
		Resolver:      resolver,
	}, nil
}
