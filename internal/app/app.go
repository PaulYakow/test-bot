package app

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/config"
	bot "github.com/PaulYakow/test-bot/internal/mymmrac_telego"
)

func Run(ctx context.Context, cfg *config.Config) {
	bot.Start(ctx, cfg)
}
