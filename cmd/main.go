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

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	b, err := bot.New(os.Getenv("TG_TOKEN"))
	if nil != err {
		log.Println(err)
		os.Exit(1)
	}

	helpCmd := models.BotCommand{
		Command:     "help",
		Description: "Узнать подробности",
	}
	addUserCmd := models.BotCommand{
		Command:     "add_user",
		Description: "Добавить пользователя",
	}
	inlineKbCmd := models.BotCommand{
		Command:     "inline_kb",
		Description: "Пример inline клавиатуры",
	}

	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{
		Commands: []models.BotCommand{
			addUserCmd,
			inlineKbCmd,
		},
	})

	// Register callback
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "button", bot.MatchTypePrefix, inlineKeyboardCallback)

	// Register handlers
	b.RegisterHandler(bot.HandlerTypeMessageText, "/"+helpCmd.Command, bot.MatchTypeExact, helpHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/"+addUserCmd.Command, bot.MatchTypeExact, addUserHandler)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/"+inlineKbCmd.Command, bot.MatchTypeExact, inlineKeyboardHandler)

	b.SetWebhook(ctx, &bot.SetWebhookParams{
		URL: "https://vm-8dae0697.na4u.ru/test-bot",
	})

	go func() {
		err = http.ListenAndServe(":21021", b.WebhookHandler())
		if err != nil {
			log.Println("http error:", err)
		}
	}()

	b.StartWebhook(ctx)
}

func inlineKeyboardCallback(ctx context.Context, b *bot.Bot, update *models.Update) {
	// answering callback query first to let Telegram know that we received the callback query,
	// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
	// as it thinks the callback query doesn't reach to our application. learn more by
	// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.
	ok, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
	if err != nil {
		log.Println("callback answer:", ok, err)
		return
	}
	log.Println("callback:", ok)

	msg := fmt.Sprintf("Ваша кнопка: %s\nUsername: %s",
		update.CallbackQuery.Data,
		update.CallbackQuery.Message.Chat.Username)
	res, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Chat.ID,
		Text:   msg,
	})
	if err != nil {
		log.Println("callback send message:", err)
		return
	}
	log.Println("callback send message:", res)

	res, err = b.EditMessageReplyMarkup(ctx, &bot.EditMessageReplyMarkupParams{
		ChatID:          update.CallbackQuery.Message.Chat.ID,
		InlineMessageID: update.CallbackQuery.InlineMessageID,
		ReplyMarkup:     nil,
	})
	if err != nil {
		log.Println("callback edit message:", err)
		return
	}
	log.Println("callback edit message:", res)
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

	res, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Click by button",
		ReplyMarkup: kb,
	})
	if err != nil {
		log.Println("inline KB send message:", err)
		return
	}
	log.Println("inline KB send message:", res)
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	// TODO: при добавлении команд необходимо изменять данную функцию - связать с созданием команд и сделать через шаблон
	msg := `<b>Команды для взаимодействия:</b>
<i>/start</i> начало работы с ботом
<i>/add_user</i> добавить пользователя
<i>/add_absence</i> добавить новую запись об отсутствии работника
`
	res, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      msg,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		log.Println("help send message:", err)
		return
	}
	log.Println("help send message:", res)
}

type user struct {
	lastname      string
	firstname     string
	middlename    string
	birthday      string
	position      string
	serviceNumber int
}

var newUser = user{}

func addUserHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	msg, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Давай создадим нового сотрудника",
		ReplyMarkup: models.ForceReply{
			ForceReply:            true,
			InputFieldPlaceholder: "Иванов",
		},
	})
	if err != nil {
		log.Println("addUser send message error:", err)
		return
	}

	log.Println("addUser send message:", msg)
}
