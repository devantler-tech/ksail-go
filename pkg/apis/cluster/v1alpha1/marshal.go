package v1alpha1

import (
	"encoding/json"
	"fmt"
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

	data, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cluster: %w", err)
	}

	return data, nil
}

// buildClusterOutput converts a Cluster into a YAML/JSON-friendly projection with omitempty tags.
type clusterOutput struct {
	APIVersion string             `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string             `json:"kind,omitempty"       yaml:"kind,omitempty"`
	Spec       *clusterSpecOutput `json:"spec,omitempty"       yaml:"spec,omitempty"`
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
	Flux          *fluxOptionsOutput          `json:"flux,omitempty"          yaml:"flux,omitempty"`
	LocalRegistry *localRegistryOptionsOutput `json:"localRegistry,omitempty" yaml:"localRegistry,omitempty"`
}

type fluxOptionsOutput struct {
	Interval string `json:"interval,omitempty" yaml:"interval,omitempty"`
}

type localRegistryOptionsOutput struct {
	HostPort int32 `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
}

// buildConnectionOutput converts Connection to clusterConnectionOutput for marshaling.
func buildConnectionOutput(conn Connection) *clusterConnectionOutput {
	var out clusterConnectionOutput

	if conn.Kubeconfig != "" {
		out.Kubeconfig = conn.Kubeconfig
	}

	if conn.Context != "" {
		out.Context = conn.Context
	}

	if conn.Timeout.Duration != 0 {
		out.Timeout = conn.Timeout.Duration.String()
	}

	if out.Kubeconfig == "" && out.Context == "" && out.Timeout == "" {
		return nil
	}

	return &out
}

// buildOptionsOutput converts Options to clusterOptionsOutput for marshaling.
func buildOptionsOutput(opts Options) *clusterOptionsOutput {
	var out clusterOptionsOutput

	hasOpts := false

	if opts.Flux.Interval.Duration != 0 {
		out.Flux = &fluxOptionsOutput{Interval: opts.Flux.Interval.Duration.String()}
		hasOpts = true
	}

	if opts.LocalRegistry.HostPort != 0 {
		out.LocalRegistry = &localRegistryOptionsOutput{HostPort: opts.LocalRegistry.HostPort}
		hasOpts = true
	}

	if !hasOpts {
		return nil
	}

	return &out
}

// buildSpecComponentFields populates component-related fields in the spec output.
func buildSpecComponentFields(spec *clusterSpecOutput, clusterSpec Spec) bool {
	hasFields := false

	if clusterSpec.CNI != "" {
		spec.CNI = string(clusterSpec.CNI)
		hasFields = true
	}

	if clusterSpec.CSI != "" {
		spec.CSI = string(clusterSpec.CSI)
		hasFields = true
	}

	if clusterSpec.MetricsServer != "" {
		spec.MetricsServer = string(clusterSpec.MetricsServer)
		hasFields = true
	}

	if clusterSpec.LocalRegistry != "" {
		spec.LocalRegistry = string(clusterSpec.LocalRegistry)
		hasFields = true
	}

	if clusterSpec.GitOpsEngine != "" {
		spec.GitOpsEngine = string(clusterSpec.GitOpsEngine)
		hasFields = true
	}

	return hasFields
}

// buildSpecOutput converts Spec to clusterSpecOutput for marshaling.
// Returns nil if the spec has no non-empty fields.
func buildSpecOutput(clusterSpec Spec) *clusterSpecOutput {
	var spec clusterSpecOutput

	hasSpec := false

	if clusterSpec.Distribution != "" {
		spec.Distribution = string(clusterSpec.Distribution)
		hasSpec = true
	}

	if trimmed := strings.TrimSpace(clusterSpec.DistributionConfig); trimmed != "" {
		spec.DistributionConfig = trimmed
		hasSpec = true
	}

	if clusterSpec.SourceDirectory != "" {
		spec.SourceDirectory = clusterSpec.SourceDirectory
		hasSpec = true
	}

	if conn := buildConnectionOutput(clusterSpec.Connection); conn != nil {
		spec.Connection = conn
		hasSpec = true
	}

	if buildSpecComponentFields(&spec, clusterSpec) {
		hasSpec = true
	}

	if opts := buildOptionsOutput(clusterSpec.Options); opts != nil {
		spec.Options = opts
		hasSpec = true
	}

	if !hasSpec {
		return nil
	}

	return &spec
}

