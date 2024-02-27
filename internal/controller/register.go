package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot"
	tele "gopkg.in/telebot.v3"
)

const (
	dateLayout       = "02.01.2006"
	lastNameKey      = "last_name"
	firstNameKey     = "first_name"
	middleNameKey    = "middle_name"
	birthdayKey      = "birthday"
	positionKey      = "position"
	serviceNumberKey = "service_number"
)

var (
	// RegisterSG - группа состояний reg (префикс). Хранит состояния для регистрации пользователя.
	RegisterSG = fsm.NewStateGroup("reg")

	// Последовательность состояний процесса регистрации пользователя

	RegisterLastNameState      = RegisterSG.New(lastNameKey)
	RegisterFirstNameState     = RegisterSG.New(firstNameKey)
	RegisterMiddleNameState    = RegisterSG.New(middleNameKey)
	RegisterBirthdayState      = RegisterSG.New(birthdayKey)
	RegisterPositionState      = RegisterSG.New(positionKey)
	RegisterServiceNumberState = RegisterSG.New(serviceNumberKey)
	RegisterConfirmState       = RegisterSG.New("confirm")

	// Кнопки общие для всего процесса (запуск, отмена на любом шаге)
	//regBtn    = tele.Btn{Text: "📝 Начать добавление пользователя"}
	cancelBtn = tele.Btn{Text: "❌ Отменить добавление пользователя"}

	// Кнопки при завершении процесса регистрации
	confirmRegisterBtn = tele.Btn{Text: "✅ Подтвердить и отправить", Unique: "confirm"}
	resetRegisterBtn   = tele.Btn{Text: "🔄 Сбросить", Unique: "reset"}
	cancelRegisterBtn  = tele.Btn{Text: "❌ Отменить", Unique: "cancel"}
)

func (c *controller) registerProcessInit() {
	// Buttons
	//c.manager.Bind(&regBtn, fsm.DefaultState, StartRegisterHandler)
	c.manager.Bind(&cancelBtn, fsm.AnyState, cancelRegisterHandler)

	// User add process
	c.manager.Bind(tele.OnText, RegisterLastNameState, registerLastNameHandler)
	c.manager.Bind(tele.OnText, RegisterFirstNameState, registerFirstNameHandler)
	c.manager.Bind(tele.OnText, RegisterMiddleNameState, registerMiddleNameHandler)
	c.manager.Bind(tele.OnText, RegisterBirthdayState, registerBirthdayHandler)
	c.manager.Bind(tele.OnText, RegisterPositionState, registerPositionHandler)
	c.manager.Bind(tele.OnText, RegisterServiceNumberState, registerServiceNumberHandler)
	c.manager.Bind(&confirmRegisterBtn, RegisterConfirmState, registerConfirmHandler, editFormMessage("Проверьте", "Введённые"))
	c.manager.Bind(&resetRegisterBtn, RegisterConfirmState, registerResetHandler, editFormMessage("Проверьте", "Старые"))
	c.manager.Bind(&cancelRegisterBtn, RegisterConfirmState, cancelRegisterHandler, deleteAfterHandler)
}

func startRegisterHandler(tc tele.Context, state fsm.Context) error {
	menu := &tele.ReplyMarkup{}
	menu.Reply(menu.Row(cancelBtn))
	menu.ResizeKeyboard = true

	fmt.Println("start handler:", state.Set(RegisterLastNameState))
	return tc.Send("Введите фамилию сотрудника", menu)
}

func registerLastNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(lastNameKey, input)

	go fmt.Println("last_name handler:", state.Set(RegisterFirstNameState))
	return tc.Send("Введите имя сотрудника")
}

func registerFirstNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(firstNameKey, input)

	go fmt.Println("first_name handler:", state.Set(RegisterMiddleNameState))
	return tc.Send("Введите отчество сотрудника")
}

func registerMiddleNameHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(middleNameKey, input)

	go fmt.Println("middle_name handler:", state.Set(RegisterBirthdayState))
	return tc.Send("Введите дату рождения сотрудника в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func registerBirthdayHandler(tc tele.Context, state fsm.Context) error {
	input, err := time.Parse(dateLayout, tc.Message().Text)
	if err != nil {
		return tc.Send("Дата должна иметь формат ДД.ММ.ГГГГ (например, 01.01.2001)")
	}
	go state.Update(birthdayKey, input)

	go state.Set(RegisterPositionState)
	return tc.Send("Введите должность сотрудника")
}

func registerPositionHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text
	go state.Update(positionKey, input)

	go state.Set(RegisterServiceNumberState)
	return tc.Send("Введите табельный номер сотрудника")
}

func registerServiceNumberHandler(tc tele.Context, state fsm.Context) error {
	serviceNumber, err := strconv.Atoi(tc.Message().Text)
	if err != nil || serviceNumber <= 0 {
		return tc.Send("Некорректный номер. Попробуйте ещё раз.")
	}
	go state.Update(serviceNumberKey, serviceNumber)

	go state.Set(RegisterConfirmState)

	reply := &tele.ReplyMarkup{}
	reply.Inline(
		reply.Row(confirmRegisterBtn),
		reply.Row(resetRegisterBtn, cancelRegisterBtn),
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
<i>Имя</i>: %d
<i>Отчество</i>: %q
<i>Дата рождения</i>: %v
<i>Должность</i>: %q
<i>Табельный номер</i>: %d`,
		lastName,
		firstName,
		middleName,
		birthday,
		position,
		serviceNumber,
	), reply)
}

func registerConfirmHandler(tc tele.Context, state fsm.Context) error {
	defer state.Finish(true)

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

	// TODO: добавить сохранение в БД
	//if err != nil {
	//	tc.Bot().OnError(err, tc)
	//}
	return tc.Send("Данные приняты", tele.RemoveKeyboard)
}

func registerResetHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(RegisterLastNameState)
	return tc.Send(`Начнём заново.
Введите фамилию сотрудника.
`)
}

func cancelRegisterHandler(tc tele.Context, state fsm.Context) error {
	//menu := &tele.ReplyMarkup{}
	//menu.Reply(menu.Row(regBtn))
	//menu.ResizeKeyboard = true

	go state.Finish(true)
	return tc.Send("Процесс добавления отменён. Введённые данные удалены.")
}
