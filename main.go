// Package main provides the s3-backup CLI tool for backing up files to AWS S3.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"s3-backup/internal/config"
	"s3-backup/internal/s3"
	"syscall"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	os.Exit(run())
}

func run() int {
	// Create context that cancels on interrupt signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		slog.Info("received shutdown signal", "signal", sig)
		cancel()
	}()

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("failed to create S3 config", "error", err)
		return 1
	}

	slog.Info("configuration loaded successfully",
		"aws_region", cfg.GetAWSRegion(),
		"s3_bucket", cfg.GetS3Bucket(),
		"cron_schedule", cfg.GetCronSchedule())

	s3Service, err := s3.NewS3Service(ctx, cfg)
	if err != nil {
		slog.Error("failed to create S3 service", "error", err)
		return 1
	}

	// Check if cron schedule is configured
	if cfg.GetCronSchedule() != "" {
		slog.Info("starting backup scheduler", "schedule", cfg.GetCronSchedule())
		if err := s3Service.Start(ctx); err != nil {
			slog.Error("scheduler failed", "error", err)
			return 1
		}
		slog.Info("scheduler stopped gracefully")
		return 0
	}
	// One-time backup
	slog.Info("running one-time backup")
	if err := s3Service.Backup(ctx); err != nil {
		slog.Error("backup failed", "error", err)
		return 1
	}
	slog.Info("backup completed successfully")
	return 0
}
