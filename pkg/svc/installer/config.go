package installer

import "github.com/devantler-tech/ksail-go/pkg/client/helm"

// HelmRepoConfig holds repository configuration for a Helm chart.
//
// Deprecated: Use helm.RepoConfig instead. This type is maintained for backward compatibility.
type HelmRepoConfig = helm.RepoConfig

// HelmChartConfig holds chart installation configuration.
//
// Deprecated: Use helm.ChartConfig instead. This type is maintained for backward compatibility.
type HelmChartConfig = helm.ChartConfig
