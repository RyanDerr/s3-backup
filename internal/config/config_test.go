package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	// Not run in parallel because it modifies global environment variables
	ctx := context.Background()

	tc := map[string]struct {
		setup   func(t *testing.T)
		wantErr bool
	}{
		"from environment variables": {
			setup: func(t *testing.T) {
				setupConfigFromEnv(t, 2)
			},
		},
		"from YAML file": {
			setup: func(t *testing.T) {
				setupConfigFromYAML(t, 2)
			},
		},
		"env vars override YAML": {
			setup: func(t *testing.T) {
				setupConfigFromYAML(t, 1)
				setupConfigFromEnv(t, 2) // Override
			},
		},
		"missing backup dirs": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west-2")
				setupEnv(t, EnvS3Bucket, "test-bucket")
			},
			wantErr: true,
		},
		"missing AWS region": {
			setup: func(t *testing.T) {
				setupEnvWithDirs(t, 1)
				setupEnv(t, EnvS3Bucket, "test-bucket")
			},
			wantErr: true,
		},
		"missing S3 bucket": {
			setup: func(t *testing.T) {
				setupEnvWithDirs(t, 1)
				setupEnv(t, EnvAWSRegion, "us-west-2")
			},
			wantErr: true,
		},
		"invalid AWS region": {
			setup: func(t *testing.T) {
				setupEnvWithDirs(t, 1)
				setupEnv(t, EnvAWSRegion, "invalid-region")
				setupEnv(t, EnvS3Bucket, "test-bucket")
			},
			wantErr: true,
		},
		"directory does not exist": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvBackupDirs, "/nonexistent/path")
				setupEnv(t, EnvAWSRegion, "us-west-2")
				setupEnv(t, EnvS3Bucket, "test-bucket")
			},
			wantErr: true,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}

			got, err := NewConfig(ctx)
			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.NotEmpty(t, got.BackupDirs)
			assert.NotEmpty(t, got.AWSRegion)
			assert.NotEmpty(t, got.S3Bucket)
		})
	}
}

func TestConfig_GetBackupDirs(t *testing.T) {
	t.Parallel()

	t.Run("returns configured directories", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{BackupDirs: []string{"/dir1", "/dir2"}}

		result := cfg.GetBackupDirs()

		assert.Equal(t, []string{"/dir1", "/dir2"}, result)
	})

	t.Run("returns a copy not a reference", func(t *testing.T) {
		t.Parallel()
		original := []string{"/dir1", "/dir2"}
		cfg := &Config{BackupDirs: original}

		returned := cfg.GetBackupDirs()
		returned[0] = "/modified"

		// The original config should not be affected by changes to returned slice
		assert.Equal(t, "/dir1", cfg.BackupDirs[0], "modifying returned slice should not affect original")
		assert.Equal(t, original, cfg.BackupDirs, "original config should remain unchanged")
	})
}

func TestConfig_GetAWSRegion(t *testing.T) {
	t.Parallel()

	cfg := &Config{AWSRegion: "us-east-1"}
	assert.Equal(t, "us-east-1", cfg.GetAWSRegion())
}

func TestConfig_GetS3Bucket(t *testing.T) {
	t.Parallel()

	cfg := &Config{S3Bucket: "my-bucket"}
	assert.Equal(t, "my-bucket", cfg.GetS3Bucket())
}

func TestConfig_GetAWSConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := &Config{AWSRegion: "us-west-2"}

	awsCfg, err := cfg.GetAWSConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, "us-west-2", awsCfg.Region)
}

// setupEnv sets an environment variable for the duration of the test.
// The variable is automatically cleaned up after the test completes.
func setupEnv(t *testing.T, key, value string) {
	t.Helper()
	err := os.Setenv(key, value)
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Unsetenv(key)
	})
}

// createTempDirs creates multiple temporary directories for testing.
// Each directory is automatically cleaned up after the test completes.
func createTempDirs(t *testing.T, count int) []string {
	t.Helper()
	dirs := make([]string, count)
	for i := 0; i < count; i++ {
		dirs[i] = t.TempDir()
	}
	return dirs
}

// setupEnvWithDirs creates temporary directories and sets the BACKUP_DIRS environment variable.
// Creates count temporary directories, joins them with commas, and sets EnvBackupDirs.
func setupEnvWithDirs(t *testing.T, count int) {
	t.Helper()
	dirs := createTempDirs(t, count)
	setupEnv(t, EnvBackupDirs, strings.Join(dirs, ","))
}

// setupConfigFromEnv sets up a complete configuration using environment variables.
// Creates dirCount temporary directories and sets all required env vars (backup dirs, AWS region, S3 bucket).
func setupConfigFromEnv(t *testing.T, dirCount int) {
	t.Helper()
	setupEnvWithDirs(t, dirCount)
	setupEnv(t, EnvAWSRegion, "us-west-2")
	setupEnv(t, EnvS3Bucket, "test-bucket")
}

// setupConfigFromYAML creates a YAML configuration file and sets the config file path.
// Creates dirCount temporary directories and writes a complete YAML config with backup dirs, AWS region, and S3 bucket.
func setupConfigFromYAML(t *testing.T, dirCount int) {
	t.Helper()
	dirs := createTempDirs(t, dirCount)

	var yamlContent strings.Builder
	yamlContent.WriteString("backup_dirs:\n")
	for _, dir := range dirs {
		yamlContent.WriteString(fmt.Sprintf("  - %s\n", dir))
	}
	yamlContent.WriteString("aws_region: eu-west-1\n")
	yamlContent.WriteString("s3_bucket: yaml-bucket\n")

	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent.String()), 0644)
	require.NoError(t, err)

	setupEnv(t, EnvConfigFile, tmpFile)
}
