// Package s3 provides S3 backup functionality including file collection and upload.
package s3

import "errors"

var (
	// ErrNilConfig indicates that a nil config was provided.
	ErrNilConfig = errors.New("config cannot be nil")

	// ErrEmptyFilename indicates that an empty filename was provided.
	ErrEmptyFilename = errors.New("filename cannot be empty")

	// ErrEmptyDirectory indicates that an empty directory path was provided.
	ErrEmptyDirectory = errors.New("directory path cannot be empty")

	// ErrDirectoryNotFound indicates that a directory does not exist.
	ErrDirectoryNotFound = errors.New("directory does not exist")

	// ErrNotADirectory indicates that a path is not a directory.
	ErrNotADirectory = errors.New("path is not a directory")
)
