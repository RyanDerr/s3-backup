package s3

import (
	"context"
	cfg "s3-backup/internal/config/s3"
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
func NewS3Service(ctx context.Context, cfg *cfg.S3Config, opts ...func(*s3.Options)) *S3Service {
	s3Client := s3.NewFromConfig(*cfg.Config, opts...)
	return &S3Service{
		client: s3Client,
		bName:  cfg.GetBucketName(),
	}
}

func (s *S3Service) Backup(ctx context.Context) error {
	var err error
	s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bName,
	})

	return err
}
