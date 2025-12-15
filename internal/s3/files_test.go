package s3

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectFilesFromDir(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup        func(t *testing.T) (dir string, recursive bool)
		wantMinFiles int
		wantErr      error
	}{
		"empty directory non-recursive": {
			setup: func(t *testing.T) (string, bool) {
				return t.TempDir(), false
			},
			wantMinFiles: 0,
		},
		"empty directory recursive": {
			setup: func(t *testing.T) (string, bool) {
				return t.TempDir(), true
			},
			wantMinFiles: 0,
		},
		"single file non-recursive": {
			setup: func(t *testing.T) (string, bool) {
				dir := t.TempDir()
				createFile(t, dir, "file1.txt", "content1")
				return dir, false
			},
			wantMinFiles: 1,
		},
		"multiple files non-recursive": {
			setup: func(t *testing.T) (string, bool) {
				dir := t.TempDir()
				createFile(t, dir, "file1.txt", "content1")
				createFile(t, dir, "file2.txt", "content2")
				createFile(t, dir, "file3.log", "content3")
				return dir, false
			},
			wantMinFiles: 3,
		},
		"files with subdirectory non-recursive": {
			setup: func(t *testing.T) (string, bool) {
				dir := t.TempDir()
				createFile(t, dir, "file1.txt", "content1")
				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.Mkdir(subdir, 0750))
				createFile(t, subdir, "file2.txt", "content2")
				return dir, false
			},
			wantMinFiles: 1, // Only root file
		},
		"files with subdirectory recursive": {
			setup: func(t *testing.T) (string, bool) {
				dir := t.TempDir()
				createFile(t, dir, "file1.txt", "content1")
				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.Mkdir(subdir, 0750))
				createFile(t, subdir, "file2.txt", "content2")
				return dir, true
			},
			wantMinFiles: 2, // Both files
		},
		"nested subdirectories recursive": {
			setup: func(t *testing.T) (string, bool) {
				dir := t.TempDir()
				createFile(t, dir, "root.txt", "root")

				level1 := filepath.Join(dir, "level1")
				require.NoError(t, os.Mkdir(level1, 0750))
				createFile(t, level1, "file1.txt", "content1")

				level2 := filepath.Join(level1, "level2")
				require.NoError(t, os.Mkdir(level2, 0750))
				createFile(t, level2, "file2.txt", "content2")

				return dir, true
			},
			wantMinFiles: 3,
		},
		"empty directory path": {
			setup: func(_ *testing.T) (string, bool) {
				return "", false
			},
			wantErr: ErrEmptyDirectory,
		},
		"nonexistent directory": {
			setup: func(_ *testing.T) (string, bool) {
				return "/nonexistent/path", false
			},
			wantErr: os.ErrNotExist,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir, recursive := tc.setup(t)
			svc := &Service{
				backupDirs: []string{dir},
				recursive:  recursive,
			}

			files, err := svc.collectFilesFromDir(ctx, dir, recursive)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Len(t, files, tc.wantMinFiles)

			// Verify files are prefixed with base directory name
			if len(files) > 0 {
				baseDir := filepath.Base(dir)
				for _, f := range files {
					assert.Contains(t, f, baseDir, "file should be prefixed with base directory")
				}
			}
		})
	}
}

func TestCollectFilesFromDir_ContextCancellation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createFile(t, dir, "file1.txt", "content1")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	svc := &Service{
		backupDirs: []string{dir},
		recursive:  false,
	}

	_, err := svc.collectFilesFromDir(ctx, dir, false)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestCollectAllFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup        func(t *testing.T) *Service
		wantMinFiles int
		wantErr      bool
	}{
		"single directory with files": {
			setup: func(t *testing.T) *Service {
				dir := t.TempDir()
				createFile(t, dir, "file1.txt", "content1")
				createFile(t, dir, "file2.txt", "content2")
				return &Service{
					backupDirs: []string{dir},
					recursive:  false,
				}
			},
			wantMinFiles: 2,
		},
		"multiple directories with files": {
			setup: func(t *testing.T) *Service {
				dir1 := t.TempDir()
				createFile(t, dir1, "file1.txt", "content1")

				dir2 := t.TempDir()
				createFile(t, dir2, "file2.txt", "content2")
				createFile(t, dir2, "file3.txt", "content3")

				return &Service{
					backupDirs: []string{dir1, dir2},
					recursive:  false,
				}
			},
			wantMinFiles: 3,
		},
		"recursive directories": {
			setup: func(t *testing.T) *Service {
				dir := t.TempDir()
				createFile(t, dir, "root.txt", "root")

				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.Mkdir(subdir, 0750))
				createFile(t, subdir, "sub.txt", "sub")

				return &Service{
					backupDirs: []string{dir},
					recursive:  true,
				}
			},
			wantMinFiles: 2,
		},
		"empty directories": {
			setup: func(t *testing.T) *Service {
				return &Service{
					backupDirs: []string{t.TempDir()},
					recursive:  false,
				}
			},
			wantMinFiles: 0,
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := tc.setup(t)
			files, err := svc.collectAllFiles(ctx)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			assert.GreaterOrEqual(t, len(files), tc.wantMinFiles)
		})
	}
}

