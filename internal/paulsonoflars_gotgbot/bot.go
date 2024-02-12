package paulsonoflars_gotgbot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/PaulYakow/test-bot/internal/config"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

func Start(_ context.Context, cfg *config.Config) {
	b, err := gotgbot.NewBot(cfg.Token, nil)
	if err != nil {
		log.Println("failed to create new bot:", err)
		os.Exit(1)
	}

	// Create updater and dispatcher.
	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		// If an error is returned by a handler, log it and continue going.
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			log.Println("an error occurred while handling update:", err.Error())
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dispatcher, nil)

	// Add echo handler to reply to all text messages.
	dispatcher.AddHandler(handlers.NewMessage(message.Text, echo))

	// Start the webhook server. We start the server before we set the webhook itself, so that when telegram starts
	// sending updates, the server is already ready.
	webhookOpts := ext.WebhookOpts{
		ListenAddr: "0.0.0.0:" + cfg.WebhookPort, // This example assumes you're in a dev environment running ngrok on 8080.
	}

	// The bot's urlPath can be anything. Here, we use "custom-path/<TOKEN>" as an example.
	// It can be a good idea for the urlPath to contain the bot token, as that makes it very difficult for outside
	// parties to find the update endpoint (which would allow them to inject their own updates).
	err = updater.StartWebhook(b, "custom-path/"+cfg.Token, webhookOpts)
	if err != nil {
		log.Println("failed to start webhook:", err)
		os.Exit(1)
	}

	err = updater.SetAllBotWebhooks(cfg.WebhookURL, &gotgbot.SetWebhookOpts{
		MaxConnections:     100,
		DropPendingUpdates: true,
	})
	if err != nil {
		log.Println("failed to set webhook:", err)
		os.Exit(1)
	}

	log.Printf("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

func echo(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(b, ctx.EffectiveMessage.Text, nil)
	if err != nil {
		return fmt.Errorf("failed to echo message: %w", err)
	}
	return nil
}
