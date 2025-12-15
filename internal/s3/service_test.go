package s3

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"s3-backup/internal/config"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewS3Service(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup   func(t *testing.T) *config.Config
		wantErr error
	}{
		"valid config with single directory": {
			setup: func(t *testing.T) *config.Config {
				return createTestConfig(t, 1, false)
			},
		},
		"valid config with multiple directories": {
			setup: func(t *testing.T) *config.Config {
				return createTestConfig(t, 3, false)
			},
		},
		"valid config with recursive enabled": {
			setup: func(t *testing.T) *config.Config {
				return createTestConfig(t, 2, true)
			},
		},
		"nil config": {
			setup: func(_ *testing.T) *config.Config {
				return nil
			},
			wantErr: ErrNilConfig,
		},
		"nonexistent directory": {
			setup: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t, 1, false)
				cfg.BackupDirs = []string{"/nonexistent/path"}
				return cfg
			},
			wantErr: ErrDirectoryNotFound,
		},
		"empty directory path": {
			setup: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t, 1, false)
				cfg.BackupDirs = append(cfg.BackupDirs, "")
				return cfg
			},
			wantErr: ErrEmptyDirectory,
		},
		"path is not a directory": {
			setup: func(t *testing.T) *config.Config {
				cfg := createTestConfig(t, 1, false)
				// Create a file instead of directory
				filePath := filepath.Join(t.TempDir(), "file.txt")
					require.NoError(t, os.WriteFile(filePath, []byte("test"), 0600))
				cfg.BackupDirs = append(cfg.BackupDirs, filePath)
				return cfg
			},
			wantErr: ErrNotADirectory,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := tc.setup(t)

			svc, err := NewS3Service(ctx, cfg)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, svc)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, svc)
			assert.NotNil(t, svc.client)
			assert.NotEmpty(t, svc.bucketName)
			assert.NotEmpty(t, svc.backupDirs)
			assert.NotEmpty(t, svc.cronSchedule)
			assert.NotNil(t, svc.stopCh)
		})
	}
}

func TestValidateDirectories(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		setup   func(t *testing.T) []string
		wantErr error
	}{
		"valid single directory": {
			setup: func(t *testing.T) []string {
				return []string{t.TempDir()}
			},
		},
		"valid multiple directories": {
			setup: func(t *testing.T) []string {
				return createTempDirs(t, 3)
			},
		},
		"empty directory path": {
			setup: func(_ *testing.T) []string {
				return []string{""}
			},
			wantErr: ErrEmptyDirectory,
		},
		"nonexistent directory": {
			setup: func(_ *testing.T) []string {
				return []string{"/nonexistent/path"}
			},
			wantErr: ErrDirectoryNotFound,
		},
		"not a directory": {
			setup: func(t *testing.T) []string {
				filePath := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(filePath, []byte("test"), 0600))
				return []string{filePath}
			},
			wantErr: ErrNotADirectory,
		},
		"mix of valid and invalid": {
			setup: func(t *testing.T) []string {
				dirs := createTempDirs(t, 1)
				return append(dirs, "/nonexistent")
			},
			wantErr: ErrDirectoryNotFound,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dirs := tc.setup(t)
			err := validateDirectories(dirs)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_GetBackupDirs(t *testing.T) {
	t.Parallel()

	t.Run("returns configured directories", func(t *testing.T) {
		t.Parallel()
		dirs := []string{"/dir1", "/dir2"}
		svc := &Service{backupDirs: dirs}

		result := svc.getBackupDirs()

		assert.Equal(t, dirs, result)
	})

	t.Run("returns a copy not a reference", func(t *testing.T) {
		t.Parallel()
		original := []string{"/dir1", "/dir2"}
		svc := &Service{backupDirs: original}

		returned := svc.getBackupDirs()
		returned[0] = "/modified"

		assert.Equal(t, "/dir1", svc.backupDirs[0], "modifying returned slice should not affect original")
		assert.Equal(t, original, svc.backupDirs, "original should remain unchanged")
	})
}

func TestService_IsRecursive(t *testing.T) {
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
			svc := &Service{recursive: tc.recursive}
			assert.Equal(t, tc.want, svc.isRecursive())
		})
	}
}

