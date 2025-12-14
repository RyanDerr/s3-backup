package s3

import (
	"context"
	"fmt"
	"s3-backup/internal/config"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Service wraps the AWS S3 client.
type S3Service struct {
	client *s3.Client
	bName  string

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
		client: s3Client,
		bName:  cfg.GetS3Bucket(),
	}, nil
}

func (s *S3Service) Backup(ctx context.Context) error {
	var err error
	s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bName,
	})

	return err
}
