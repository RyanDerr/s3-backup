package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

// Config holds all application configuration including backup directories and AWS S3 settings.
type Config struct {
	// Backup configuration
	BackupDirs   []string `yaml:"backup_dirs"`
	Recursive    bool     `yaml:"recursive"`
	CronSchedule string   `yaml:"cron_schedule"`

	// AWS S3 configuration
	AWSRegion string `yaml:"aws_region"`
	S3Bucket  string `yaml:"s3_bucket"`

	sync.RWMutex
}

// NewConfig creates a new Config by loading from YAML file or environment variables.
// Environment variables take precedence over YAML configuration.
func NewConfig(ctx context.Context) (*Config, error) {
	const op = "config.NewConfig"

	cfg := &Config{}

	// Load from YAML file if specified
	if err := loadFromFile(cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Environment variables override YAML
	loadFromEnv(cfg)

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return cfg, nil
}

// GetBackupDirs returns a copy of the configured backup directories.
func (c *Config) GetBackupDirs() []string {
	c.RLock()
	defer c.RUnlock()

	dirs := make([]string, len(c.BackupDirs))
	copy(dirs, c.BackupDirs)
	return dirs
}

// GetAWSRegion returns the configured AWS region.
func (c *Config) GetAWSRegion() string {
	c.RLock()
	defer c.RUnlock()
	return c.AWSRegion
}

// GetS3Bucket returns the configured S3 bucket name.
func (c *Config) GetS3Bucket() string {
	c.RLock()
	defer c.RUnlock()
	return c.S3Bucket
}

// IsRecursive returns whether we should perform recursive backup of nested directories and files.
func (c *Config) IsRecursive() bool {
	c.RLock()
	defer c.RUnlock()
	return c.Recursive
}

// GetCronSchedule returns the configured cron schedule.
// Returns DefaultCronSchedule if not configured.
func (c *Config) GetCronSchedule() string {
	c.RLock()
	defer c.RUnlock()
	if c.CronSchedule == "" {
		return DefaultCronSchedule
	}
	return c.CronSchedule
}

// GetAWSConfig loads and returns the AWS SDK config with the configured region.
func (c *Config) GetAWSConfig(ctx context.Context) (aws.Config, error) {
	c.RLock()
	defer c.RUnlock()
	region := c.AWSRegion

	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file if EnvConfigFile is set.
func loadFromFile(cfg *Config) error {
	configFile := os.Getenv(EnvConfigFile)
	if configFile == "" {
		return nil
	}

	if err := loadFromYaml(configFile, cfg); err != nil {
		return fmt.Errorf("failed to load YAML config: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables.
// Environment variables override any values loaded from YAML.
func loadFromEnv(cfg *Config) {
	// Load backup directories
	if envDirs := os.Getenv(EnvBackupDirs); envDirs != "" {
		cfg.BackupDirs = parseCommaSeparated(envDirs)
	}

	// Load recursive flag
	if recursive := os.Getenv(EnvRecursive); recursive != "" {
		cfg.Recursive = strings.ToLower(recursive) == "true"
	}

	// Load cron schedule
	if cronSchedule := os.Getenv(EnvCronSchedule); cronSchedule != "" {
		cfg.CronSchedule = cronSchedule
	}

	// Load AWS region
	if region := os.Getenv(EnvAWSRegion); region != "" {
		cfg.AWSRegion = region
	}

	// Load S3 bucket
	if bucket := os.Getenv(EnvS3Bucket); bucket != "" {
		cfg.S3Bucket = bucket
	}
}

// parseCommaSeparated parses a comma-separated string into a slice,
// trimming whitespace and filtering out empty strings.
func parseCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
