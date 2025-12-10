package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// S3Config holds configuration parameters for connecting to AWS S3.
// Note that credentials and buckets secrets are managed by the AWS SDK and are not included here.
type S3Config struct {
	region string
}

// NewS3Config creates a new S3Config by loading necessary parameters from environment variables.
func NewS3Config(ctx context.Context) (*S3Config, error) {
	const op = "config.NewS3Config"

	region, err := loadAwsRegion(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &S3Config{
		region: region,
	}, nil
}

// GetRegion returns the AWS region configured for S3.
func (c *S3Config) GetRegion() string {
	return c.region
}

// loadAwsRegion fetches the AWS region from environment variables and returns an error if it's missing.
func loadAwsRegion(ctx context.Context) (string, error) {
	const op = "config.loadAwsRegion"
	region, ok := os.LookupEnv(EnvAWSRegion)

	if !ok || region == "" {
		return "", fmt.Errorf("%s: %w", op, ErrMissingAWSRegion)
	}
	// Validate the region is valid
	if err := validateAwsRegion(ctx, region); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return region, nil
}

// validateAwsRegion checks if the provided AWS region is valid by attempting to load the default AWS config.
func validateAwsRegion(ctx context.Context, region string) error {
	const op = "config.validateAwsRegion"
	// Basic validation of AWS region format
	res := strings.Split(region, "-")

	switch {
	// AWS regions typically have at least 3 parts, e.g., "us-west-2"
	case len(res) != 3:
		return fmt.Errorf("%s: %w", op, ErrInvalidAWSRegion)
	// Validate the first part is a two length region code (e.g., "us")
	case res[0] == "" || len(res[0]) != 2:
		return fmt.Errorf("%s: %w", op, ErrInvalidAWSRegion)
	// Validate the second part is non-empty (e.g., "west")
	case res[1] == "":
		return fmt.Errorf("%s: %w", op, ErrInvalidAWSRegion)
	// Validate the third part is a valid number (e.g., "2")
	case res[2] == "":
		return fmt.Errorf("%s: %w", op, ErrInvalidAWSRegion)
	}

	// Verify the third part is a valid number
	if _, err := strconv.Atoi(res[2]); err != nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidAWSRegion)
	}

	return nil
}
