package fluxinstaller

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// FluxInstaller implements the installer.Installer interface for Flux.
type FluxInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
}

// NewFluxInstaller creates a new Flux installer instance.
func NewFluxInstaller(
	kubeconfig, context string,
	timeout time.Duration,
) *FluxInstaller {
	return &FluxInstaller{
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}
}

// Install installs Flux using the flux install command.
func (f *FluxInstaller) Install(ctx context.Context) error {
	args := f.buildInstallArgs()
	
	timeoutCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "flux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install Flux: %w (output: %s)", err, string(output))
	}

	return nil
}

// Uninstall removes Flux using the flux uninstall command.
func (f *FluxInstaller) Uninstall(ctx context.Context) error {
	args := f.buildUninstallArgs()
	
	timeoutCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "flux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall Flux: %w (output: %s)", err, string(output))
	}

	return nil
}

// --- internals ---

// buildInstallArgs constructs the arguments for flux install command.
func (f *FluxInstaller) buildInstallArgs() []string {
	args := []string{
		"install",
		"--namespace=flux-system",
		"--network-policy=false",
	}

	if f.kubeconfig != "" {
		args = append(args, fmt.Sprintf("--kubeconfig=%s", f.kubeconfig))
	}

	if f.context != "" {
		args = append(args, fmt.Sprintf("--context=%s", f.context))
	}

	return args
}

// buildUninstallArgs constructs the arguments for flux uninstall command.
func (f *FluxInstaller) buildUninstallArgs() []string {
	args := []string{
		"uninstall",
		"--namespace=flux-system",
		"--silent",
	}

	if f.kubeconfig != "" {
		args = append(args, fmt.Sprintf("--kubeconfig=%s", f.kubeconfig))
	}

	if f.context != "" {
		args = append(args, fmt.Sprintf("--context=%s", f.context))
	}

	return args
}
