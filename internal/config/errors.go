package config

import "errors"

var (
	// ErrNoBackupDirs is returned when no backup directories are configured.
	ErrNoBackupDirs = errors.New("no backup directories configured")
	// ErrInvalidDir is returned when a directory does not exist or is not a directory.
	ErrInvalidDir = errors.New("directory does not exist or is not a directory")

	// ErrMissingAWSRegion is returned when AWS region is not configured.
	ErrMissingAWSRegion = errors.New("missing AWS region")
	// ErrInvalidAWSRegion is returned when AWS region format is invalid.
	ErrInvalidAWSRegion = errors.New("invalid AWS region format")
	// ErrMissingS3BucketName is returned when S3 bucket name is not configured.
	ErrMissingS3BucketName = errors.New("missing S3 bucket name")
	// ErrInvalidConfigFile is returned when configuration file is invalid.
	ErrInvalidConfigFile = errors.New("invalid configuration file")
)