func TestCollectAllFiles_ContextCancellation(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createFile(t, dir, "file1.txt", "content1")

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately
	cancel()

	svc := &Service{
		backupDirs: []string{dir},
		recursive:  false,
	}

	_, err := svc.collectAllFiles(ctx)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestFileCollector_Walk(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tc := map[string]struct {
		setup     func(t *testing.T) (fc *fileCollector, path string, entry os.DirEntry)
		wantErr   error
		wantFiles int
	}{
		"processes regular file": {
			setup: func(t *testing.T) (*fileCollector, string, os.DirEntry) {
				dir := t.TempDir()
				filePath := filepath.Join(dir, "test.txt")
				createFile(t, dir, "test.txt", "content")

				fc := &fileCollector{
					ctx:       ctx,
					dir:       dir,
					baseDir:   filepath.Base(dir),
					recursive: false,
					files:     make([]string, 0),
				}

				entries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 1)

				return fc, filePath, entries[0]
			},
			wantFiles: 1,
		},
		"skips subdirectory when not recursive": {
			setup: func(t *testing.T) (*fileCollector, string, os.DirEntry) {
				dir := t.TempDir()
				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.Mkdir(subdir, 0750))

				fc := &fileCollector{
					ctx:       ctx,
					dir:       dir,
					baseDir:   filepath.Base(dir),
					recursive: false,
					files:     make([]string, 0),
				}

				entries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 1)

				return fc, subdir, entries[0]
			},
			wantErr:   filepath.SkipDir,
			wantFiles: 0,
		},
		"processes subdirectory when recursive": {
			setup: func(t *testing.T) (*fileCollector, string, os.DirEntry) {
				dir := t.TempDir()
				subdir := filepath.Join(dir, "subdir")
				require.NoError(t, os.Mkdir(subdir, 0750))

				fc := &fileCollector{
					ctx:       ctx,
					dir:       dir,
					baseDir:   filepath.Base(dir),
					recursive: true,
					files:     make([]string, 0),
				}

				entries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 1)

				return fc, subdir, entries[0]
			},
			wantFiles: 0, // Directory itself is not added, only files
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fc, path, entry := tc.setup(t)
			err := fc.walk(path, entry, nil)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, fc.files, tc.wantFiles)
		})
	}

	// Context cancellation test kept separate due to unique setup
	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		filePath := filepath.Join(dir, "test.txt")
		createFile(t, dir, "test.txt", "content")

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		fc := &fileCollector{
			ctx:       ctx,
			dir:       dir,
			baseDir:   filepath.Base(dir),
			recursive: false,
			files:     make([]string, 0),
		}

		entries, err := os.ReadDir(dir)
		require.NoError(t, err)
		require.Len(t, entries, 1)

		err = fc.walk(filePath, entries[0], nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestBuildObjectKey(t *testing.T) {
	t.Parallel()

	tc := map[string]struct {
		fileName string
		ts       time.Time
		want     string
	}{
		"simple filename": {
			fileName: "file.txt",
			ts:       time.Date(2025, 12, 15, 10, 30, 45, 0, time.UTC),
			want:     "2025-12-15T10-30-45/file.txt",
		},
		"filename with path": {
			fileName: "dir/subdir/file.log",
			ts:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			want:     "2025-01-01T00-00-00/dir/subdir/file.log",
		},
		"filename with spaces": {
			fileName: "my file.txt",
			ts:       time.Date(2025, 6, 15, 14, 22, 33, 0, time.UTC),
			want:     "2025-06-15T14-22-33/my file.txt",
		},
	}

	for name, tc := range tc {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := buildObjectKey(tc.fileName, tc.ts)

			assert.Equal(t, tc.want, result)
		})
	}
}

// createFile creates a file with the given content in the specified directory.
func createFile(t *testing.T, dir, name, content string) {
	t.Helper()
	filePath := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0600))
}
