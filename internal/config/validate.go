package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// validateConfig validates the entire configuration.
func validateConfig(cfg *Config) error {
	if err := validateBackupDirs(cfg.BackupDirs); err != nil {
		return err
	}

	if err := validateAWSConfig(cfg.AWSRegion, cfg.S3Bucket); err != nil {
		return err
	}

	return nil
}

// validateBackupDirs ensures backup directories are configured and exist.
func validateBackupDirs(dirs []string) error {
	if len(dirs) == 0 {
		return fmt.Errorf("%w (set %s or configure in YAML)", ErrNoBackupDirs, EnvBackupDirs)
	}

	for _, dir := range dirs {
		if err := validateDirectory(dir); err != nil {
			return err
		}
	}

	return nil
}

// validateDirectory checks if a directory exists and is accessible.
func validateDirectory(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("backup directory %s: %w", dir, ErrInvalidDir)
	}

	if !fi.IsDir() {
		return fmt.Errorf("backup directory %s: %w", dir, ErrInvalidDir)
	}

	return nil
}

// validateAWSConfig ensures AWS region and S3 bucket are configured and valid.
func validateAWSConfig(region, bucket string) error {
	if region == "" {
		return fmt.Errorf("%w (set %s or configure in YAML)", ErrMissingAWSRegion, EnvAWSRegion)
	}

	if err := validateAWSRegion(region); err != nil {
		return err
	}

	if bucket == "" {
		return fmt.Errorf("%w (set %s or configure in YAML)", ErrMissingS3BucketName, EnvS3Bucket)
	}

	return nil
}

// validateAWSRegion checks if the AWS region format is valid.
// AWS regions follow the pattern: {code}-{direction}-{number} (e.g., us-west-2)
func validateAWSRegion(region string) error {
	parts := strings.Split(region, "-")

	if len(parts) != 3 {
		return fmt.Errorf("%w: expected format {code}-{direction}-{number}", ErrInvalidAWSRegion)
	}

	// Validate region code (e.g., "us", "eu", "ap")
	if len(parts[0]) != 2 || parts[0] == "" {
		return fmt.Errorf("%w: invalid region code", ErrInvalidAWSRegion)
	}

	// Validate direction (e.g., "east", "west", "central")
	if parts[1] == "" {
		return fmt.Errorf("%w: invalid direction", ErrInvalidAWSRegion)
	}

	// Validate zone number
	if parts[2] == "" {
		return fmt.Errorf("%w: invalid zone number", ErrInvalidAWSRegion)
	}

	if _, err := strconv.Atoi(parts[2]); err != nil {
		return fmt.Errorf("%w: zone must be a number", ErrInvalidAWSRegion)
	}

	return nil
}
