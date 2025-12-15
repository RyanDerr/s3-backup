package config

import "errors"

var (
	// Backup errors
	ErrNoBackupDirs = errors.New("no backup directories configured")
	ErrInvalidDir   = errors.New("directory does not exist or is not a directory")

	// AWS/S3 errors
	ErrMissingAWSRegion    = errors.New("missing AWS region")
	ErrInvalidAWSRegion    = errors.New("invalid AWS region format")
	ErrMissingS3BucketName = errors.New("missing S3 bucket name")
	ErrInvalidConfigFile   = errors.New("invalid configuration file")
)
