package s3

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Config(t *testing.T) {
	// Not run in parallel because it modifies global environment variables,
	// which would cause race conditions with other tests.
	ctx := context.Background()

	tc := map[string]struct {
		setup   func(t *testing.T)
		wantErr error
		expect  *S3Config
	}{
		"happy path": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west-2")
				setupEnv(t, EnvS3Bucket, "test-bucket")
			},
			expect: &S3Config{
				bucketName: "test-bucket",
			},
		},
		"missing AWS region": {
			wantErr: ErrMissingAWSRegion,
		},
		"missing bucket name": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west-2")
			},
			wantErr: ErrMissingS3BucketName,
		},
		"invalid AWS region": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "fake region")
			},
			wantErr: ErrInvalidAWSRegion,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			// Setup environment variables with automatic cleanup via t.Cleanup
			if tc.setup != nil {
				tc.setup(t)
			}

			got, err := NewS3Config(ctx)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, got)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.NotNil(t, got.Config)
			assert.Equal(t, tc.expect.bucketName, got.bucketName)
		})
	}
}

func TestGetRegion(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		config *S3Config
		expect string
	}{
		"happy path": {
			config: &S3Config{Config: &aws.Config{Region: "us-west-2"}},
			expect: "us-west-2",
		},
		"empty region": {
			config: &S3Config{Config: &aws.Config{Region: ""}},
			expect: "",
		},
		"different region": {
			config: &S3Config{Config: &aws.Config{Region: "eu-west-1"}},
			expect: "eu-west-1",
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := tc.config.GetRegion()
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestGetBucketName(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		config *S3Config
		expect string
	}{
		"happy path": {
			config: &S3Config{bucketName: "my-test-bucket"},
			expect: "my-test-bucket",
		},
		"empty bucket name": {
			config: &S3Config{bucketName: ""},
			expect: "",
		},
		"different bucket name": {
			config: &S3Config{bucketName: "production-backup-bucket"},
			expect: "production-backup-bucket",
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := tc.config.GetBucketName()
			assert.Equal(t, tc.expect, got)
		})
	}
}

func Test_loadAwsRegion(t *testing.T) {
	// Not run in parallel because it modifies global environment variables,
	// which would cause race conditions with other tests.
	ctx := context.Background()

	tc := map[string]struct {
		setup   func(t *testing.T)
		wantErr error
		expect  string
	}{
		"happy path": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-east-1")
			},
			expect: "us-east-1",
		},
		"missing environment variable": {
			wantErr: ErrMissingAWSRegion,
		},
		"empty environment variable": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "")
			},
			wantErr: ErrMissingAWSRegion,
		},
		"invalid region format - too few parts": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west")
			},
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid region format - too many parts": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west-2-extra")
			},
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid region format - bad number": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west-abc")
			},
			wantErr: ErrInvalidAWSRegion,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}

			got, err := loadAwsRegion(ctx)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Empty(t, got)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expect, got)
		})
	}
}

func Test_validateAwsRegion(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		region  string
		wantErr error
	}{
		"happy path - us-east-1": {
			region: "us-east-1",
		},
		"happy path - us-west-2": {
			region: "us-west-2",
		},
		"happy path - eu-west-1": {
			region: "eu-west-1",
		},
		"happy path - ap-south-1": {
			region: "ap-south-1",
		},
		"invalid - too few parts": {
			region:  "us-west",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - too many parts": {
			region:  "us-west-2-extra",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - empty string": {
			region:  "",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - first part empty": {
			region:  "-west-2",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - first part wrong length": {
			region:  "usa-west-2",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - second part empty": {
			region:  "us--2",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - third part empty": {
			region:  "us-west-",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - third part not a number": {
			region:  "us-west-abc",
			wantErr: ErrInvalidAWSRegion,
		},
		"invalid - contains special characters": {
			region:  "us-west-2!",
			wantErr: ErrInvalidAWSRegion,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := validateAwsRegion(tc.region)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func Test_loadBucketName(t *testing.T) {
	// Not run in parallel because it modifies global environment variables,
	// which would cause race conditions with other tests.

	tc := map[string]struct {
		setup   func(t *testing.T)
		wantErr error
		expect  string
	}{
		"happy path": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvS3Bucket, "my-test-bucket")
			},
			expect: "my-test-bucket",
		},
		"missing environment variable": {
			wantErr: ErrMissingS3BucketName,
		},
		"empty environment variable": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvS3Bucket, "")
			},
			wantErr: ErrMissingS3BucketName,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t)
			}

			got, err := loadBucketName()
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Empty(t, got)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expect, got)
		})
	}
}

// setupEnv sets an environment variable for testing and registers a cleanup function via t.Cleanup.
// This ensures the environment variable is properly unset after the test completes, even if the test fails.
func setupEnv(t *testing.T, key, value string) {
	t.Helper()

	// Set the new value for the test
	err := os.Setenv(key, value)
	require.NoError(t, err)

	// Register cleanup function that will run automatically
	t.Cleanup(func() {
		// Unset the variable if it didn't exist before
		os.Unsetenv(key)
	})
}
