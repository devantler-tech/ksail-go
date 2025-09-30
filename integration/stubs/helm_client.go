package stubs

import (
	"context"

	helmclient "github.com/mittwald/go-helm-client"
)

// HelmClientStub is a stub implementation of HelmClient interface.
type HelmClientStub struct {
	InstallError   error
	UninstallError error
	
	InstallCalls   []string
	UninstallCalls []string
}

// NewHelmClientStub creates a new HelmClientStub.
func NewHelmClientStub() *HelmClientStub {
	return &HelmClientStub{}
}

// Install simulates Helm chart installation.
func (h *HelmClientStub) Install(ctx context.Context, spec *helmclient.ChartSpec) error {
	h.InstallCalls = append(h.InstallCalls, spec.ReleaseName)
	return h.InstallError
}

// Uninstall simulates Helm chart uninstallation.
func (h *HelmClientStub) Uninstall(name string) error {
	h.UninstallCalls = append(h.UninstallCalls, name)
	return h.UninstallError
}