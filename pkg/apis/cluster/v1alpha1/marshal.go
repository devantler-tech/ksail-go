package v1alpha1

import (
	"encoding/json"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MarshalYAML trims default values before emitting YAML.
func (c Cluster) MarshalYAML() (any, error) {
	pruned := pruneClusterDefaults(c)
	out := buildClusterOutput(pruned)
	return out, nil
}

// MarshalJSON trims default values before emitting JSON (used by YAML library).
func (c Cluster) MarshalJSON() ([]byte, error) {
	pruned := pruneClusterDefaults(c)
	out := buildClusterOutput(pruned)
	return json.Marshal(out)
}

// buildClusterOutput converts a Cluster into a YAML/JSON-friendly projection with omitempty tags.
type clusterOutput struct {
	APIVersion string             `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string             `json:"kind,omitempty" yaml:"kind,omitempty"`
	Spec       *clusterSpecOutput `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type clusterSpecOutput struct {
	Distribution       string                   `json:"distribution,omitempty"       yaml:"distribution,omitempty"`
	DistributionConfig string                   `json:"distributionConfig,omitempty" yaml:"distributionConfig,omitempty"`
	SourceDirectory    string                   `json:"sourceDirectory,omitempty"    yaml:"sourceDirectory,omitempty"`
	Connection         *clusterConnectionOutput `json:"connection,omitempty"         yaml:"connection,omitempty"`
	CNI                string                   `json:"cni,omitempty"                yaml:"cni,omitempty"`
	CSI                string                   `json:"csi,omitempty"                yaml:"csi,omitempty"`
	MetricsServer      string                   `json:"metricsServer,omitempty"      yaml:"metricsServer,omitempty"`
	LocalRegistry      string                   `json:"localRegistry,omitempty"      yaml:"localRegistry,omitempty"`
	GitOpsEngine       string                   `json:"gitOpsEngine,omitempty"       yaml:"gitOpsEngine,omitempty"`
	Options            *clusterOptionsOutput    `json:"options,omitempty"            yaml:"options,omitempty"`
}

type clusterConnectionOutput struct {
	Kubeconfig string `json:"kubeconfig,omitempty" yaml:"kubeconfig,omitempty"`
	Context    string `json:"context,omitempty"    yaml:"context,omitempty"`
	Timeout    string `json:"timeout,omitempty"    yaml:"timeout,omitempty"`
}

type clusterOptionsOutput struct {
	Flux          *fluxOptionsOutput          `json:"flux,omitempty" yaml:"flux,omitempty"`
	LocalRegistry *localRegistryOptionsOutput `json:"localRegistry,omitempty" yaml:"localRegistry,omitempty"`
}

type fluxOptionsOutput struct {
	Interval string `json:"interval,omitempty" yaml:"interval,omitempty"`
}

type localRegistryOptionsOutput struct {
	HostPort int32 `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
}

//nolint:cyclop // marshalling logic requires checking multiple optional fields
func buildClusterOutput(c Cluster) clusterOutput {
	var spec clusterSpecOutput

	hasSpec := false

	if c.Spec.Distribution != "" {
		spec.Distribution = string(c.Spec.Distribution)
		hasSpec = true
	}

	if trimmed := strings.TrimSpace(c.Spec.DistributionConfig); trimmed != "" {
		spec.DistributionConfig = trimmed
		hasSpec = true
	}

	if c.Spec.SourceDirectory != "" {
		spec.SourceDirectory = c.Spec.SourceDirectory
		hasSpec = true
	}

	var conn clusterConnectionOutput
	if c.Spec.Connection.Kubeconfig != "" {
		conn.Kubeconfig = c.Spec.Connection.Kubeconfig
	}
	if c.Spec.Connection.Context != "" {
		conn.Context = c.Spec.Connection.Context
	}

	if c.Spec.Connection.Timeout.Duration != 0 {
		conn.Timeout = c.Spec.Connection.Timeout.Duration.String()
	}

	if conn.Kubeconfig != "" || conn.Context != "" || conn.Timeout != "" {
		spec.Connection = &conn
		hasSpec = true
	}

	if c.Spec.CNI != "" {
		spec.CNI = string(c.Spec.CNI)
		hasSpec = true
	}

	if c.Spec.CSI != "" {
		spec.CSI = string(c.Spec.CSI)
		hasSpec = true
	}

	if c.Spec.MetricsServer != "" {
		spec.MetricsServer = string(c.Spec.MetricsServer)
		hasSpec = true
	}

	if c.Spec.LocalRegistry != "" {
		spec.LocalRegistry = string(c.Spec.LocalRegistry)
		hasSpec = true
	}

	if c.Spec.GitOpsEngine != "" {
		spec.GitOpsEngine = string(c.Spec.GitOpsEngine)
		hasSpec = true
	}

	var opts clusterOptionsOutput

	hasOpts := false

	if c.Spec.Options.Flux.Interval.Duration != 0 {
		opts.Flux = &fluxOptionsOutput{Interval: c.Spec.Options.Flux.Interval.Duration.String()}
		hasOpts = true
	}

	if c.Spec.Options.LocalRegistry.HostPort != 0 {
		opts.LocalRegistry = &localRegistryOptionsOutput{HostPort: c.Spec.Options.LocalRegistry.HostPort}
		hasOpts = true
	}
	if hasOpts {
		spec.Options = &opts
		hasSpec = true
	}

	var specPtr *clusterSpecOutput
	if hasSpec {
		specPtr = &spec
	}

	return clusterOutput{
		APIVersion: c.APIVersion,
		Kind:       c.Kind,
		Spec:       specPtr,
	}
}

// pruneClusterDefaults zeroes fields that match default values so they are omitted when marshalled.
func pruneClusterDefaults(c Cluster) Cluster {
//nolint:cyclop // default pruning requires checking multiple fields
	// Distribution defaults
	distribution := c.Spec.Distribution
	if distribution == "" {
		distribution = DistributionKind
	}

	if c.Spec.Distribution == DistributionKind {
		c.Spec.Distribution = ""
	}

	expectedDistConfig := ExpectedDistributionConfigName(distribution)

	trimmedConfig := strings.TrimSpace(c.Spec.DistributionConfig)
	if trimmedConfig == "" || trimmedConfig == expectedDistConfig {
		c.Spec.DistributionConfig = ""
	}

	if c.Spec.SourceDirectory == DefaultSourceDirectory || c.Spec.SourceDirectory == "" {
		c.Spec.SourceDirectory = ""
	}

	if c.Spec.Connection.Kubeconfig == DefaultKubeconfigPath || c.Spec.Connection.Kubeconfig == "" {
		c.Spec.Connection.Kubeconfig = ""
	}

	if defaultCtx := ExpectedContextName(distribution); c.Spec.Connection.Context == defaultCtx {
		c.Spec.Connection.Context = ""
	}

	if c.Spec.Connection.Timeout.Duration == 0 {
		c.Spec.Connection.Timeout = metav1.Duration{}
	}

	if c.Spec.CNI == CNIDefault {
		c.Spec.CNI = ""
	}

	if c.Spec.CSI == CSIDefault {
		c.Spec.CSI = ""
	}

	if c.Spec.MetricsServer == MetricsServerEnabled || c.Spec.MetricsServer == "" {
		c.Spec.MetricsServer = ""
	}

	if c.Spec.LocalRegistry == LocalRegistryDisabled || c.Spec.LocalRegistry == "" {
		c.Spec.LocalRegistry = ""
	}

	if c.Spec.GitOpsEngine == GitOpsEngineNone || c.Spec.GitOpsEngine == "" {
		c.Spec.GitOpsEngine = ""
	}

	if c.Spec.Options.Flux.Interval == DefaultFluxInterval || c.Spec.Options.Flux.Interval.Duration == 0 {
		c.Spec.Options.Flux.Interval = metav1.Duration{}
	}

	if c.Spec.Options.LocalRegistry.HostPort == DefaultLocalRegistryPort || c.Spec.Options.LocalRegistry.HostPort == 0 {
		c.Spec.Options.LocalRegistry.HostPort = 0
	}

	return c
}
