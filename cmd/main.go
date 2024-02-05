package main

import (
	"context"
	"fmt"
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
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv("TG_TOKEN"), opts...)
	if nil != err {
		log.Println(err)
		os.Exit(1)
	}

	// FIXME: нет списка команд с описанием при вводе '/'
	addUserCmd := models.BotCommand{
		Command:     "add_user",
		Description: "Добавить пользователя",
	}
	inlineCmd := models.BotCommand{
		Command:     "inline",
		Description: "Пример inline-режима",
	}
	inlineKbCmd := models.BotCommand{
		Command:     "inline_kb",
		Description: "Пример inline клавиатуры",
	}

	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			addUserCmd,
			inlineCmd,
			inlineKbCmd,
		},
	})

	// Register callback
	b.RegisterHandler(bot.HandlerTypeMessageText, "button", bot.MatchTypePrefix, callbackHandler)

	// Register handlers
	b.RegisterHandler(bot.HandlerTypeMessageText, addUserCmd.Command, bot.MatchTypePrefix, addUserHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, inlineCmd.Command, bot.MatchTypePrefix, inlineHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, inlineKbCmd.Command, bot.MatchTypePrefix, inlineKeyboardHandler)

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

	msg := fmt.Sprintf("You selected the button: %s\nUsername: %s",
		update.CallbackQuery.Data,
		update.CallbackQuery.Message.Chat.Username)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Chat.ID,
		Text:   msg,
	})
}

func inlineKeyboardHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Button 1", CallbackData: "button_1"},
				{Text: "Button 2", CallbackData: "button_2"},
			}, {
				{Text: "Button 3", CallbackData: "button_3"},
			},
		},
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Click by button",
		ReplyMarkup: kb,
	})
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// FIXME: нет обработки Markdown
	msg := `*Команды для взаимодействия:*
_/start_ - начало работы с ботом (происходит запись пользователя в БД),
_/add-user_ - добавить пользователя (Фамилия Имя Должность Ник_Телеграм)
_/add-absence_ - добавить новую запись об отсутствии работника (Работник (id?) Код_отсутствия Дата_начала).
`
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msg,
		ParseMode: "MarkdownV2",
	})
}

func addUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg := fmt.Sprintf("Пользователь %s", update.Message.Chat.Username)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})

}

func inlineHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.InlineQuery == nil {
		return
	}

	results := []models.InlineQueryResult{
		&models.InlineQueryResultArticle{ID: "1", Title: "Foo 1", InputMessageContent: &models.InputTextMessageContent{MessageText: "foo 1"}},
		&models.InlineQueryResultArticle{ID: "2", Title: "Foo 2", InputMessageContent: &models.InputTextMessageContent{MessageText: "foo 2"}},
		&models.InlineQueryResultArticle{ID: "3", Title: "Foo 3", InputMessageContent: &models.InputTextMessageContent{MessageText: "foo 3"}},
	}

	b.AnswerInlineQuery(ctx, &bot.AnswerInlineQueryParams{
		InlineQueryID: update.InlineQuery.ID,
		Results:       results,
	})
}
