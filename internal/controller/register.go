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

var (
	// registerSG - группа состояний reg (префикс). Хранит состояния для регистрации пользователя.
	registerSG = fsm.NewStateGroup("user")

	// Последовательность состояний процесса регистрации пользователя
	registerLastNameState      = registerSG.New("last_name")
	registerFirstNameState     = registerSG.New("first_name")
	registerMiddleNameState    = registerSG.New("middle_name")
	registerBirthdayState      = registerSG.New("birthday")
	registerPositionState      = registerSG.New("position")
	registerServiceNumberState = registerSG.New("service_number")
	registerConfirmState       = registerSG.New("confirm")

	registerLastNameKey      = registerLastNameState.GoString()
	registerFirstNameKey     = registerFirstNameState.GoString()
	registerMiddleNameKey    = registerMiddleNameState.GoString()
	registerBirthdayKey      = registerBirthdayState.GoString()
	registerPositionKey      = registerPositionState.GoString()
	registerServiceNumberKey = registerServiceNumberState.GoString()
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
	go state.Update(registerLastNameKey, input)

	go state.Set(registerFirstNameState)
	return tc.Send("Введите имя сотрудника")
}

func registerFirstNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(registerFirstNameKey, input)

	go state.Set(registerMiddleNameState)
	return tc.Send("Введите отчество сотрудника")
}

func registerMiddleNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(registerMiddleNameKey, input)

	go state.Set(registerBirthdayState)
	return tc.Send("Введите дату рождения сотрудника в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func registerBirthdayHandler(tc tele.Context, state fsm.Context) error {
	input, err := time.Parse(dateLayout, tc.Message().Text)
	if err != nil {
		return tc.Send("Дата должна иметь формат ДД.ММ.ГГГГ (например, 01.01.2001)")
	}
	go state.Update(registerBirthdayKey, input)

	go state.Set(registerPositionState)
	return tc.Send("Введите должность сотрудника")
}

func registerPositionHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(registerPositionKey, input)

	go state.Set(registerServiceNumberState)
	return tc.Send("Введите табельный номер сотрудника")
}

func registerServiceNumberHandler(tc tele.Context, state fsm.Context) error {
	serviceNumber, err := strconv.Atoi(tc.Message().Text)
	if err != nil || serviceNumber <= 0 {
		return tc.Send("Некорректный номер. Попробуйте ещё раз.")
	}
	go state.Update(registerServiceNumberKey, serviceNumber)

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
	return model.User{
		LastName:      dataFromState[string](state, registerLastNameKey),
		FirstName:     dataFromState[string](state, registerFirstNameKey),
		MiddleName:    dataFromState[string](state, registerMiddleNameKey),
		Birthday:      dataFromState[time.Time](state, registerBirthdayKey),
		Position:      dataFromState[string](state, registerPositionKey),
		ServiceNumber: dataFromState[int](state, registerServiceNumberKey),
	}
}
