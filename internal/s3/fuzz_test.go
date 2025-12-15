package s3

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FuzzBuildObjectKey tests S3 object key generation with fuzzy input.
func FuzzBuildObjectKey(f *testing.F) {
	f.Add("simple.txt", int64(1234567890))
	f.Add("path/to/file.txt", int64(0))
	f.Add("../../../etc/passwd", int64(1234567890))
	f.Add("file with spaces.txt", int64(1234567890))
	f.Add("文件名.txt", int64(1234567890))
	f.Add("file\x00name.txt", int64(1234567890))
	f.Add(strings.Repeat("a", 1000), int64(1234567890))
	f.Add("", int64(1234567890))
	f.Add("./file.txt", int64(1234567890))
	f.Add("file'; DROP TABLE;--.txt", int64(1234567890))

	f.Fuzz(func(t *testing.T, filename string, unixTime int64) {
		if unixTime < -62135596800 || unixTime > 253402300799 {
			t.Skip("unix time outside valid range for time.Unix")
		}

		if strings.Contains(filename, "\x00") {
			t.Skip("filename contains null byte (invalid in filesystems)")
		}

		ts := time.Unix(unixTime, 0)
		key := buildObjectKey(filename, ts)

		expectedPrefix := ts.Format("2006-01-02T15-04-05")
		if !strings.Contains(key, expectedPrefix) {
			t.Errorf("key missing timestamp prefix: got %q, want prefix %q", key, expectedPrefix)
		}
	})
}

// FuzzCollectFilesFromDir tests directory file collection with fuzzy paths.
// This tests the public API rather than internal walk implementation.
func FuzzCollectFilesFromDir(f *testing.F) {
	f.Add("file.txt", false)
	f.Add("subdir/file.txt", true)
	f.Add("file with spaces.txt", true)
	f.Add(strings.Repeat("a", 500)+".txt", false)

	f.Fuzz(func(t *testing.T, relPath string, recursive bool) {
		if relPath == "" || strings.Contains(relPath, "\x00") {
			t.Skip("invalid path for filesystem")
		}

		tmpDir := t.TempDir()
		fullPath := filepath.Join(tmpDir, filepath.Clean(relPath))

		if !strings.HasPrefix(fullPath, tmpDir) {
			t.Skip("path escapes temp directory")
		}

		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			t.Skip("cannot create parent directory")
		}

		if err := os.WriteFile(fullPath, []byte("test"), 0600); err != nil {
			t.Skip("cannot create test file")
		}

		svc := &Service{
			recursive: recursive,
		}

		files, err := svc.collectFilesFromDir(context.Background(), tmpDir, recursive)
		if err != nil {
			t.Logf("collectFilesFromDir returned error: %v", err)
		}

		for _, file := range files {
			if strings.Contains(file, "\x00") {
				t.Errorf("collected file contains null byte: %q", file)
			}
		}
	})
}

// FuzzValidateDirectories tests directory validation with fuzzy input.
func FuzzValidateDirectories(f *testing.F) {
	f.Add("/tmp/test")
	f.Add("/tmp/a,/tmp/b")
	f.Add("")
	f.Add("../../../etc")
	f.Add("/nonexistent")
	f.Add("relative/path")
	f.Add("/tmp/file\x00name")
	f.Add(strings.Repeat("/tmp/", 100))
	f.Add("/tmp/'; DROP TABLE;--")

	f.Fuzz(func(t *testing.T, dirPath string) {
		dirs := strings.Split(dirPath, ",")
		_ = validateDirectories(dirs)

		if dirPath != "" && !strings.Contains(dirPath, "\x00") {
			tmpDir := t.TempDir()
			_ = validateDirectories([]string{tmpDir})
		}
	})
}
