package main

import (
	"context"
	"log/slog"
	"os"
	"s3-backup/internal/config"
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

}
