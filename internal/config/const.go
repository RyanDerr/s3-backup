package config

const (
	// EnvConfigFile is the path to the YAML configuration file
	EnvConfigFile = "S3_BACKUP_CONFIG_FILE"

	// Backup configuration
	EnvBackupDirs   = "BACKUP_DIRS"
	EnvRecursive    = "BACKUP_RECURSIVE"
	EnvCronSchedule = "BACKUP_CRON_SCHEDULE"

	// AWS S3 configuration
	EnvAWSRegion = "AWS_REGION"
	EnvS3Bucket  = "S3_BUCKET"

	// DefaultCronSchedule is the default backup schedule (every 3 days)
	DefaultCronSchedule = "0 0 */3 * *"
)
