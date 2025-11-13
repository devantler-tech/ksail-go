package installer

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

// InstallOrUpgradeHelmChart performs a Helm install or upgrade operation.
//
// Deprecated: Use helm.InstallOrUpgradeChart instead. This function is maintained for backward compatibility.
func InstallOrUpgradeHelmChart(
	ctx context.Context,
	client helm.Interface,
	repoConfig HelmRepoConfig,
	chartConfig HelmChartConfig,
	timeout time.Duration,
) error {
	err := helm.InstallOrUpgradeChart(ctx, client, repoConfig, chartConfig, timeout)
	if err != nil {
		return fmt.Errorf("install or upgrade helm chart: %w", err)
	}

	return nil
}
