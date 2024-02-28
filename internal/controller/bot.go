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

const (
	dateLayout = "02.01.2006"
)

var (
	// Кнопки общие для всех процессов
	cancelProcessBtn = tele.Btn{Text: "❌ Отменить операцию"}

	// Кнопки (inline) при завершении процессов
	confirmBtn = tele.Btn{Text: "✅ Подтвердить и сохранить", Unique: "confirm"}
	resetBtn   = tele.Btn{Text: "🔄 Сбросить", Unique: "reset"}
	cancelBtn  = tele.Btn{Text: "❌ Отменить", Unique: "cancel"}
)

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

	helpCmd := tele.Command{
		Text:        "help",
		Description: "Узнать подробности",
	}
	regCmd := tele.Command{
		Text:        "reg",
		Description: "Добавить нового сотрудника",
	}

	addAbsenceCmd := tele.Command{
		Text:        "add_absence",
		Description: "Добавить причину неявки сотрудника",
	}

	testCmd := tele.Command{
		Text:        "test",
		Description: "Для тестов функционала",
	}

	err := c.bot.SetCommands([]tele.Command{
		helpCmd,
		regCmd,
		addAbsenceCmd,
		testCmd,
	})
	log.Println(fmt.Sprintf("%s: %v", op, err))

	// Commands
	c.bot.Handle("/"+helpCmd.Text, helpHandler)
	c.bot.Handle("/"+testCmd.Text, testHandler)
	c.manager.Bind("/"+regCmd.Text, fsm.DefaultState, startRegisterHandler)
	c.manager.Bind("/"+addAbsenceCmd.Text, fsm.DefaultState, startAbsenceHandler)
	c.manager.Bind("/cancel", fsm.AnyState, cancelHandler)
	c.manager.Bind(&cancelProcessBtn, fsm.AnyState, cancelHandler)

	c.manager.Bind("/state", fsm.AnyState, func(c tele.Context, state fsm.Context) error {
		s, err := state.State()
		if err != nil {
			return c.Send(fmt.Sprintf("can't get state: %s", err))
		}
		return c.Send(s.GoString())
	})

	c.registerProcessInit()
	c.absenceProcessInit()

	log.Println("Handlers configured")

	c.bot.Start()
}

func helpHandler(tc tele.Context) error {
	msg := `<b>Доступные команды:</b>
/reg - запуск процесса добавления нового сотрудника
/add_absence - запуск процесса добавления причины отсутствия
/cancel - отмена на любом шаге`
	return tc.Send(msg)
}

func cancelHandler(tc tele.Context, state fsm.Context) error {
	go state.Finish(true)
	return tc.Send("Процесс добавления отменён. Введённые данные удалены.")
}

func editFormMessage(old, new string) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			strOffset := utf8.RuneCountInString(old)
			if nLen := utf8.RuneCountInString(new); nLen > 1 {
				strOffset -= nLen - 1
			}
			fmt.Printf("edit message: strOffset=%d\n", strOffset)

			entities := make(tele.Entities, len(c.Message().Entities))
			for i, entity := range c.Message().Entities {
				entity.Offset -= strOffset
				entities[i] = entity
			}
			fmt.Printf("edit message: entities=%v\n", entities)

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
