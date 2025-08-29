// Package eksgenerator provides utilities for generating EKS cluster configurations.
package eksgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// EKSGenerator generates an EKS ClusterConfig YAML.
type EKSGenerator struct {
	io.FileWriter

	Marshaller marshaller.Marshaller[*v1alpha5.ClusterConfig]
}

// NewEKSGenerator creates and returns a new EKSGenerator instance.
func NewEKSGenerator() *EKSGenerator {
	m := yamlmarshaller.NewMarshaller[*v1alpha5.ClusterConfig]()

	return &EKSGenerator{
		FileWriter: io.FileWriter{},
		Marshaller: m,
	}
}

// Generate creates an EKS cluster YAML configuration and writes it to the specified output.
func (g *EKSGenerator) Generate(cluster *v1alpha1.Cluster, opts yamlgenerator.Options) (string, error) {
	cfg := &v1alpha5.ClusterConfig{
		TypeMeta: v1alpha5.ClusterConfigTypeMeta(),
		Metadata: &v1alpha5.ClusterMeta{
			Name:   cluster.Metadata.Name,
			Region: getRegion(cluster),
		},
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name:         fmt.Sprintf("%s-workers", cluster.Metadata.Name),
					InstanceType: getNodeType(cluster),
					ScalingConfig: &v1alpha5.ScalingConfig{
						MinSize:         getMinNodes(cluster),
						MaxSize:         getMaxNodes(cluster),
						DesiredCapacity: getDesiredNodes(cluster),
					},
				},
			},
		},
	}

	// Set Kubernetes version if specified
	if version := getKubernetesVersion(cluster); version != "" {
		cfg.Metadata.Version = version
	}

	out, err := g.Marshaller.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal EKS config: %w", err)
	}

	// write to file if output path is specified
	if opts.Output != "" {
		result, err := g.TryWrite(out, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("write EKS config: %w", err)
		}

		return result, nil
	}

	return out, nil
}

// getRegion extracts the AWS region from cluster options or returns default.
func getRegion(cluster *v1alpha1.Cluster) string {
	if cluster.Spec.Options.EKS.AWSRegion != "" {
		return cluster.Spec.Options.EKS.AWSRegion
	}
	return "us-west-2" // default region
}

// getNodeType extracts the node type from cluster options or returns default.
func getNodeType(cluster *v1alpha1.Cluster) string {
	if cluster.Spec.Options.EKS.NodeType != "" {
		return cluster.Spec.Options.EKS.NodeType
	}
	return "m5.large" // default instance type
}

// getMinNodes extracts the minimum nodes from cluster options or returns default.
func getMinNodes(cluster *v1alpha1.Cluster) *int {
	if cluster.Spec.Options.EKS.MinNodes > 0 {
		minNodes := cluster.Spec.Options.EKS.MinNodes
		return &minNodes
	}
	defaultMin := 1
	return &defaultMin
}

// getMaxNodes extracts the maximum nodes from cluster options or returns default.
func getMaxNodes(cluster *v1alpha1.Cluster) *int {
	if cluster.Spec.Options.EKS.MaxNodes > 0 {
		maxNodes := cluster.Spec.Options.EKS.MaxNodes
		return &maxNodes
	}
	defaultMax := 3
	return &defaultMax
}

// getDesiredNodes extracts the desired nodes from cluster options or returns default.
func getDesiredNodes(cluster *v1alpha1.Cluster) *int {
	if cluster.Spec.Options.EKS.DesiredNodes > 0 {
		desiredNodes := cluster.Spec.Options.EKS.DesiredNodes
		return &desiredNodes
	}
	defaultDesired := 2
	return &defaultDesired
}

// getKubernetesVersion extracts the Kubernetes version from cluster options.
func getKubernetesVersion(cluster *v1alpha1.Cluster) string {
	return cluster.Spec.Options.EKS.KubernetesVersion
}