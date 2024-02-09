package niconex_echotron

import (
	"context"

	"github.com/NicoNex/echotron/v3"

	"github.com/PaulYakow/test-bot/internal/config"
)

type bot struct {
	chatID int64
	echotron.API
}

func (b *bot) Update(update *echotron.Update) {
	if update.Message.Text == "/start" {
		b.SendMessage("Hello world", b.chatID, nil)
	}
}

func Start(_ context.Context, cfg *config.Config) {
	newBot := func(chatID int64) echotron.Bot {
		return &bot{
			chatID,
			echotron.NewAPI(cfg.Token),
		}
	}

	dsp := echotron.NewDispatcher(cfg.Token, newBot)
	dsp.ListenWebhook(cfg.WebhookURL)
}
