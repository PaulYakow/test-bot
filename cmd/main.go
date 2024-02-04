package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Send any text message to the bot after the bot has been started

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, callbackHandler),
	}

	b, err := bot.New(os.Getenv("TG_TOKEN"), opts...)
	if nil != err {
		log.Println(err)
		os.Exit(1)
	}

	b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: "https://vm-8dae0697.na4u.ru/test-bot",
	})

	go func() {
		err = http.ListenAndServe(":21021", b.WebhookHandler())
		if err != nil {
			log.Println(err)
		}
	}()

	b.StartWebhook(ctx)
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// answering callback query first to let Telegram know that we received the callback query,
	// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
	// as it thinks the callback query doesn't reach to our application. learn more by
	// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.
	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
	msg := "You selected the button: " + update.CallbackQuery.Data + update.Message.Chat.Username
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Chat.ID,
		Text:   msg,
	})
}

func handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Button 1", CallbackData: "button #1"},
				{Text: "Button 2", CallbackData: "button #2"},
			}, {
				{Text: "Button 3", CallbackData: "button #3"},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Click by button",
		ReplyMarkup: kb,
	})
}