func buildClusterOutput(c Cluster) clusterOutput {
	specPtr := buildSpecOutput(c.Spec)

	return clusterOutput{
		APIVersion: c.APIVersion,
		Kind:       c.Kind,
		Spec:       specPtr,
	}
}

// pruneDistributionDefaults zeroes distribution-related fields that match base defaults.
// Only prunes values that match the absolute base defaults, not derived defaults for other distributions.
func pruneDistributionDefaults(spec *Spec, distribution Distribution) {
	// Only prune if distribution matches the base default
	if spec.Distribution == DefaultDistribution {
		spec.Distribution = ""
	}

	// Only prune distributionConfig if it matches the base default (kind.yaml)
	// Do NOT prune if it's a derived default for a non-default distribution (e.g., k3d.yaml when using K3d)
	trimmedConfig := strings.TrimSpace(spec.DistributionConfig)
	if trimmedConfig == "" || trimmedConfig == DefaultDistributionConfig {
		spec.DistributionConfig = ""
	}

	if spec.SourceDirectory == DefaultSourceDirectory || spec.SourceDirectory == "" {
		spec.SourceDirectory = ""
	}
}

// pruneConnectionDefaults zeroes connection fields that match base defaults or distribution-specific defaults.
func pruneConnectionDefaults(conn *Connection, distribution Distribution) {
	if conn.Kubeconfig == DefaultKubeconfigPath || conn.Kubeconfig == "" {
		conn.Kubeconfig = ""
	}

	// Prune context if it matches the expected default for the current distribution
	// This is intentional: the context is a derived default that changes with distribution
	if defaultCtx := ExpectedContextName(distribution); conn.Context == defaultCtx {
		conn.Context = ""
	}

	if conn.Timeout.Duration == 0 {
		conn.Timeout = metav1.Duration{}
	}
}

// pruneComponentDefaults zeroes component fields (CNI, CSI, MetricsServer, etc.) that match base defaults.
func pruneComponentDefaults(spec *Spec) {
	if spec.CNI == DefaultCNI {
		spec.CNI = ""
	}

	if spec.CSI == DefaultCSI {
		spec.CSI = ""
	}

	if spec.MetricsServer == DefaultMetricsServer || spec.MetricsServer == "" {
		spec.MetricsServer = ""
	}

	if spec.LocalRegistry == DefaultLocalRegistry || spec.LocalRegistry == "" {
		spec.LocalRegistry = ""
	}

	if spec.GitOpsEngine == DefaultGitOpsEngine || spec.GitOpsEngine == "" {
		spec.GitOpsEngine = ""
	}
}

// pruneOptionsDefaults zeroes option fields that match defaults.
func pruneOptionsDefaults(opts *Options) {
	if opts.Flux.Interval == DefaultFluxInterval || opts.Flux.Interval.Duration == 0 {
		opts.Flux.Interval = metav1.Duration{}
	}

	if opts.LocalRegistry.HostPort == DefaultLocalRegistryPort || opts.LocalRegistry.HostPort == 0 {
		opts.LocalRegistry.HostPort = 0
	}
}

// pruneClusterDefaults zeroes fields that match default values so they are omitted when marshalled.
func pruneClusterDefaults(cluster Cluster) Cluster {
	// Distribution defaults
	distribution := cluster.Spec.Distribution
	if distribution == "" {
		distribution = DistributionKind
	}

	pruneDistributionDefaults(&cluster.Spec, distribution)
	pruneConnectionDefaults(&cluster.Spec.Connection, distribution)
	pruneComponentDefaults(&cluster.Spec)
	pruneOptionsDefaults(&cluster.Spec.Options)

	return cluster
}
