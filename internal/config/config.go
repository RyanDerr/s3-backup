package config

import (
	"fmt"
	"os"
)

// S3Config holds the configuration for connecting to an S3 bucket.
type S3Config struct {
	Region          string
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
}

// NewS3Config creates a new S3Config with default values.
func NewS3Config() (*S3Config, error) {
	const op = "config.NewS3Config"
	var conf S3Config

	region, err := loadAwsRegion()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	conf.Region = region

	return &conf, nil
}

// loadAwsRegion fetches the AWS region from environment variables and returns an error if it's missing.
func loadAwsRegion() (string, error) {
	const op = "config.loadAwsRegion"
	if res := os.Getenv(EnvAWSRegion); res != "" {
		return res, nil
	}
	return "", fmt.Errorf("%s: %w", op, ErrMissingAWSRegion)
}
