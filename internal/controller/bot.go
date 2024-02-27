package controller

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/vitaliy-ukiru/fsm-telebot"
	"github.com/vitaliy-ukiru/fsm-telebot/storages/memory"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/PaulYakow/test-bot/internal/config"
)

const SuperuserId tele.ChatID = 384688499 // TODO: move to config

type controller struct {
	bot     *tele.Bot
	manager *fsm.Manager
	us      UserService
}

func New(cfg *config.Config, us UserService) (*controller, error) {
	const op = "controller: create new"

	bot, err := tele.NewBot(tele.Settings{
		Token: cfg.Token,
		Poller: &tele.Webhook{
			Listen: "0.0.0.0:" + cfg.WebhookPort,
			Endpoint: &tele.WebhookEndpoint{
				PublicURL: cfg.WebhookURL,
			},
		},
		ParseMode: tele.ModeHTML,
		OnError: func(err error, c tele.Context) {
			log.Printf("[ERR] %q chat=%s", err, c.Recipient())
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := memory.NewStorage()
	defer storage.Close()

	manager := fsm.NewManager(bot, nil, storage, nil)

	return &controller{
		bot:     bot,
		manager: manager,
		us:      us,
	}, nil
}

func (c *controller) Start() {
	const op = "controller: bot start"

	c.bot.Use(middleware.AutoRespond())

	// TODO: set list of the bot's commands	(bot.SetCommands)

	// Commands
	c.bot.Handle("/start", startHandler)
	c.manager.Bind("/reg", fsm.DefaultState, startRegisterHandler)
	c.manager.Bind("/cancel", fsm.AnyState, cancelRegisterHandler)

	c.manager.Bind("/state", fsm.AnyState, func(c tele.Context, state fsm.Context) error {
		s, err := state.State()
		if err != nil {
			return c.Send(fmt.Sprintf("can't get state: %s", err))
		}
		return c.Send(s.GoString())
	})

	c.registerProcessInit()

	log.Println("Handlers configured")

	c.bot.Start()
}

func startHandler(tc tele.Context) error {
	msg := `<b>Доступные команды:</b>
/reg - запуск процесса добавления нового сотрудника
/cancel - отмена на любом шаге`
	return tc.Send(msg)
}

func editFormMessage(old, new string) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			strOffset := utf8.RuneCountInString(old)
			if nLen := utf8.RuneCountInString(new); nLen > 1 {
				strOffset -= nLen - 1
			}

			entities := make(tele.Entities, len(c.Message().Entities))
			for i, entity := range c.Message().Entities {
				entity.Offset -= strOffset
				entities[i] = entity
			}
			defer func() {
				err := c.EditOrSend(strings.Replace(c.Message().Text, old, new, 1), entities)
				if err != nil {
					c.Bot().OnError(err, c)
				}
			}()
			return next(c)
		}
	}
}

func deleteAfterHandler(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		defer func(c tele.Context) {
			if err := c.Delete(); err != nil {
				c.Bot().OnError(err, c)
			}
		}(c)
		return next(c)
	}
}
