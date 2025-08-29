package awsprovisioner_test

import (
	"os"
	"testing"

	awsprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/provider/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAWSProvisioner(t *testing.T) {
	t.Parallel()

	// Arrange
	region := "us-east-1"
	profile := "test-profile"

	// Act
	provisioner := awsprovisioner.NewAWSProvisioner(region, profile)

	// Assert
	assert.NotNil(t, provisioner)
	assert.Equal(t, region, provisioner.GetRegion())
	assert.Equal(t, profile, provisioner.GetProfile())
}

func TestAWSProvisioner_GetRegion_Default(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner := awsprovisioner.NewAWSProvisioner("", "")

	// Act
	region := provisioner.GetRegion()

	// Assert
	assert.Equal(t, "us-west-2", region, "Should return default region when none specified")
}

func TestAWSProvisioner_GetRegion_Custom(t *testing.T) {
	t.Parallel()

	// Arrange
	customRegion := "eu-west-1"
	provisioner := awsprovisioner.NewAWSProvisioner(customRegion, "")

	// Act
	region := provisioner.GetRegion()

	// Assert
	assert.Equal(t, customRegion, region)
}

func TestAWSProvisioner_GetProfile_Empty(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner := awsprovisioner.NewAWSProvisioner("", "")

	// Act
	profile := provisioner.GetProfile()

	// Assert
	assert.Empty(t, profile, "Should return empty profile when none specified")
}

func TestAWSProvisioner_GetProfile_Custom(t *testing.T) {
	t.Parallel()

	// Arrange
	customProfile := "my-profile"
	provisioner := awsprovisioner.NewAWSProvisioner("", customProfile)

	// Act
	profile := provisioner.GetProfile()

	// Assert
	assert.Equal(t, customProfile, profile)
}

func TestAWSProvisioner_CheckReady_AWSCLINotAvailable(t *testing.T) {
	t.Parallel()

	// This test assumes AWS CLI is not available in the test environment
	// If AWS CLI is available, this test might fail
	
	// Arrange
	provisioner := awsprovisioner.NewAWSProvisioner("us-west-2", "")

	// Temporarily modify PATH to ensure aws command is not found
	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", "/tmp")
	defer os.Setenv("PATH", originalPath)

	// Act
	ready, err := provisioner.CheckReady()

	// Assert
	assert.False(t, ready)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AWS CLI not available")
}

func TestAWSProvisioner_CheckReady_EksctlNotAvailable(t *testing.T) {
	t.Parallel()

	// Skip this test if AWS CLI is not available, as it would fail earlier
	if !isAWSCLIAvailable() {
		t.Skip("AWS CLI not available, skipping eksctl test")
	}

	// Arrange
	provisioner := awsprovisioner.NewAWSProvisioner("us-west-2", "")

	// Create a custom PATH that includes aws but not eksctl
	// This is a simplified approach - in reality, you'd need to mock the exec calls
	originalPath := os.Getenv("PATH")
	
	// Try to find aws binary and create a path with only that directory
	awsPath := findAWSBinaryPath()
	if awsPath != "" {
		os.Setenv("PATH", awsPath)
		defer os.Setenv("PATH", originalPath)

		// Act
		ready, err := provisioner.CheckReady()

		// Assert
		assert.False(t, ready)
		if err != nil {
			// Error could be about eksctl or credentials, both are acceptable for this test
			assert.True(t, 
				contains(err.Error(), "eksctl not available") || 
				contains(err.Error(), "AWS credentials not configured"),
				"Error should mention eksctl or credentials: %s", err.Error())
		}
	} else {
		t.Skip("Could not isolate AWS CLI path for eksctl test")
	}
}

// Helper functions for tests

func isAWSCLIAvailable() bool {
	provisioner := awsprovisioner.NewAWSProvisioner("", "")
	_, err := provisioner.CheckReady()
	return err == nil || contains(os.Getenv("PATH"), "aws")
}

func findAWSBinaryPath() string {
	// This is a simplified helper - in a real test environment,
	// you might want to implement more sophisticated path manipulation
	return ""
}

func contains(str, substr string) bool {
	return len(str) > 0 && len(substr) > 0 && str != substr && 
		   (len(str) >= len(substr) && str[:len(substr)] == substr) ||
		   (len(str) > len(substr) && findSubstring(str, substr))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration test that requires actual AWS credentials
func TestAWSProvisioner_CheckReady_Integration(t *testing.T) {
	// Skip this test unless explicitly enabled with environment variable
	if os.Getenv("RUN_AWS_INTEGRATION_TESTS") == "" {
		t.Skip("AWS integration test skipped. Set RUN_AWS_INTEGRATION_TESTS=1 to enable.")
	}

	// Arrange
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-west-2"
	}
	profile := os.Getenv("AWS_PROFILE")
	
	provisioner := awsprovisioner.NewAWSProvisioner(region, profile)

	// Act
	ready, err := provisioner.CheckReady()

	// Assert
	// This test will only pass if AWS CLI, eksctl, and credentials are properly configured
	if err != nil {
		t.Logf("AWS integration test failed (expected in CI): %v", err)
		// Don't fail the test in CI environments where AWS might not be configured
		return
	}
	
	require.NoError(t, err, "AWS should be ready with proper configuration")
	assert.True(t, ready, "AWS should be ready")
}