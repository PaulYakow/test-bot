package controller

import (
	"fmt"
	"log"

	"github.com/vitaliy-ukiru/fsm-telebot"
	"github.com/vitaliy-ukiru/fsm-telebot/storages/memory"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"

	"github.com/PaulYakow/test-bot/internal/config"
	"github.com/PaulYakow/test-bot/internal/model"
)

// TODO: После завершения (или отмены) процесса удалять все сообщения кроме команды и последнего сообщения (в которое добавить информацию о созданной записи)

const (
	dateLayout = "02.01.2006"

	helpEnd       = "/help"
	addUserEnd    = "/add_user"
	addAbsenceEnd = "/add_absence"
	testEnd       = "/test"
	cancelEnd     = "/cancel"
)

var (
	// Кнопки общие для всех процессов
	cancelProcessBtn = tele.Btn{Text: "❌ Отменить операцию"}
	skipStateBtn     = tele.Btn{Text: "↪️ Пропустить"}

	// Кнопки (inline) при завершении процессов
	confirmBtn = tele.Btn{Text: "✅ Подтвердить и сохранить", Unique: "confirm"}
	resetBtn   = tele.Btn{Text: "🔄 Сбросить", Unique: "reset"}
	cancelBtn  = tele.Btn{Text: "❌ Отменить", Unique: "cancel"}
)

func New(cfg *config.Config, set Set) (*Controller, error) {
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

	return &Controller{
		bot:     bot,
		manager: manager,
		user:    set.User,
		absence: set.Absence,
	}, nil
}

func (c *Controller) Start() {
	const op = "controller: bot start"

	c.bot.Use(middleware.AutoRespond())
	c.bot.Use(middleware.Recover(func(err error) {
		log.Printf("[ERR-RECOVER] %q", err)
	}))

	c.setCommands()

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

func (c *Controller) setCommands() {
	const op = "controller: set commands"

	// TODO: add const for commands - cmdHelp = "/help" - and use it
	helpCmd := tele.Command{
		Text:        helpEnd,
		Description: "Узнать подробности",
	}
	addUserCmd := tele.Command{
		Text:        addUserEnd,
		Description: "Добавить нового сотрудника",
	}

	addAbsenceCmd := tele.Command{
		Text:        addAbsenceEnd,
		Description: "Добавить причину неявки сотрудника",
	}

	testCmd := tele.Command{
		Text:        testEnd,
		Description: "Тестирование функционала",
	}

	err := c.bot.SetCommands([]tele.Command{
		helpCmd,
		addUserCmd,
		addAbsenceCmd,
		testCmd,
	})
	log.Println(fmt.Sprintf("%s: %v", op, err))

	c.bot.Handle(helpEnd, helpHandler)
	c.bot.Handle(testEnd, testHandler)

	c.manager.Bind(addUserEnd, fsm.DefaultState, startRegisterHandler)
	c.manager.Bind(addAbsenceEnd, fsm.DefaultState, startAbsenceHandler)
	c.manager.Bind(cancelEnd, fsm.AnyState, cancelHandler)
	c.manager.Bind(&cancelProcessBtn, fsm.AnyState, cancelHandler)
}

func helpHandler(tc tele.Context) error {
	msg := `<b>Доступные команды:</b>
/add_user - запуск процесса добавления нового сотрудника
/add_absence - запуск процесса добавления причины отсутствия
/cancel - отмена на любом шаге`
	return tc.Send(msg)
}

func cancelHandler(tc tele.Context, state fsm.Context) error {
	go state.Finish(true)
	return tc.Send("Процесс добавления отменён. Введённые данные удалены.")
}

// TODO: move to view/ui
func replyMarkupWithCancel() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Reply(rm.Row(cancelProcessBtn))
	rm.ResizeKeyboard = true

	return rm
}

func replyMarkupWithCancelAndSkip() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Reply(
		rm.Row(cancelProcessBtn),
		rm.Row(skipStateBtn),
	)
	rm.ResizeKeyboard = true

	return rm
}

func replyMarkupForConfirmState() *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(confirmBtn),
		rm.Row(resetBtn, cancelBtn),
	)

	return rm
}

func replyMarkupList(btn tele.Btn, list []model.RecordInfo) *tele.ReplyMarkup {
	rm := &tele.ReplyMarkup{}
	rows := make([]tele.Row, len(list))
	for i, item := range list {
		btn.Text = item.Description
		btn.Data = item.ID
		rows[i] = rm.Row(btn)
	}
	rm.Inline(rows...)
	rm.ResizeKeyboard = true
	rm.OneTimeKeyboard = true

	return rm
}
