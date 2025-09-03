// Package fluxinstaller provides an installer for installing flux on a Kubernetes cluster.
package fluxinstaller

import (
	"context"

	helmclient "github.com/mittwald/go-helm-client"
)

// HelmClient defines the subset of Helm operations used by the installer.
type HelmClient interface {
	Install(ctx context.Context, spec *helmclient.ChartSpec) error
	Uninstall(name string) error
}
