package config

import "errors"

var (
	ErrMissingAWSRegion = errors.New("missing AWS region configuration")
)
