package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Config(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup   func(t *testing.T)
		wantErr error
		expect  *S3Config
	}{
		"happy path": {
			setup: func(t *testing.T) {
				setupEnv(t, EnvAWSRegion, "us-west-2")
			},
			expect: &S3Config{
				region: "us-west-2",
			},
		},
		"missing AWS region": {
			wantErr: ErrMissingAWSRegion,
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
			assert.EqualValues(t, tc.expect, got)
		})
	}
}

func TestGetRegion(t *testing.T) {
	t.Parallel()
}

func Test_loadAwsRegion(t *testing.T) {
	t.Parallel()
}

func Test_validateAwsRegion(t *testing.T) {
	t.Parallel()

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
