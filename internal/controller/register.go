package controller

import (
	"context"
	"fmt"
	"log"
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
	state.Set(registerLastNameState)
	return tc.Send("Введите фамилию сотрудника", replyMarkupWithCancel())
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

	u := userFromStateStorage(state)

	return tc.Send(fmt.Sprintf(
		`<b>Проверьте данные:</b>
<i>Фамилия</i>: %q
<i>Имя</i>: %q
<i>Отчество</i>: %q
<i>Дата рождения</i>: %v
<i>Должность</i>: %q
<i>Табельный номер</i>: %d`,
		u.LastName,
		u.FirstName,
		u.MiddleName,
		u.Birthday.Format(dateLayout),
		u.Position,
		serviceNumber,
	),
		replyMarkupForConfirmState())
}

func (c *controller) registerConfirmHandler(tc tele.Context, state fsm.Context) error {
	defer state.Finish(true)

	id, err := c.user.Add(context.Background(), userFromStateStorage(state))
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

func userFromStateStorage(state fsm.Context) model.User {
	mu := model.User{}

	state.MustGet(lastNameKey, mu.LastName)
	state.MustGet(firstNameKey, mu.FirstName)
	state.MustGet(middleNameKey, mu.MiddleName)
	state.MustGet(birthdayKey, mu.Birthday)
	state.MustGet(positionKey, mu.Position)
	state.MustGet(serviceNumberKey, mu.ServiceNumber)

	log.Printf("register user: from state storage %v\n", mu)

	return mu
}
