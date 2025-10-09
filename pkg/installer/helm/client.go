package helminstaller

import (
	"context"

	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/release"
)

// HelmClient defines the subset of Helm operations used by the installer.
type HelmClient interface {
	InstallChart(ctx context.Context, spec *helmclient.ChartSpec, opts *helmclient.GenericHelmOptions) (*release.Release, error)
	UninstallReleaseByName(name string) error
}
