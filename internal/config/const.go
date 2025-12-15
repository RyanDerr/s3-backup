package config

const (
	// EnvConfigFile is the path to the YAML configuration file
	EnvConfigFile = "S3_BACKUP_CONFIG_FILE"

	// EnvBackupDirs is the environment variable for backup directories.
	EnvBackupDirs = "BACKUP_DIRS"
	// EnvRecursive is the environment variable for recursive backup mode.
	EnvRecursive = "BACKUP_RECURSIVE"
	// EnvCronSchedule is the environment variable for cron schedule.
	EnvCronSchedule = "BACKUP_CRON_SCHEDULE"

	// EnvAWSRegion is the environment variable for AWS region.
	EnvAWSRegion = "AWS_REGION"
	// EnvS3Bucket is the environment variable for S3 bucket name.
	EnvS3Bucket = "S3_BUCKET"
)
