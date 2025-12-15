package s3

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"s3-backup/internal/config"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/robfig/cron/v3"
)

// API defines the interface for S3 operations needed by Service.
type API interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// Service wraps the AWS S3 client and provides backup functionality.
// The client, bucketName, backupDirs, recursive, and cronSchedule fields
// are immutable after NewS3Service returns.
type Service struct {
	client       API
	bucketName   string
	backupDirs   []string
	recursive    bool
	cronSchedule string

	stopCh   chan struct{}
	stopOnce sync.Once
}

// NewS3Service creates a new Service with the provided Config and optional client options.
// It validates that all backup directories exist and are accessible.
func NewS3Service(ctx context.Context, cfg *config.Config, opts ...func(*s3.Options)) (*Service, error) {
	const op = "s3.NewS3Service"

	if cfg == nil {
		return nil, fmt.Errorf("%s: %w", op, ErrNilConfig)
	}

	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get AWS config: %w", op, err)
	}

	s3Client := s3.NewFromConfig(awsCfg, opts...)

	backupDirs := cfg.GetBackupDirs()
	if err := validateDirectories(backupDirs); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Service{
		client:       s3Client,
		bucketName:   cfg.GetS3Bucket(),
		backupDirs:   backupDirs,
		recursive:    cfg.IsRecursive(),
		cronSchedule: cfg.GetCronSchedule(),
		stopCh:       make(chan struct{}),
	}, nil
}

// validateDirectories ensures all provided directories exist and are accessible.
func validateDirectories(dirs []string) error {
	const op = "s3.validateDirectories"
	for _, dir := range dirs {
		if dir == "" {
			return ErrEmptyDirectory
		}

		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("%s: %w: %s", op, ErrDirectoryNotFound, dir)
			}
			return fmt.Errorf("%s: failed to stat directory %s: %w", op, dir, err)
		}

		if !info.IsDir() {
			return fmt.Errorf("%s: %w: %s", op, ErrNotADirectory, dir)
		}
	}
	return nil
}

// getBackupDirs returns a copy of the configured backup directories.
// This method is safe to call concurrently.
func (s *Service) getBackupDirs() []string {
	dirs := make([]string, len(s.backupDirs))
	copy(dirs, s.backupDirs)
	return dirs
}

// isRecursive returns whether recursive backup is enabled.
// This method is safe to call concurrently.
func (s *Service) isRecursive() bool {
	return s.recursive
}

// Backup performs the backup of files from the configured directories to the S3 bucket.
// It respects context cancellation and returns all errors encountered during the backup.
func (s *Service) Backup(ctx context.Context) error {
	const op = "s3.Service.Backup"

	files, err := s.collectAllFiles(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to collect files: %w", op, err)
	}

	if err := s.backupAllFiles(ctx, files); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// backupAllFiles uploads all provided files to the S3 bucket.
// It continues processing all files even if some fail, collecting all errors.
func (s *Service) backupAllFiles(ctx context.Context, files []string) error {
	const op = "s3.Service.backupAllFiles"

	if len(files) == 0 {
		return nil
	}

	var joinedErrs error
	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s: %w", op, ctx.Err())
		default:
		}

		if err := s.backupFile(ctx, file); err != nil {
			joinedErrs = errors.Join(joinedErrs, err)
		}
	}

	if joinedErrs != nil {
		return fmt.Errorf("%s: one or more files failed to backup: %w", op, joinedErrs)
	}
	return nil
}

// backupFile uploads a single file to the configured S3 bucket.
// The S3 object key is constructed with a timestamp prefix and the file's relative path.
func (s *Service) backupFile(ctx context.Context, fileName string) error {
	const op = "s3.Service.backupFile"

	if fileName == "" {
		return fmt.Errorf("%s: %w", op, ErrEmptyFilename)
	}

	//nolint:gosec // G304: fileName comes from user's configured backup directories
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("%s: failed to open file %s: %w", op, fileName, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			slog.Warn("failed to close file", "file", fileName, "error", closeErr)
		}
	}()

	s3Key, err := s.buildS3Key(fileName)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	key := buildObjectKey(s3Key, time.Now())

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucketName,
		Key:    &key,
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("%s: failed to put object to S3 (key=%s): %w", op, key, err)
	}

	return nil
}

// buildS3Key constructs an S3 key from the full file path by finding the backup directory
// it belongs to and creating a relative path with the base directory name as prefix.
// For example: /data/documents/invoices/invoice-001.txt -> documents/invoices/invoice-001.txt
func (s *Service) buildS3Key(filePath string) (string, error) {
	const op = "s3.Service.buildS3Key"

	// Find which backup directory this file belongs to
	for _, dir := range s.backupDirs {
		// Check if the file path starts with this backup directory
		relPath, err := filepath.Rel(dir, filePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			// File is not under this directory, try next one
			continue
		}

		// Found the matching directory - construct S3 key with base directory name
		baseDir := filepath.Base(dir)
		return filepath.Join(baseDir, relPath), nil
	}

	return "", fmt.Errorf("%s: file %s does not belong to any configured backup directory", op, filePath)
}

// Start begins the scheduled backup process in the background.
// It runs backups according to the configured cron schedule.
// The scheduler will stop when the context is cancelled or Stop() is called.
func (s *Service) Start(ctx context.Context) error {
	const op = "s3.Service.Start"

	schedule := s.cronSchedule

	c := cron.New()
	_, err := c.AddFunc(schedule, func() {
		// Create a new context for each backup job that respects the parent context
		backupCtx := ctx
		if ctx.Err() != nil {
			slog.Warn("skipping scheduled backup: context cancelled")
			return
		}
		slog.Info("starting scheduled backup", "time", time.Now().Format(time.RFC3339))
		if err := s.Backup(backupCtx); err != nil {
			slog.Error("scheduled backup failed", "error", err)
		} else {
			slog.Info("scheduled backup completed successfully", "time", time.Now().Format(time.RFC3339))
		}
	})

	if err != nil {
		return fmt.Errorf("%s: invalid cron schedule %q: %w", op, schedule, err)
	}

	c.Start()

	slog.Info("backup scheduler started", "schedule", schedule)

	// Block until stop signal or context cancellation
	select {
	case <-s.stopCh:
		slog.Info("received stop signal")
	case <-ctx.Done():
		slog.Info("context cancelled, stopping scheduler")
	}

	// Graceful shutdown
	shutdownCtx := c.Stop()
	<-shutdownCtx.Done()

	slog.Info("backup scheduler stopped")
	return nil
}

// Stop gracefully stops the scheduled backup process.
// It is safe to call multiple times.
func (s *Service) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
	})
}
