package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"s3-backup/internal/config"
	"s3-backup/internal/s3"
)

// Version is set during build via ldflags
var Version = "dev"

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	// Command-line flags
	versionFlag := flag.Bool("version", false, "Print version and exit")
	helpFlag := flag.Bool("help", false, "Print usage information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("s3-backup version %s\n", Version)
		return
	}

	if *helpFlag {
		printUsage()
		return
	}

	ctx := context.Background()

	cfg, err := config.NewConfig(ctx)
	if err != nil {
		slog.Error("failed to create S3 config", "error", err)
		os.Exit(1)
	}

	slog.Info("configuration loaded successfully",
		"version", Version,
		"aws_region", cfg.GetAWSRegion(),
		"s3_bucket", cfg.GetS3Bucket(),
		"cron_schedule", cfg.GetCronSchedule())

	s3Service, err := s3.NewS3Service(ctx, cfg)
	if err != nil {
		slog.Error("failed to create S3 service", "error", err)
		os.Exit(1)
	}

	// Check if cron schedule is configured
	if cfg.GetCronSchedule() != "" {
		slog.Info("starting backup scheduler", "schedule", cfg.GetCronSchedule())
		if err := s3Service.Start(ctx); err != nil {
			slog.Error("scheduler failed", "error", err)
			os.Exit(1)
		}
	} else {
		// One-time backup
		slog.Info("running one-time backup")
		if err := s3Service.Backup(ctx); err != nil {
			slog.Error("backup failed", "error", err)
			os.Exit(1)
		}
		slog.Info("backup completed successfully")
	}
}

func printUsage() {
	fmt.Println("s3-backup - Backup local directories to AWS S3")
	fmt.Printf("Version: %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  s3-backup [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --help            Show this help message")
	fmt.Println("  --version         Show version information")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Set via environment variables or YAML config file")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  S3_BACKUP_CONFIG_FILE      Path to YAML config file")
	fmt.Println("  BACKUP_DIRS                Comma-separated list of directories to backup (required)")
	fmt.Println("  BACKUP_RECURSIVE           Enable recursive backup (true/false, default: false)")
	fmt.Println("  BACKUP_CRON_SCHEDULE       Cron schedule for automatic backups (default: '0 0 */3 * *')")
	fmt.Println("  AWS_REGION                 AWS region (required)")
	fmt.Println("  S3_BUCKET                  S3 bucket name (required)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # One-time backup using environment variables")
	fmt.Println("  export BACKUP_DIRS=/home/user/documents")
	fmt.Println("  export AWS_REGION=us-west-2")
	fmt.Println("  export S3_BUCKET=my-backup-bucket")
	fmt.Println("  s3-backup")
	fmt.Println()
	fmt.Println("  # Scheduled backup using config file")
	fmt.Println("  export S3_BACKUP_CONFIG_FILE=config.yaml")
	fmt.Println("  s3-backup")
	fmt.Println()
	fmt.Println("For more information, see: https://github.com/RyanDerr/s3-backup")
}
