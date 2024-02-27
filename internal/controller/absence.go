package controller

import (
	"github.com/vitaliy-ukiru/fsm-telebot"
	tele "gopkg.in/telebot.v3"
)

/*
user_id bigint NOT NULL - использовать фамилию (+ табельный, если нашлось несколько)
"type" absence_code NOT NULL - выгружать доступные коды из БД и формировать кнопки для выбора
date_begin date NOT NULL -
date_end date - может заполняться позже (предусмотреть пропуск, например, по кнопке)

Вначале можно выдавать список записей (по запросу), в которых date_end IS NULL - для выбора заполнения date_end (сразу переходить к нему)
*/

const (
	//TODO: задать более конкретные имена для ключей (чтобы было видно с каким процессом они связаны)
	absenceUserKey  = "user"
	absenceTypeKey  = "type"
	absenceBeginKey = "begin"
	absenceEndKey   = "end"
)

var (
	// absenceSG - группа состояний absence (префикс). Хранит состояния для добавления причины отсутствия работника.
	absenceSG = fsm.NewStateGroup("absence")

	// Последовательность состояний процесса добавления причины отсутствия работника
	absenceUserState    = absenceSG.New(absenceUserKey)
	absenceTypeState    = absenceSG.New(absenceTypeKey)
	absenceBeginState   = absenceSG.New(absenceBeginKey)
	absenceEndState     = absenceSG.New(absenceEndKey)
	absenceConfirmState = absenceSG.New("confirm")
)

func (c *controller) absenceProcessInit() {
	// Buttons
	c.manager.Bind(&cancelBtn, fsm.AnyState, cancelRegisterHandler)

	// Absence add process
	c.manager.Bind(tele.OnText, absenceUserState, absenceUserHandler)
	c.manager.Bind(tele.OnText, absenceTypeState, absenceTypeHandler)
	c.manager.Bind(tele.OnText, absenceBeginState, absenceBeginHandler)
	c.manager.Bind(tele.OnText, absenceEndState, absenceEndHandler)
	//c.manager.Bind(&confirmRegisterBtn, absenceConfirmState, c.registerConfirmHandler)
	//c.manager.Bind(&resetRegisterBtn, absenceConfirmState, registerResetHandler)
	//c.manager.Bind(&cancelRegisterBtn, absenceConfirmState, cancelRegisterHandler, deleteAfterHandler)
}

func startAbsenceHandler(tc tele.Context, state fsm.Context) error {
	menu := &tele.ReplyMarkup{}
	menu.Reply(menu.Row(cancelBtn))
	menu.ResizeKeyboard = true

	state.Set(absenceUserState)
	return tc.Send("Введите фамилию сотрудника", menu)
}

func absenceUserHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(absenceUserKey, input)

	go state.Set(absenceTypeState)
	return tc.Send("Выберите причину неявки")
}

func absenceTypeHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(absenceTypeKey, input)

	go state.Set(absenceBeginState)
	return tc.Send("Введите начальную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func absenceBeginHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(absenceBeginKey, input)

	go state.Set(absenceEndState)
	return tc.Send("Введите конечную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func absenceEndHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(absenceEndKey, input)

	go state.Set(absenceConfirmState)
	return tc.Send("Данные приняты")
}
