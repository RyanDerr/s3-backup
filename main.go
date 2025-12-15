package main

import (
	"context"
	"log/slog"
	"os"
	"s3-backup/internal/config"
	"s3-backup/internal/s3"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	cfg, err := config.NewConfig(context.Background())
	if err != nil {
		slog.Error("failed to create S3 config", "error", err)
		return
	}

	slog.Info("configuration loaded successfully", "aws_region", cfg.GetAWSRegion(), "s3_bucket", cfg.GetS3Bucket())

	s3Service, err := s3.NewS3Service(context.Background(), cfg)
	if err != nil {
		slog.Error("failed to create S3 service", "error", err)
		return
	}

	if err := s3Service.Backup(context.Background()); err != nil {
		slog.Error("backup failed", "error", err)
		return
	}
	slog.Info("backup completed successfully")
}
