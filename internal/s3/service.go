package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"s3-backup/internal/config"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Service wraps the AWS S3 client.
type S3Service struct {
	client     *s3.Client
	bucketName string
	backupDirs []string

	sync.RWMutex
}

// NewS3Service creates a new S3Service with the provided S3Config and optional client options.
func NewS3Service(ctx context.Context, cfg *config.Config, opts ...func(*s3.Options)) (*S3Service, error) {
	const op = "s3.NewS3Service"

	awsCfg, err := cfg.GetAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get AWS config: %w", op, err)
	}

	s3Client := s3.NewFromConfig(awsCfg, opts...)

	return &S3Service{
		client:     s3Client,
		bucketName: cfg.GetS3Bucket(),
		backupDirs: cfg.GetBackupDirs(),
	}, nil
}

// Backup performs the backup of files from the configured directories to the S3 bucket.
func (s *S3Service) Backup(ctx context.Context) error {
	const op = "s3.S3Service.Backup"

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	files, err := s.collectAllFiles()
	if err != nil {
		return err
	}

	ts := time.Now()

	return s.backupAllFiles(ctx, files, ts)
}

// collectAllFiles aggregates all files from the configured backup directories.
// Returns a combined list of files and any errors encountered during collection.
func (s *S3Service) collectAllFiles() ([]string, error) {
	const op = "s3.S3Service.collectAllFiles"
	var aggFiles []string
	var joinedErrs error

	for _, dir := range s.backupDirs {
		files, err := getFilesInDirectory(dir)
		if err != nil {
			joinedErrs = errors.Join(joinedErrs, fmt.Errorf("%s: failed to get files from %s: %w", op, dir, err))
			continue
		}
		aggFiles = append(aggFiles, files...)
	}

	return aggFiles, joinedErrs
}

// backupAllFiles uploads all provided files to the S3 bucket.
// Returns any errors encountered during the backup process.
func (s *S3Service) backupAllFiles(ctx context.Context, files []string, ts time.Time) error {
	const op = "s3.S3Service.backupAllFiles"
	var joinedErrs error

	for _, file := range files {
		if err := s.backupFile(ctx, file, ts); err != nil {
			joinedErrs = errors.Join(joinedErrs, fmt.Errorf("%s: failed to backup file %s: %w", op, file, err))
		}
	}

	return joinedErrs
}

// backupFile uploads a single file to the configured S3 bucket.
func (s *S3Service) backupFile(ctx context.Context, fileName string, ts time.Time) error {
	const op = "s3.S3Service.backupFile"

	content, err := readFileContent(fileName)
	if err != nil {
		return fmt.Errorf("%s: failed to read file content: %w", op, err)
	}

	prefix := fmt.Sprintf("%s/%s", ts.Format("2006-01-02T15-04-05"), fileName)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucketName,
		Key:    &prefix,
		Body:   bytes.NewBuffer(content),
	})

	if err != nil {
		return fmt.Errorf("%s: failed to put object to S3: %w", op, err)
	}

	return nil
}
