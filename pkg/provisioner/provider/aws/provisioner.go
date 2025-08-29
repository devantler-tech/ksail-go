// Package awsprovisioner provides an AWS implementation of the ProviderProvisioner interface.
package awsprovisioner

import (
	"context"
	"fmt"

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

// CheckReady checks if AWS credentials are available and valid.
func (a *AWSProvisioner) CheckReady() (bool, error) {
	// Check if AWS credentials are configured and valid
	if err := a.checkCredentials(); err != nil {
		return false, fmt.Errorf("AWS credentials not configured or invalid: %w", err)
	}

	return true, nil
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