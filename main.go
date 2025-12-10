package main

import (
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	cfg, err := config.LoadDefaultConfig()
	if err != nil {
		slog.Error("unable to load AWS SDK from config", err)
		return
	}

}
