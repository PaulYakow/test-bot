package mymmrac_telego

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fasthttp/router"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/valyala/fasthttp"

	"github.com/PaulYakow/test-bot/internal/config"
)

func Webhook(_ context.Context, bot *telego.Bot, secret string) telego.WebhookServer {
	return telego.FastHTTPWebhookServer{
		Logger:      bot.Logger(),
		Server:      &fasthttp.Server{},
		Router:      router.New(),
		SecretToken: secret,
	}
}

func Start(_ context.Context, cfg *config.Config) {
	// Note: Please keep in mind that default logger may expose sensitive information, use in development only
	bot, err := telego.NewBot(cfg.Token, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("Create bot: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Note: Creating a secret token like this is not secure,
	// but at least better than having a plain bot token as is in requests
	secretBytes := sha256.Sum256([]byte(cfg.Token))
	secret := hex.EncodeToString(secretBytes[:])

	srv := Webhook(ctx, bot, secret)

	updates, err := bot.UpdatesViaWebhook(
		"/test-bot",
		telego.WithWebhookServer(srv),
		telego.WithWebhookSet(tu.Webhook(cfg.WebhookURL).WithSecretToken(secret)),
		telego.WithWebhookContext(ctx),
	)
	if err != nil {
		log.Fatalf("Updates via webhoo: %s", err)
	}

	bh, err := th.NewBotHandler(bot, updates, th.WithDone(ctx.Done()), th.WithStopTimeout(time.Second*10))
	if err != nil {
		log.Fatalf("Bot handler: %s", err)
	}

	RegisterHandlers(bh)

	done := make(chan struct{}, 1)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Println("Stopping...")

		err = bot.StopWebhook()
		if err != nil {
			log.Println("Failed to stop webhook properly:", err)
		}
		bh.Stop()

		done <- struct{}{}
	}()

	go bh.Start()
	log.Println("Handling updates...")

	go func() {
		err = bot.StartWebhook(":" + cfg.WebhookPort)
		if err != nil {
			log.Fatalf("Failed to start webhook: %s", err)
		}
	}()

	<-done
	log.Println("Done")
}

func RegisterHandlers(bh *th.BotHandler) {
	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID), "Menu").
			WithReplyMarkup(tu.Keyboard(
				tu.KeyboardRow(
					tu.KeyboardButton("Sub menu 1"),
					tu.KeyboardButton("Sub menu 2"),
				),
				tu.KeyboardRow(
					tu.KeyboardButton("Sub menu 3"),
				),
			).WithResizeKeyboard()))
		if err != nil {
			log.Printf("Error on start: %s", err)
		}
	}, th.Union(th.CommandEqual("start"), th.TextEqual("Back")))

	subMenu := bh.Group(th.TextPrefix("Sub menu"))
	subMenu.Use(func(bot *telego.Bot, update telego.Update, next th.Handler) {
		log.Println("Sub menu group")
		next(bot, update)
	})

	subMenu.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID), "Sub menu 1 content").
			WithReplyMarkup(tu.Keyboard(tu.KeyboardRow(tu.KeyboardButton("Back"))).WithResizeKeyboard()))
		if err != nil {
			log.Printf("Error on sub menu 1: %s", err)
		}
	}, th.TextSuffix("1"))

	subMenu.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID), "Sub menu 2 content").
			WithReplyMarkup(tu.Keyboard(tu.KeyboardRow(tu.KeyboardButton("Back"))).WithResizeKeyboard()))
		if err != nil {
			log.Printf("Error on sub menu 2: %s", err)
		}
	}, th.TextSuffix("2"))

	subMenu.HandleMessage(func(bot *telego.Bot, message telego.Message) {
		_, err := bot.SendMessage(tu.Message(tu.ID(message.Chat.ID), "Sub menu 3 content").
			WithReplyMarkup(tu.Keyboard(tu.KeyboardRow(tu.KeyboardButton("Back"))).WithResizeKeyboard()))
		if err != nil {
			log.Printf("Error on sub menu 3: %s", err)
		}
	}, th.TextSuffix("3"))
}
