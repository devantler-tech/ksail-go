package validators

import (
	"errors"
	"fmt"

	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	v1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ConfigValidator validates consistency between ksail cluster config and distribution config files.
// Only a subset of the .NET validator is implemented (cluster name + context naming) because
// other advanced spec fields (mirror registries, flux source, metrics server toggles, etc.) are not yet modeled.
type ConfigValidator struct {
	cfg     *ksailcluster.Cluster
	kindCfg *v1alpha4.Cluster
	k3dCfg  *v1alpha5.SimpleConfig
}

func NewConfigValidator(cfg *ksailcluster.Cluster) *ConfigValidator {
	return &ConfigValidator{cfg: cfg}
}

// SetDistributionConfigs sets the distribution configs for validation.
func (v *ConfigValidator) SetDistributionConfigs(kindCfg *v1alpha4.Cluster, k3dCfg *v1alpha5.SimpleConfig) {
	v.kindCfg = kindCfg
	v.k3dCfg = k3dCfg
}

// Validate performs validation of configuration files.
func (v *ConfigValidator) Validate() error {
	if v.cfg == nil {
		return errors.New("config is nil")
	}

	fmt.Println("🕵 Validating project files and config")
	fmt.Println("► validating config")

	if err := v.checkContextName(v.cfg); err != nil {
		return err
	}

	if err := v.checkDistributionConfig(v.cfg); err != nil {
		return err
	}

	fmt.Println("✔ configuration is valid")

	return nil
}

// --- internals ---

func (v *ConfigValidator) checkContextName(cfg *ksailcluster.Cluster) error {
	var expected string

	switch cfg.Spec.Distribution {
	case ksailcluster.DistributionKind:
		expected = fmt.Sprintf("kind-%s", cfg.Metadata.Name)
	case ksailcluster.DistributionK3d:
		expected = fmt.Sprintf("k3d-%s", cfg.Metadata.Name)
	default:
		return fmt.Errorf("unsupported distribution '%s'", cfg.Spec.Distribution)
	}

	if ctx := cfg.Spec.Connection.Context; ctx != "" && ctx != expected {
		return fmt.Errorf("spec.connection.context '%s' does not match expected '%s'", ctx, expected)
	}

	return nil
}

func (v *ConfigValidator) checkDistributionConfig(cfg *ksailcluster.Cluster) error {
	switch cfg.Spec.Distribution {
	case ksailcluster.DistributionKind:
		if v.kindCfg != nil {
			return v.validateKindConfig(v.kindCfg, cfg)
		}
	case ksailcluster.DistributionK3d:
		if v.k3dCfg != nil {
			return v.validateK3dConfig(v.k3dCfg, cfg)
		}
	}

	return nil
}

func (v *ConfigValidator) validateKindConfig(kindCfg *v1alpha4.Cluster, cfg *ksailcluster.Cluster) error {
	if kindCfg.Name != "" && kindCfg.Name != cfg.Metadata.Name {
		return fmt.Errorf("kind config name '%s' does not match ksail.yaml metadata.name '%s'", kindCfg.Name, cfg.Metadata.Name)
	}

	return nil
}

func (v *ConfigValidator) validateK3dConfig(k3dCfg *v1alpha5.SimpleConfig, cfg *ksailcluster.Cluster) error {
	if k3dCfg.Name != "" && k3dCfg.Name != cfg.Metadata.Name {
		return fmt.Errorf("k3d config metadata.name '%s' does not match ksail.yaml metadata.name '%s'", k3dCfg.Name, cfg.Metadata.Name)
	}

	return nil
}
