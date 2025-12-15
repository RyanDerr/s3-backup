package s3

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

// collectAllFiles aggregates all files from the configured backup directories.
// If recursion is enabled, it traverses subdirectories.
// Returns a combined list of file paths with their S3-ready prefixes.
func (s *S3Service) collectAllFiles(ctx context.Context) ([]string, error) {
	const op = "s3.S3Service.collectAllFiles"

	recursive := s.isRecursive()
	dirs := s.getBackupDirs()

	var allFiles []string
	var joinedErrs error

	for _, dir := range dirs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: %w", op, ctx.Err())
		default:
		}

		files, err := s.collectFilesFromDir(ctx, dir, recursive)
		if err != nil {
			joinedErrs = errors.Join(joinedErrs, err)
			continue
		}
		allFiles = append(allFiles, files...)
	}

	if joinedErrs != nil {
		return allFiles, fmt.Errorf("%s: encountered error(s) when attempting to collect files to backup: %w", op, joinedErrs)
	}

	return allFiles, nil
}

// collectFilesFromDir collects all file paths from a single directory.
// Files are prefixed with the base directory name for S3 organization.
func (s *S3Service) collectFilesFromDir(ctx context.Context, dir string, recursive bool) ([]string, error) {
	const op = "s3.S3Service.collectFilesFromDir"

	if dir == "" {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyDirectory)
	}

	collector := &fileCollector{
		ctx:       ctx,
		dir:       dir,
		baseDir:   filepath.Base(dir),
		recursive: recursive,
		files:     make([]string, 0),
	}

	if err := filepath.WalkDir(dir, collector.walk); err != nil {
		return nil, fmt.Errorf("%s: failed to walk directory %s: %w", op, dir, err)
	}

	return collector.files, nil
}

// fileCollector is a helper type for collecting files during directory traversal.
type fileCollector struct {
	ctx       context.Context
	dir       string
	baseDir   string
	recursive bool
	files     []string
}

// walk is the filepath.WalkDirFunc that processes each entry during directory traversal.
func (fc *fileCollector) walk(path string, d fs.DirEntry, err error) error {
	const op = "s3.fileCollector.walk"
	// Check for context cancellation
	select {
	case <-fc.ctx.Done():
		return fmt.Errorf("%s: %w", op, fc.ctx.Err())
	default:
	}

	if err != nil {
		return fmt.Errorf("%s: error accessing path %s: %w", op, path, err)
	}

	// Skip directories
	if d.IsDir() {
		// If not recursive and this is a subdirectory, skip it
		if !fc.recursive && path != fc.dir {
			return fs.SkipDir
		}
		return nil
	}

	// Calculate relative path from the base directory
	relPath, err := filepath.Rel(fc.dir, path)
	if err != nil {
		return fmt.Errorf("%s: failed to get relative path for %s: %w", op, path, err)
	}

	// Prefix with base directory name and add to collection
	fc.files = append(fc.files, filepath.Join(fc.baseDir, relPath))
	return nil
}

// buildObjectKey constructs the S3 object key with a timestamp prefix.
// Format: YYYY-MM-DDTHH-MM-SS/filename
func buildObjectKey(fn string, ts time.Time) string {
	return fmt.Sprintf("%s/%s", ts.Format("2006-01-02T15-04-05"), fn)
}
