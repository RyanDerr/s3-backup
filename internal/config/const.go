package config

const (
	// EnvConfigFile is the path to the YAML configuration file
	EnvConfigFile = "S3_BACKUP_CONFIG_FILE"

	// Backup configuration
	EnvBackupDirs = "BACKUP_DIRS"

	// AWS S3 configuration
	EnvAWSRegion = "AWS_REGION"
	EnvS3Bucket  = "S3_BUCKET"
)
