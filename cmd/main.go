package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/PaulYakow/test-bot/internal/app"
	"github.com/PaulYakow/test-bot/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to get config: %s", err.Error()))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app.Run(ctx, cfg)
}
