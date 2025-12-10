package config

import "errors"

var (
	ErrMissingAWSRegion = errors.New("missing AWS region")
	ErrInvalidAWSRegion = errors.New("invalid AWS region")
)
