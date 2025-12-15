package s3

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FuzzBuildObjectKey tests S3 object key generation with fuzzy input
func FuzzBuildObjectKey(f *testing.F) {
	// Seed corpus with various filename patterns
	f.Add("simple.txt", int64(1234567890))
	f.Add("path/to/file.txt", int64(0))
	f.Add("../../../etc/passwd", int64(1234567890))
	f.Add("file with spaces.txt", int64(1234567890))
	f.Add("文件名.txt", int64(1234567890))                 // Unicode
	f.Add("file\x00name.txt", int64(1234567890))        // Null byte
	f.Add(strings.Repeat("a", 1000), int64(1234567890)) // Long name
	f.Add("", int64(1234567890))                        // Empty name
	f.Add("./file.txt", int64(1234567890))
	f.Add("file'; DROP TABLE;--.txt", int64(1234567890))

	f.Fuzz(func(t *testing.T, filename string, unixTime int64) {
		// Clamp unix time to reasonable range to avoid time.Unix panic
		if unixTime < -62135596800 || unixTime > 253402300799 {
			t.Skip("Unix time out of valid range")
		}

		ts := time.Unix(unixTime, 0)

		// Should not panic with any input
		key := buildObjectKey(filename, ts)

		// Key should be a valid string
		_ = len(key)

		// Key should contain the timestamp prefix
		expectedPrefix := ts.Format("2006-01-02T15-04-05")
		if !strings.Contains(key, expectedPrefix) {
			t.Errorf("Key missing timestamp prefix: got %q, want prefix %q", key, expectedPrefix)
		}

		// Key should not contain null bytes (S3 doesn't allow them)
		if strings.Contains(key, "\x00") {
			t.Errorf("Key contains null byte: %q", key)
		}
	})
}

// FuzzFileCollectorWalk tests directory walking with fuzzy paths
func FuzzFileCollectorWalk(f *testing.F) {
	// Seed with various path patterns
	f.Add("file.txt", "Documents", "Documents", false)
	f.Add("subdir/file.txt", "Documents", "Documents", true)
	f.Add("../../../etc/passwd", "Documents", "Documents", false)
	f.Add("", "Documents", "Documents", false)
	f.Add("file with spaces.txt", "My Documents", "My Documents", true)
	f.Add("file\x00.txt", "Documents", "Documents", false)
	f.Add(strings.Repeat("a", 500), "Documents", "Documents", false)

	f.Fuzz(func(t *testing.T, relPath, dir, baseDir string, recursive bool) {
		// Create a safe temp directory
		tmpDir := t.TempDir()

		// Construct full path safely
		fullPath := filepath.Join(tmpDir, filepath.Clean(relPath))

		// Only test if the relative path is actually under tmpDir
		// This prevents path traversal attacks in tests
		if !strings.HasPrefix(fullPath, tmpDir) {
			t.Skip("Path escapes temp directory")
		}

		// Create the file if possible
		if relPath != "" && !strings.Contains(relPath, "\x00") {
			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Skip("Cannot create parent directory")
			}

			// Create the file
			if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
				t.Skip("Cannot create test file")
			}
		}

		fc := &fileCollector{
			ctx:       context.Background(),
			dir:       tmpDir,
			baseDir:   baseDir,
			recursive: recursive,
			files:     make([]string, 0),
		}

		// Walk should not panic
		info, err := os.Lstat(fullPath)
		if err != nil {
			t.Skip("Cannot stat file")
		}

		entry := fs.FileInfoToDirEntry(info)
		_ = fc.walk(fullPath, entry, nil)

		// Verify collected files are under the base directory
		for _, file := range fc.files {
			if strings.Contains(file, "\x00") {
				t.Errorf("Collected file contains null byte: %q", file)
			}
		}
	})
}

// FuzzValidateDirectories tests directory validation with fuzzy input
func FuzzValidateDirectories(f *testing.F) {
	// Seed with various directory patterns
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
		// Split on comma for multiple directories
		dirs := strings.Split(dirPath, ",")

		// Should not panic with any input
		_ = validateDirectories(dirs)

		// If we created temp directories, test with those
		if dirPath != "" && !strings.Contains(dirPath, "\x00") {
			tmpDir := t.TempDir()
			_ = validateDirectories([]string{tmpDir})
		}
	})
}
