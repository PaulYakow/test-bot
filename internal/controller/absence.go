package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot"
	tele "gopkg.in/telebot.v3"

	"github.com/PaulYakow/test-bot/internal/model"
)

/*
user_id bigint NOT NULL - использовать фамилию (+ табельный, если нашлось несколько)
"type" absence_code NOT NULL - выгружать доступные коды из БД и формировать кнопки для выбора
date_begin date NOT NULL -
date_end date - может заполняться позже (предусмотреть пропуск, например, по кнопке)

Вначале можно выдавать список записей (по запросу), в которых date_end IS NULL - для выбора заполнения date_end (сразу переходить к нему)
*/

var (
	absenceAddRecordBtn  = tele.Btn{Text: "🆕 Добавить новую запись"}
	absenceEditRecordBtn = tele.Btn{Text: "📝 Обновить существующую запись"}
	absenceSkipEndBtn    = tele.Btn{Text: "↪️ Пропустить", Unique: "skip"}

	// absenceSG - группа состояний absence (префикс). Хранит состояния для добавления причины отсутствия работника.
	absenceSG = fsm.NewStateGroup("absence")

	// Последовательность состояний процесса добавления причины отсутствия работника
	absenceSelectActionState = absenceSG.New("select_action")
	absenceInputUserState    = absenceSG.New("input_user")
	absenceSelectUserState   = absenceSG.New("select_user")
	absenceNoUserState       = absenceSG.New("no_user")
	absenceSelectRecordState = absenceSG.New("select_record")
	absenceSelectTypeState   = absenceSG.New("select_type")
	absenceBeginState        = absenceSG.New("begin")
	absenceEndState          = absenceSG.New("end")
	absenceConfirmState      = absenceSG.New("confirm")

	absenceLastNameKey = absenceSG.Prefix + "@last_name"
	absenceUserIDKey   = absenceSG.Prefix + "@user_id"
	absenceTypeKey     = absenceSelectTypeState.GoString()
	absenceBeginKey    = absenceBeginState.GoString()
	absenceEndKey      = absenceEndState.GoString()
)

func (c *controller) absenceProcessInit() {
	c.manager.Bind(&absenceAddRecordBtn, absenceSelectActionState, absenceAddRecordHandler, deleteAfterHandler)
	c.manager.Bind(&absenceEditRecordBtn, absenceSelectActionState, absenceEditRecordHandler, deleteAfterHandler)
	c.manager.Bind(tele.OnText, absenceInputUserState, c.absenceInputUserHandler)
	c.manager.Bind(tele.OnText, absenceNoUserState, absenceNoUserHandler)
	c.manager.Bind(tele.OnText, absenceSelectUserState, c.absenceSelectUserHandler)
	c.manager.Bind(tele.OnText, absenceSelectRecordState, absenceSelectRecordHandler)
	c.manager.Bind(tele.OnText, absenceSelectTypeState, absenceSelectTypeHandler)
	c.manager.Bind(tele.OnText, absenceBeginState, absenceBeginHandler)
	c.manager.Bind(tele.OnText, absenceEndState, absenceEndHandler)
	c.manager.Bind(&absenceSkipEndBtn, absenceEndState, absenceSkipEndHandler)
	c.manager.Bind(&confirmBtn, absenceConfirmState, c.absenceConfirmHandler)
	c.manager.Bind(&resetBtn, absenceConfirmState, absenceResetHandler)
	c.manager.Bind(&cancelBtn, absenceConfirmState, cancelHandler, deleteAfterHandler)
}

func startAbsenceHandler(tc tele.Context, state fsm.Context) error {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(absenceAddRecordBtn),
		rm.Row(absenceEditRecordBtn))

	rm.Reply(rm.Row(cancelProcessBtn))
	rm.ResizeKeyboard = true

	state.Set(absenceSelectActionState)
	return tc.Send("❕<b>Выберите действие</b>❕", rm)
}

func absenceAddRecordHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(absenceInputUserState)
	return tc.Send(
		`Введите фамилию (либо начало фамилии) сотрудника.
<i>Регистр ввода не имеет значения.</i>`)
}

func absenceEditRecordHandler(tc tele.Context, state fsm.Context) error {
	// TODO: необходим список записей (в виде кнопок), в которых date_end IS NULL: "Фамилия И.О. - Причина (Дата начала)"
	return nil
}

func (c *controller) absenceInputUserHandler(tc tele.Context, state fsm.Context) error {
	input := tc.Message().Text

	// Проверка количества записей:
	// 0 - сотрудники с такой фамилией не найдены (absenceNoUserState)
	// =1 - найден один сотрудник (absenceSelectTypeState)
	// >1 - найдено несколько сотрудников (absenceSelectUserHandler)
	count, err := c.us.CountUsersWithLastName(context.Background(), input)
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send("Ошибка при поиске сотрудника в БД")
	}

	switch count {
	case 0:
		go state.Set(absenceNoUserState)
		return tc.Send("Сотрудников с заданной фамилией не найдено")
	case 1:
		id, err := c.us.UserIDWithLastName(context.Background(), input)
		if err != nil {
			tc.Bot().OnError(err, tc)
			state.Finish(true)
			return tc.Send("Ошибка при поиске id сотрудника в БД")
		}

		go state.Update(absenceUserIDKey, id)
		go state.Set(absenceSelectTypeState)
		return tc.Send("Выберите причину неявки")
	default:
		go state.Update(absenceLastNameKey, input)
		go state.Set(absenceSelectUserState)
		return tc.Send(`❗️ Найдено более одного сотрудника.
Выберите требуемого:`)
	}
}

func (c *controller) absenceSelectUserHandler(tc tele.Context, state fsm.Context) error {
	// TODO: необходим список пользователей (в виде кнопок): "Фамилия И.О. (Таб. №)"
	var lastName string
	state.MustGet(absenceLastNameKey, &lastName)

	usersInfo, err := c.us.ListUsersWithLastName(context.Background(), lastName)
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send("Ошибка при поиске сотрудников в БД")
	}

	inline := &tele.ReplyMarkup{}
	for _, info := range usersInfo {
		inline.Inline(
			inline.Row(tele.Btn{
				Text:   info.Description,
				Data:   info.ID,
				Unique: "",
			}),
		)
	}

	inline.ResizeKeyboard = true

	return nil
}

func absenceNoUserHandler(tc tele.Context, state fsm.Context) error {
	return nil
}

func absenceSelectRecordHandler(tc tele.Context, state fsm.Context) error {
	return nil
}

func absenceSelectTypeHandler(tc tele.Context, state fsm.Context) error {
	// TODO: необходим список причин (в виде кнопок)
	input := tc.Message().Text
	go state.Update(absenceTypeKey, input)

	go state.Set(absenceBeginState)
	return tc.Send("Введите начальную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func absenceBeginHandler(tc tele.Context, state fsm.Context) error {
	// TODO: календарь для выбора даты
	input := tc.Message().Text
	go state.Update(absenceBeginKey, input)

	go state.Set(absenceEndState)
	return tc.Send("Введите конечную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func absenceEndHandler(tc tele.Context, state fsm.Context) error {
	// TODO: календарь для выбора даты
	input := tc.Message().Text
	go state.Update(absenceEndKey, input)

	go state.Set(absenceConfirmState)
	return tc.Send("Данные приняты")
}

func absenceSkipEndHandler(tc tele.Context, state fsm.Context) error {
	return nil
}

func (c *controller) absenceConfirmHandler(tc tele.Context, state fsm.Context) error {
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

func absenceResetHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(absenceInputUserState)
	return tc.Send(`Начнём заново.
Введите фамилию сотрудника.
`)
}
