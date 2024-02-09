package tucnak_telebot

import (
	"context"
	"log"
	"os"

	tele "gopkg.in/telebot.v3"

	"github.com/PaulYakow/test-bot/internal/config"
)

func Start(_ context.Context, cfg *config.Config) {
	pref := tele.Settings{
		Token: cfg.Token,
		Poller: &tele.Webhook{
			Listen: "0.0.0.0:" + cfg.WebhookPort,
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: cfg.WebhookURL,
			},
		},
		ParseMode: tele.ModeHTML,
	}

	b, err := tele.NewBot(pref)
	if nil != err {
		log.Println(err)
		os.Exit(1)
	}

	b.Handle("/hello", func(c tele.Context) error {
		return c.Send("Hello!")
	})

	b.Start()
}
