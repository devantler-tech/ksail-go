// Package awsprovisioner provides an AWS implementation of the ProviderProvisioner interface.
package awsprovisioner

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AWSProvisioner implements ProviderProvisioner for AWS.
type AWSProvisioner struct {
	region  string
	profile string
}

// NewAWSProvisioner creates a new AWSProvisioner with the specified region and profile.
func NewAWSProvisioner(region, profile string) *AWSProvisioner {
	return &AWSProvisioner{
		region:  region,
		profile: profile,
	}
}

// CheckReady checks if AWS CLI and credentials are available.
func (a *AWSProvisioner) CheckReady() (bool, error) {
	// Check if AWS CLI is available
	if err := a.checkAWSCLI(); err != nil {
		return false, fmt.Errorf("AWS CLI not available: %w", err)
	}

	// Check if eksctl is available
	if err := a.checkEksctl(); err != nil {
		return false, fmt.Errorf("eksctl not available: %w", err)
	}

	// Check if AWS credentials are configured
	if err := a.checkCredentials(); err != nil {
		return false, fmt.Errorf("AWS credentials not configured: %w", err)
	}

	return true, nil
}

// checkAWSCLI verifies that the AWS CLI is installed and available.
func (a *AWSProvisioner) checkAWSCLI() error {
	cmd := exec.Command("aws", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("aws command not found or failed: %w", err)
	}
	return nil
}

// checkEksctl verifies that eksctl is installed and available.
func (a *AWSProvisioner) checkEksctl() error {
	cmd := exec.Command("eksctl", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("eksctl command not found or failed: %w", err)
	}
	return nil
}

// checkCredentials verifies that AWS credentials are configured and can be used.
func (a *AWSProvisioner) checkCredentials() error {
	ctx := context.Background()

	// Set up config options
	var opts []func(*config.LoadOptions) error
	if a.region != "" {
		opts = append(opts, config.WithRegion(a.region))
	}
	if a.profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(a.profile))
		// Also set the AWS_PROFILE environment variable for CLI compatibility
		os.Setenv("AWS_PROFILE", a.profile)
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create STS client to test credentials
	stsClient := sts.NewFromConfig(cfg)

	// Try to get caller identity to verify credentials work
	_, err = stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("failed to verify AWS credentials: %w", err)
	}

	return nil
}

// GetRegion returns the configured AWS region.
func (a *AWSProvisioner) GetRegion() string {
	if a.region != "" {
		return a.region
	}
	return "us-west-2" // default region
}

// GetProfile returns the configured AWS profile.
func (a *AWSProvisioner) GetProfile() string {
	return a.profile
}