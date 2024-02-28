package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot"
	tele "gopkg.in/telebot.v3"

	"github.com/PaulYakow/test-bot/internal/model"
)

const (
	lastNameKey      = "last_name"
	firstNameKey     = "first_name"
	middleNameKey    = "middle_name"
	birthdayKey      = "birthday"
	positionKey      = "position"
	serviceNumberKey = "service_number"
)

var (
	// registerSG - группа состояний reg (префикс). Хранит состояния для регистрации пользователя.
	registerSG = fsm.NewStateGroup("reg")

	// Последовательность состояний процесса регистрации пользователя
	registerLastNameState      = registerSG.New(lastNameKey)
	registerFirstNameState     = registerSG.New(firstNameKey)
	registerMiddleNameState    = registerSG.New(middleNameKey)
	registerBirthdayState      = registerSG.New(birthdayKey)
	registerPositionState      = registerSG.New(positionKey)
	registerServiceNumberState = registerSG.New(serviceNumberKey)
	registerConfirmState       = registerSG.New("confirm")
)

func (c *controller) registerProcessInit() {
	// User add process
	c.manager.Bind(tele.OnText, registerLastNameState, registerLastNameHandler)
	c.manager.Bind(tele.OnText, registerFirstNameState, registerFirstNameHandler)
	c.manager.Bind(tele.OnText, registerMiddleNameState, registerMiddleNameHandler)
	c.manager.Bind(tele.OnText, registerBirthdayState, registerBirthdayHandler)
	c.manager.Bind(tele.OnText, registerPositionState, registerPositionHandler)
	c.manager.Bind(tele.OnText, registerServiceNumberState, registerServiceNumberHandler)
	c.manager.Bind(&confirmBtn, registerConfirmState, c.registerConfirmHandler, editFormMessage("Проверьте", "Введённые"))
	c.manager.Bind(&resetBtn, registerConfirmState, registerResetHandler, editFormMessage("Проверьте", "Старые"))
	c.manager.Bind(&cancelBtn, registerConfirmState, cancelHandler, deleteAfterHandler)
}

func startRegisterHandler(tc tele.Context, state fsm.Context) error {
	menu := &tele.ReplyMarkup{}
	menu.Reply(menu.Row(cancelProcessBtn))
	menu.ResizeKeyboard = true

	state.Set(registerLastNameState)
	return tc.Send("Введите фамилию сотрудника", menu)
}

func registerLastNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(lastNameKey, input)

	go state.Set(registerFirstNameState)
	return tc.Send("Введите имя сотрудника")
}

func registerFirstNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(firstNameKey, input)

	go state.Set(registerMiddleNameState)
	return tc.Send("Введите отчество сотрудника")
}

func registerMiddleNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(middleNameKey, input)

	go state.Set(registerBirthdayState)
	return tc.Send("Введите дату рождения сотрудника в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func registerBirthdayHandler(tc tele.Context, state fsm.Context) error {
	input, err := time.Parse(dateLayout, tc.Message().Text)
	if err != nil {
		return tc.Send("Дата должна иметь формат ДД.ММ.ГГГГ (например, 01.01.2001)")
	}
	go state.Update(birthdayKey, input)

	go state.Set(registerPositionState)
	return tc.Send("Введите должность сотрудника")
}

func registerPositionHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(positionKey, input)

	go state.Set(registerServiceNumberState)
	return tc.Send("Введите табельный номер сотрудника")
}

func registerServiceNumberHandler(tc tele.Context, state fsm.Context) error {
	serviceNumber, err := strconv.Atoi(tc.Message().Text)
	if err != nil || serviceNumber <= 0 {
		return tc.Send("Некорректный номер. Попробуйте ещё раз.")
	}
	go state.Update(serviceNumberKey, serviceNumber)

	go state.Set(registerConfirmState)

	reply := &tele.ReplyMarkup{}
	reply.Inline(
		reply.Row(confirmBtn),
		reply.Row(resetBtn, cancelBtn),
	)

	var (
		lastName   string
		firstName  string
		middleName string
		birthday   time.Time
		position   string
	)
	state.MustGet(lastNameKey, &lastName)
	state.MustGet(firstNameKey, &firstName)
	state.MustGet(middleNameKey, &middleName)
	state.MustGet(birthdayKey, &birthday)
	state.MustGet(positionKey, &position)

	return tc.Send(fmt.Sprintf(
		`<b>Проверьте данные:</b>
<i>Фамилия</i>: %q
<i>Имя</i>: %q
<i>Отчество</i>: %q
<i>Дата рождения</i>: %v
<i>Должность</i>: %q
<i>Табельный номер</i>: %d`,
		lastName,
		firstName,
		middleName,
		birthday.Format(dateLayout),
		position,
		serviceNumber,
	), reply)
}

func (c *controller) registerConfirmHandler(tc tele.Context, state fsm.Context) error {
	defer state.Finish(true)

	var (
		lastName      string
		firstName     string
		middleName    string
		birthday      time.Time
		position      string
		serviceNumber int
	)
	state.MustGet(lastNameKey, &lastName)
	state.MustGet(firstNameKey, &firstName)
	state.MustGet(middleNameKey, &middleName)
	state.MustGet(birthdayKey, &birthday)
	state.MustGet(positionKey, &position)
	state.MustGet(serviceNumberKey, &serviceNumber)

	id, err := c.us.AddUser(context.Background(), model.User{
		LastName:      lastName,
		FirstName:     firstName,
		MiddleName:    middleName,
		Birthday:      birthday.Format(dateLayout),
		Position:      position,
		ServiceNumber: serviceNumber,
	})
	if err != nil {
		tc.Bot().OnError(err, tc)
		return tc.Send(fmt.Sprintf("Ошибка сохранения: %v", err))
	}

	return tc.Send(fmt.Sprintf("Данные приняты. ID нового сотрудника: %d", id), tele.RemoveKeyboard)
}

func registerResetHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(registerLastNameState)
	return tc.Send(`Начнём заново.
Введите фамилию сотрудника.
`)
}
