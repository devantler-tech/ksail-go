package installer

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

// InstallOrUpgradeHelmChart performs a Helm install or upgrade operation.
func InstallOrUpgradeHelmChart(
	ctx context.Context,
	client helm.Interface,
	repoConfig HelmRepoConfig,
	chartConfig HelmChartConfig,
	timeout time.Duration,
) error {
	repoEntry := &helm.RepositoryEntry{
		Name: repoConfig.Name,
		URL:  repoConfig.URL,
	}

	addRepoErr := client.AddRepository(ctx, repoEntry)
	if addRepoErr != nil {
		return fmt.Errorf("failed to add %s repository: %w", repoConfig.RepoName, addRepoErr)
	}

	spec := &helm.ChartSpec{
		ReleaseName:     chartConfig.ReleaseName,
		ChartName:       chartConfig.ChartName,
		Namespace:       chartConfig.Namespace,
		RepoURL:         chartConfig.RepoURL,
		CreateNamespace: chartConfig.CreateNamespace,
		Atomic:          true,
		Silent:          true,
		UpgradeCRDs:     true,
		Timeout:         timeout,
		Wait:            true,
		WaitForJobs:     true,
		SetJSONVals:     chartConfig.SetJSONVals,
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := client.InstallOrUpgradeChart(timeoutCtx, spec)
	if err != nil {
		return fmt.Errorf("failed to install %s chart: %w", repoConfig.RepoName, err)
	}

	return nil
}
