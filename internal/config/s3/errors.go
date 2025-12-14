package s3

import "errors"

var (
	ErrMissingAWSRegion    = errors.New("missing AWS region")
	ErrInvalidAWSRegion    = errors.New("invalid AWS region")
	ErrMissingS3BucketName = errors.New("missing S3 bucket name")
)
