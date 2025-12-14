package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAWSRegion(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		region  string
		wantErr bool
	}{
		"valid us-east-1":          {region: "us-east-1"},
		"valid us-west-2":          {region: "us-west-2"},
		"valid eu-west-1":          {region: "eu-west-1"},
		"valid ap-south-1":         {region: "ap-south-1"},
		"invalid too few parts":    {region: "us-west", wantErr: true},
		"invalid too many parts":   {region: "us-west-2-extra", wantErr: true},
		"invalid empty":            {region: "", wantErr: true},
		"invalid code length":      {region: "usa-west-2", wantErr: true},
		"invalid empty direction":  {region: "us--2", wantErr: true},
		"invalid non-numeric zone": {region: "us-west-abc", wantErr: true},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validateAWSRegion(tc.region)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidAWSRegion)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateBackupDirs(t *testing.T) {
	t.Parallel()

	t.Run("empty directories", func(t *testing.T) {
		t.Parallel()
		err := validateBackupDirs([]string{})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoBackupDirs)
	})

	t.Run("nil directories", func(t *testing.T) {
		t.Parallel()
		err := validateBackupDirs(nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoBackupDirs)
	})

	t.Run("valid directories", func(t *testing.T) {
		t.Parallel()
		dirs := createTempDirs(t, 2)
		err := validateBackupDirs(dirs)
		require.NoError(t, err)
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()
		err := validateBackupDirs([]string{"/nonexistent/directory"})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidDir)
	})
}

func TestValidateDirectory(t *testing.T) {
	t.Parallel()

	t.Run("valid directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		err := validateDirectory(dir)
		require.NoError(t, err)
	})

	t.Run("nonexistent path", func(t *testing.T) {
		t.Parallel()
		err := validateDirectory("/nonexistent/path")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidDir)
	})
}

func TestValidateAWSConfig(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		region  string
		bucket  string
		wantErr error
	}{
		"valid config": {
			region: "us-west-2",
			bucket: "my-bucket",
		},
		"missing region": {
			region:  "",
			bucket:  "my-bucket",
			wantErr: ErrMissingAWSRegion,
		},
		"missing bucket": {
			region:  "us-west-2",
			bucket:  "",
			wantErr: ErrMissingS3BucketName,
		},
		"invalid region": {
			region:  "invalid",
			bucket:  "my-bucket",
			wantErr: ErrInvalidAWSRegion,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validateAWSConfig(tc.region, tc.bucket)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BackupDirs: createTempDirs(t, 1),
			AWSRegion:  "us-east-1",
			S3Bucket:   "test-bucket",
		}
		err := validateConfig(cfg)
		require.NoError(t, err)
	})

	t.Run("missing backup dirs", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			AWSRegion: "us-east-1",
			S3Bucket:  "test-bucket",
		}
		err := validateConfig(cfg)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoBackupDirs)
	})

	t.Run("missing AWS region", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BackupDirs: createTempDirs(t, 1),
			S3Bucket:   "test-bucket",
		}
		err := validateConfig(cfg)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMissingAWSRegion)
	})

	t.Run("invalid directory", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			BackupDirs: []string{"/nonexistent"},
			AWSRegion:  "us-east-1",
			S3Bucket:   "test-bucket",
		}
		err := validateConfig(cfg)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidDir)
	})
}
