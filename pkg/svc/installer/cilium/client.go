package ciliuminstaller

import (
	"context"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
)

// HelmClient defines the interface for Helm operations needed by the installer.
// This wraps our consolidated helm client interface.
//
//go:generate mockery --name=HelmClient --output=. --filename=mocks.go
type HelmClient interface {
	InstallOrUpgradeChart(ctx context.Context, spec *helm.ChartSpec) (*helm.ReleaseInfo, error)
	UninstallRelease(ctx context.Context, releaseName, namespace string) error
	AddRepository(ctx context.Context, entry *helm.RepositoryEntry) error
}
