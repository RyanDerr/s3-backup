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

	tc := map[string]struct {
		setup         func(t *testing.T)
		wantErr       bool
		wantRecursive bool
	}{
		"from environment variables": {
			setup: func(t *testing.T) {
				setupConfigFromEnv(t, 2)
			},
		},
		"from environment variables with recursive enabled": {
			setup: func(t *testing.T) {
				setupConfigFromEnv(t, 2)
				setupEnv(t, EnvRecursive, "true")
			},
			wantRecursive: true,
		},
		"from environment variables with recursive disabled": {
			setup: func(t *testing.T) {
				setupConfigFromEnv(t, 2)
				setupEnv(t, EnvRecursive, "false")
			},
			wantRecursive: false,
		},
		"from YAML file": {
			setup: func(t *testing.T) {
				setupConfigFromYAML(t, 2, false)
			},
		},
		"from YAML file with recursive enabled": {
			setup: func(t *testing.T) {
				setupConfigFromYAML(t, 2, true)
			},
			wantRecursive: true,
		},
		"env vars override YAML": {
			setup: func(t *testing.T) {
				setupConfigFromYAML(t, 1, false)
				setupConfigFromEnv(t, 2) // Override
			},
		},
		"env recursive overrides YAML recursive": {
			setup: func(t *testing.T) {
				setupConfigFromYAML(t, 1, true) // YAML has recursive=true
				setupConfigFromEnv(t, 1)
				setupEnv(t, EnvRecursive, "false") // Override with env var
			},
			wantRecursive: false,
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

			got, err := NewConfig()
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
			assert.Equal(t, tc.wantRecursive, got.Recursive)
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

func TestConfig_IsRecursive(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		recursive bool
		want      bool
	}{
		"returns true when recursive is enabled": {
			recursive: true,
			want:      true,
		},
		"returns false when recursive is disabled": {
			recursive: false,
			want:      false,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{Recursive: tc.recursive}
			assert.Equal(t, tc.want, cfg.IsRecursive())
		})
	}
}

func TestConfig_GetCronSchedule(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		cronSchedule string
		want         string
	}{
		"returns configured cron schedule": {
			cronSchedule: "0 0 * * *",
			want:         "0 0 * * *",
		},
		"returns empty string when not configured": {
			cronSchedule: "",
			want:         "",
		},
		"returns custom schedule": {
			cronSchedule: "*/5 * * * *",
			want:         "*/5 * * * *",
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{CronSchedule: tc.cronSchedule}
			assert.Equal(t, tc.want, cfg.GetCronSchedule())
		})
	}
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
		_ = os.Unsetenv(key)
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
// Creates dirCount temporary directories and writes a complete YAML config with backup dirs, AWS region, S3 bucket, and recursive flag.
func setupConfigFromYAML(t *testing.T, dirCount int, recursive bool) {
	t.Helper()
	dirs := createTempDirs(t, dirCount)

	var yamlContent strings.Builder
	yamlContent.WriteString("backup_dirs:\n")
	for _, dir := range dirs {
		yamlContent.WriteString(fmt.Sprintf("  - %s\n", dir))
	}
	yamlContent.WriteString("aws_region: eu-west-1\n")
	yamlContent.WriteString("s3_bucket: yaml-bucket\n")
	yamlContent.WriteString(fmt.Sprintf("recursive: %v\n", recursive))

	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent.String()), 0600)
	require.NoError(t, err)

	setupEnv(t, EnvConfigFile, tmpFile)
}
