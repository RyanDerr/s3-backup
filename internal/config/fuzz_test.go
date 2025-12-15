package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzLoadFromFile tests YAML parsing with fuzzy input
func FuzzLoadFromFile(f *testing.F) {
	// Seed corpus with valid YAML examples
	f.Add(`backup_dirs:
  - /tmp/test
aws_region: us-west-2
s3_bucket: test-bucket`)

	f.Add(`backup_dirs: ["/tmp/a", "/tmp/b"]
recursive: true
cron_schedule: "0 0 * * *"
aws_region: us-east-1
s3_bucket: my-bucket`)

	f.Add(`{}`) // Empty YAML

	f.Add(`backup_dirs: []
aws_region: ""
s3_bucket: ""`)

	f.Fuzz(func(t *testing.T, yamlContent string) {
		// Create temporary file with fuzzy YAML content
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
			t.Skip("Failed to write test file")
		}

		// Set environment variable
		t.Setenv(EnvConfigFile, configFile)

		cfg := &Config{}

		// loadFromFile should not panic with any input
		_ = loadFromFile(cfg)

		// Basic sanity checks - if data was parsed, it should be valid types
		if cfg.BackupDirs != nil {
			for _, dir := range cfg.BackupDirs {
				// Directory paths should be strings (not panic on nil)
				_ = len(dir)
			}
		}

		if cfg.AWSRegion != "" {
			// Should be a string
			_ = len(cfg.AWSRegion)
		}

		if cfg.S3Bucket != "" {
			// Should be a string
			_ = len(cfg.S3Bucket)
		}
	})
}

// FuzzLoadFromEnv tests environment variable parsing with fuzzy input
func FuzzLoadFromEnv(f *testing.F) {
	// Seed corpus with various directory formats
	f.Add("/tmp/test", "us-west-2", "my-bucket", "true", "0 0 * * *")
	f.Add("/tmp/a,/tmp/b,/tmp/c", "us-east-1", "test", "false", "*/5 * * * *")
	f.Add("", "", "", "", "")
	f.Add("/", "invalid", "a", "maybe", "invalid cron")
	f.Add("../../../etc/passwd", "us-west-2", "'; DROP TABLE users;--", "1", "@daily")

	f.Fuzz(func(t *testing.T, dirs, region, bucket, recursive, cronSchedule string) {
		t.Setenv(EnvBackupDirs, dirs)
		t.Setenv(EnvAWSRegion, region)
		t.Setenv(EnvS3Bucket, bucket)
		t.Setenv(EnvRecursive, recursive)
		t.Setenv(EnvCronSchedule, cronSchedule)

		cfg := &Config{}

		// Should not panic with any input
		loadFromEnv(cfg)

		// Verify directory parsing doesn't crash
		if cfg.BackupDirs != nil {
			for _, dir := range cfg.BackupDirs {
				_ = len(dir)
			}
		}

		// Verify boolean parsing is safe
		_ = cfg.Recursive

		// Verify strings are safe
		_ = len(cfg.AWSRegion)
		_ = len(cfg.S3Bucket)
		_ = len(cfg.CronSchedule)
	})
}

// FuzzValidateAWSRegion tests AWS region validation with fuzzy input
func FuzzValidateAWSRegion(f *testing.F) {
	// Seed with valid and invalid regions
	f.Add("us-west-2")
	f.Add("us-east-1")
	f.Add("eu-west-1")
	f.Add("invalid-region")
	f.Add("")
	f.Add("us-west-2\x00")
	f.Add("../us-west-2")
	f.Add("us-west-2; rm -rf /")
	f.Add(strings.Repeat("a", 1000))

	f.Fuzz(func(t *testing.T, region string) {
		cfg := &Config{
			AWSRegion:  region,
			S3Bucket:   "test-bucket",
			BackupDirs: []string{t.TempDir()},
		}

		// Should not panic regardless of input
		_ = validateConfig(cfg)
	})
}

// FuzzValidateS3Bucket tests S3 bucket validation with fuzzy input
func FuzzValidateS3Bucket(f *testing.F) {
	// Seed with valid and invalid bucket names
	f.Add("valid-bucket-name")
	f.Add("my-bucket")
	f.Add("")
	f.Add("UpperCase")
	f.Add("bucket_with_underscore")
	f.Add("a")
	f.Add(strings.Repeat("a", 64))
	f.Add("../../../etc/passwd")
	f.Add("bucket'; DROP TABLE;--")
	f.Add("bucket\x00name")

	f.Fuzz(func(t *testing.T, bucket string) {
		cfg := &Config{
			AWSRegion:  "us-west-2",
			S3Bucket:   bucket,
			BackupDirs: []string{t.TempDir()},
		}

		// Should not panic regardless of input
		_ = validateConfig(cfg)
	})
}
