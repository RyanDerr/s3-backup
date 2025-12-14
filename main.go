package main

import (
	"context"
	"log/slog"
	"os"
	s3Cfg "s3-backup/internal/config/s3"
)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
}

func main() {
	s3Config, err := s3Cfg.NewS3Config(context.Background())
	if err != nil {
		slog.Error("failed to create S3 config", "error", err)
		return
	}

	slog.Info("S3 Config created", "region", s3Config.GetRegion())

}