func TestService_BackupAllFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		files   []string
		wantErr bool
	}{
		"empty file list": {
			files:   []string{},
			wantErr: false,
		},
		"nil file list": {
			files:   nil,
			wantErr: false,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc := &Service{bucketName: "test-bucket"}

			err := svc.backupAllFiles(ctx, tc.files)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestService_BackupAllFiles_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	svc := &Service{bucketName: "test-bucket"}
	files := []string{"file1.txt", "file2.txt"}

	err := svc.backupAllFiles(ctx, files)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestService_BackupFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup   func(t *testing.T) (svc *Service, fileName string)
		wantErr error
	}{
		"empty filename": {
			setup: func(_ *testing.T) (*Service, string) {
				svc := &Service{
					client:     &mockS3Client{},
					bucketName: "test-bucket",
				}
				return svc, ""
			},
			wantErr: ErrEmptyFilename,
		},
		"file does not exist": {
			setup: func(_ *testing.T) (*Service, string) {
				svc := &Service{
					client:     &mockS3Client{},
					bucketName: "test-bucket",
				}
				return svc, "/nonexistent/file.txt"
			},
			wantErr: os.ErrNotExist,
		},
		"successful upload": {
			setup: func(t *testing.T) (*Service, string) {
				dir := t.TempDir()
				filePath := filepath.Join(dir, "test.txt")
				require.NoError(t, os.WriteFile(filePath, []byte("test content"), 0600))

				svc := &Service{
					client:     &mockS3Client{},
					bucketName: "test-bucket",
				}
				return svc, filePath
			},
		},
		"S3 upload fails": {
			setup: func(t *testing.T) (*Service, string) {
				dir := t.TempDir()
				filePath := filepath.Join(dir, "test.txt")
				require.NoError(t, os.WriteFile(filePath, []byte("test content"), 0600))

				svc := &Service{
					client:     &mockS3Client{shouldFail: true},
					bucketName: "test-bucket",
				}
				return svc, filePath
			},
			wantErr: errMockS3Failure,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc, fileName := tc.setup(t)
			err := svc.backupFile(ctx, fileName)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestService_BackupAllFiles_WithErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup    func(t *testing.T) (svc *Service, files []string)
		wantErr  bool
		checkErr func(t *testing.T, err error)
	}{
		"all files succeed": {
			setup: func(t *testing.T) (*Service, []string) {
				dir := t.TempDir()
				file1 := filepath.Join(dir, "file1.txt")
				file2 := filepath.Join(dir, "file2.txt")
				require.NoError(t, os.WriteFile(file1, []byte("content1"), 0600))
				require.NoError(t, os.WriteFile(file2, []byte("content2"), 0600))

				svc := &Service{
					client:     &mockS3Client{},
					bucketName: "test-bucket",
				}
				return svc, []string{file1, file2}
			},
			wantErr: false,
		},
		"some files fail": {
			setup: func(t *testing.T) (*Service, []string) {
				dir := t.TempDir()
				file1 := filepath.Join(dir, "file1.txt")
				require.NoError(t, os.WriteFile(file1, []byte("content1"), 0600))

				svc := &Service{
					client:     &mockS3Client{},
					bucketName: "test-bucket",
				}
				// Mix valid and nonexistent files
				return svc, []string{file1, "/nonexistent/file.txt", ""}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "one or more files failed")
				assert.ErrorIs(t, err, ErrEmptyFilename)
				assert.ErrorIs(t, err, os.ErrNotExist)
			},
		},
		"all files fail": {
			setup: func(t *testing.T) (*Service, []string) {
				dir := t.TempDir()
				file1 := filepath.Join(dir, "file1.txt")
				file2 := filepath.Join(dir, "file2.txt")
				require.NoError(t, os.WriteFile(file1, []byte("content1"), 0600))
				require.NoError(t, os.WriteFile(file2, []byte("content2"), 0600))

				svc := &Service{
					client:     &mockS3Client{shouldFail: true},
					bucketName: "test-bucket",
				}
				return svc, []string{file1, file2}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "one or more files failed")
				count := strings.Count(err.Error(), "mock S3 failure")
				assert.Equal(t, 2, count, "should have 2 S3 failures")
			},
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc, files := tc.setup(t)
			err := svc.backupAllFiles(ctx, files)

			if tc.wantErr {
				require.Error(t, err)
				if tc.checkErr != nil {
					tc.checkErr(t, err)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

// mockS3Client is a simple mock for testing without actual AWS calls.
type mockS3Client struct {
	shouldFail bool
}

var errMockS3Failure = errors.New("mock S3 failure")

func (m *mockS3Client) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if m.shouldFail {
		return nil, errMockS3Failure
	}

	// Consume the body to simulate reading the file
	if params.Body != nil {
		_, _ = io.Copy(io.Discard, params.Body)
	}

	return &s3.PutObjectOutput{}, nil
}

func TestService_Start(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		cronSchedule string
		wantErr      bool
	}{
		"valid cron schedule": {
			cronSchedule: "*/5 * * * *",
			wantErr:      false,
		},
		"default cron schedule": {
			cronSchedule: config.DefaultCronSchedule,
			wantErr:      false,
		},
		"invalid cron schedule": {
			cronSchedule: "invalid schedule",
			wantErr:      true,
		},
		"too many fields": {
			cronSchedule: "* * * * * * *",
			wantErr:      true,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := &Service{
				client:       &mockS3Client{},
				bucketName:   "test-bucket",
				backupDirs:   []string{t.TempDir()},
				recursive:    false,
				cronSchedule: tc.cronSchedule,
				stopCh:       make(chan struct{}),
			}

			// Run Start in a goroutine since it blocks
			errCh := make(chan error, 1)
			go func() {
				errCh <- svc.Start(ctx)
			}()

			// Give it a moment to start and potentially fail
			select {
			case err := <-errCh:
				if tc.wantErr {
					require.Error(t, err)
					assert.Contains(t, err.Error(), "invalid cron schedule")
				} else {
					// Shouldn't get error immediately for valid schedule
					t.Errorf("Start() returned unexpectedly: %v", err)
				}
			case <-time.After(100 * time.Millisecond):
				// For valid schedules, Start should be running
				if tc.wantErr {
					t.Error("Expected error but Start() is still running")
				}
				// Stop the service
				svc.Stop()
				// Wait for it to actually stop
				select {
				case err := <-errCh:
					require.NoError(t, err)
				case <-time.After(2 * time.Second):
					t.Error("Start() did not stop in time")
				}
			}
		})
	}
}

func TestService_Stop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	svc := &Service{
		client:       &mockS3Client{},
		bucketName:   "test-bucket",
		backupDirs:   []string{t.TempDir()},
		recursive:    false,
		cronSchedule: "*/5 * * * *",
		stopCh:       make(chan struct{}),
	}

	// Start the service
	errCh := make(chan error, 1)
	go func() {
		errCh <- svc.Start(ctx)
	}()

	// Wait a bit for it to start
	time.Sleep(50 * time.Millisecond)

	// Stop should close the channel and cause Start to return
	svc.Stop()

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() did not cause Start() to return in time")
	}
}

// createTestConfig creates a test config with temporary directories.
func createTestConfig(t *testing.T, dirCount int, recursive bool) *config.Config {
	t.Helper()
	return &config.Config{
		BackupDirs: createTempDirs(t, dirCount),
		AWSRegion:  "us-west-2",
		S3Bucket:   "test-bucket",
		Recursive:  recursive,
	}
}

// createTempDirs creates multiple temporary directories for testing.
func createTempDirs(t *testing.T, count int) []string {
	t.Helper()
	dirs := make([]string, count)
	for i := 0; i < count; i++ {
		dirs[i] = t.TempDir()
	}
	return dirs
}
