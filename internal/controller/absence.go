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

/*
user_id bigint NOT NULL - использовать фамилию (+ табельный, если нашлось несколько)
"type" absence_code NOT NULL - выгружать доступные коды из БД и формировать кнопки для выбора
date_begin date NOT NULL -
date_end date - может заполняться позже (предусмотреть пропуск, например, по кнопке)

Вначале можно выдавать список записей (по запросу), в которых date_end IS NULL - для выбора заполнения date_end (сразу переходить к нему)
*/

var (
	absenceAddRecordBtn      = tele.Btn{Text: "🆕 Добавить новую запись", Unique: "absence_add_record"}
	absenceEditRecordBtn     = tele.Btn{Text: "📝 Обновить существующую запись", Unique: "absence_edit_record"}
	absenceUserConfirmBtn    = tele.Btn{Unique: "absence_confirm_user"}
	absenceCodeConfirmBtn    = tele.Btn{Unique: "absence_confirm_code"}
	absenceRecordConfirmBtn  = tele.Btn{Unique: "absence_confirm_record"}
	absenceRestartProcessBtn = tele.Btn{Text: "✅ Да", Unique: "absence_restart_process"}
	absenceCancelProcessBtn  = tele.Btn{Text: "❌ Нет", Unique: "absence_cancel_process"}

	// absenceSG - группа состояний absence (префикс). Хранит состояния для добавления причины отсутствия работника.
	absenceSG = fsm.NewStateGroup("absence")

	// Последовательность состояний процесса добавления причины отсутствия работника
	absenceSelectActionState = absenceSG.New("select_action")
	absenceInputUserState    = absenceSG.New("input_user")
	absenceSelectUserState   = absenceSG.New("select_user")
	absenceNoUserState       = absenceSG.New("no_user")
	absenceSelectRecordState = absenceSG.New("select_record")
	absenceSelectCodeState   = absenceSG.New("select_code")
	absenceBeginState        = absenceSG.New("begin")
	absenceEndState          = absenceSG.New("end")
	absenceConfirmState      = absenceSG.New("confirm")

	absenceLastNameKey = absenceSG.Prefix + "@last_name"
	absenceUserIDKey   = absenceSG.Prefix + "@user_id"
	absenceRecordIDKey = absenceSG.Prefix + "@record_id"
	absenceCodeKey     = absenceSelectCodeState.GoString()
	absenceBeginKey    = absenceBeginState.GoString()
	absenceEndKey      = absenceEndState.GoString()
)

func (c *controller) absenceProcessInit() {
	c.manager.Bind(&absenceAddRecordBtn, absenceSelectActionState, absenceAddRecordHandler, deleteAfterHandler)
	c.manager.Bind(&absenceEditRecordBtn, absenceSelectActionState, c.editRecordsHandler, deleteAfterHandler)

	c.manager.Bind(tele.OnText, absenceInputUserState, c.absenceInputUserHandler)

	c.manager.Bind(&absenceRestartProcessBtn, absenceNoUserState, absenceAddRecordHandler, deleteAfterHandler)
	c.manager.Bind(&absenceCancelProcessBtn, absenceNoUserState, cancelHandler, deleteAfterHandler)

	c.manager.Bind(&absenceUserConfirmBtn, absenceSelectUserState, c.absenceConfirmUserHandler, deleteAfterHandler)

	c.manager.Bind(&absenceRecordConfirmBtn, absenceSelectRecordState, absenceConfirmRecordHandler, deleteAfterHandler)

	c.manager.Bind(&absenceCodeConfirmBtn, absenceSelectCodeState, absenceConfirmCodeHandler, deleteAfterHandler)

	c.manager.Bind(tele.OnText, absenceBeginState, absenceBeginHandler)

	c.manager.Bind(tele.OnText, absenceEndState, c.absenceEndHandler)
	c.manager.Bind(&skipStateBtn, absenceEndState, c.absenceSkipEndHandler)

	c.manager.Bind(&confirmBtn, absenceConfirmState, c.absenceConfirmHandler, deleteAfterHandler)
	c.manager.Bind(&resetBtn, absenceConfirmState, absenceResetHandler)
	c.manager.Bind(&cancelBtn, absenceConfirmState, cancelHandler, deleteAfterHandler)
}

func startAbsenceHandler(tc tele.Context, state fsm.Context) error {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(absenceAddRecordBtn),
		rm.Row(absenceEditRecordBtn),
	)

	rm.ResizeKeyboard = true
	rm.OneTimeKeyboard = true

	state.Set(absenceSelectActionState)
	return tc.Send("❕<b>Выберите действие</b>❕", rm)
}

func absenceAddRecordHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(absenceInputUserState)
	return tc.Send(
		`Введите фамилию (либо начало фамилии) сотрудника.
<i>Регистр ввода не имеет значения.</i>`,
		replyMarkupWithCancel())
}

func (c *controller) absenceInputUserHandler(tc tele.Context, state fsm.Context) error {
	lastName := tc.Message().Text

	// Проверка количества записей:
	// 0 - сотрудники с такой фамилией не найдены (absenceNoUserState)
	// =1 - найден один сотрудник (absenceSelectCodeState)
	// >1 - найдено несколько сотрудников (absenceSelectUserHandler)
	usersInfo, err := c.user.ListWithSpecifiedLastName(context.Background(), lastName)
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send(fmt.Sprintf("Ошибка при поиске %q в БД", lastName))
	}

	switch len(usersInfo) {
	case 0:
		state.Update(absenceLastNameKey, lastName)
		go state.Set(absenceNoUserState)
		return absenceNoUserHandler(tc, state)
	case 1:
		id, err := c.user.IDWithSpecifiedLastName(context.Background(), lastName)
		if err != nil {
			tc.Bot().OnError(err, tc)
			state.Finish(true)
			return tc.Send("Ошибка при поиске id сотрудника в БД")
		}

		state.Update(absenceUserIDKey, id)
		go state.Set(absenceSelectCodeState)
		return c.absenceSelectCodeHandler(tc, state)
	default:
		state.Update(absenceLastNameKey, lastName)
		go state.Set(absenceSelectUserState)
		return c.absenceSelectUserHandler(tc, usersInfo)
	}
}

func absenceNoUserHandler(tc tele.Context, state fsm.Context) error {
	rm := &tele.ReplyMarkup{}
	rm.Inline(
		rm.Row(absenceRestartProcessBtn, absenceCancelProcessBtn),
	)
	rm.ResizeKeyboard = true
	rm.OneTimeKeyboard = true

	return tc.Send(
		fmt.Sprintf(`Сотрудников с фамилией (либо частью фамилии) %q не найдено.
Хотите повторить поиск?`,
			dataFromState[string](state, absenceLastNameKey)),
		rm)
}

func (c *controller) absenceSelectUserHandler(tc tele.Context, usersInfo []model.UserInfo) error {
	rm := &tele.ReplyMarkup{}
	rows := make([]tele.Row, len(usersInfo))
	for i, info := range usersInfo {
		absenceUserConfirmBtn.Text = info.Description
		absenceUserConfirmBtn.Data = info.ID
		rows[i] = rm.Row(absenceUserConfirmBtn)
	}
	rm.Inline(rows...)
	rm.ResizeKeyboard = true
	rm.OneTimeKeyboard = true

	return tc.Send(`❗️ Найдено более одного сотрудника.
Выберите требуемого:`,
		rm)
}

func (c *controller) absenceConfirmUserHandler(tc tele.Context, state fsm.Context) error {
	id, _ := strconv.ParseUint(tc.Callback().Data, 10, 64)

	state.Update(absenceUserIDKey, id)
	go state.Set(absenceSelectCodeState)
	return c.absenceSelectCodeHandler(tc, state)
}

func (c *controller) editRecordsHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(absenceSelectRecordState)

	return c.absenceSelectRecordHandler(tc, state)
}

func (c *controller) absenceSelectRecordHandler(tc tele.Context, state fsm.Context) error {
	absenceList, err := c.absence.ListWithNullEndDate(context.Background())
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send("Ошибка при получении списка причин неявок (с пустой датой окончания) из БД")
	}

	if len(absenceList) == 0 {
		return tc.Send("❗<b>Записей не найдено</b>")
	}

	rows := make([]tele.Row, len(absenceList))
	for i, ai := range absenceList {
		absenceCodeConfirmBtn.Text = ai.Description
		absenceCodeConfirmBtn.Data = ai.ID
		rows[i] = tele.Row{absenceRecordConfirmBtn}
	}
	rm := &tele.ReplyMarkup{}
	rm.Inline(rows...)
	rm.ResizeKeyboard = true

	log.Println("absence select record handler reply keyboard:", rm.InlineKeyboard)

	return tc.Send("<b>Выберите запись</b>", rm)
}

func absenceConfirmRecordHandler(tc tele.Context, state fsm.Context) error {
	id, _ := strconv.ParseUint(tc.Callback().Data, 10, 64)
	go state.Update(absenceRecordIDKey, id)

	go state.Set(absenceEndState)
	return tc.Send("Введите конечную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

func (c *controller) absenceSelectCodeHandler(tc tele.Context, state fsm.Context) error {
	absenceCodes, err := c.absence.ListCodes(context.Background())
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send("Ошибка при получении списка причин неявок из БД")
	}

	btns := make([]tele.Btn, len(absenceCodes))
	for i, code := range absenceCodes {
		absenceCodeConfirmBtn.Text = code
		absenceCodeConfirmBtn.Data = code
		btns[i] = absenceCodeConfirmBtn
	}
	rm := &tele.ReplyMarkup{}
	rm.Inline(rm.Split(len(btns)/2, btns)...)
	rm.ResizeKeyboard = true

	info, err := c.user.InfoWithSpecifiedID(context.Background(), dataFromState[uint64](state, absenceUserIDKey))

	return tc.Send(
		fmt.Sprintf(`Выбран %s.
Выберите причину неявки`, info),
		rm)
}

func absenceConfirmCodeHandler(tc tele.Context, state fsm.Context) error {
	data := tc.Callback().Data
	go state.Update(absenceCodeKey, data)

	go state.Set(absenceBeginState)
	return tc.Send("Введите начальную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)")
}

// TODO: календарь для выбора даты
func absenceBeginHandler(tc tele.Context, state fsm.Context) error {
	rm := replyMarkupWithCancelAndSkip()

	input, err := time.Parse(dateLayout, tc.Message().Text)
	if err != nil {
		return tc.Send("Дата должна иметь формат ДД.ММ.ГГГГ (например, 01.01.2001)")
	}
	go state.Update(absenceBeginKey, input)

	go state.Set(absenceEndState)
	return tc.Send("Введите конечную дату в формате ДД.ММ.ГГГГ (например, 01.01.2001)", rm)
}

// TODO: календарь для выбора даты
func (c *controller) absenceEndHandler(tc tele.Context, state fsm.Context) error {
	input, err := time.Parse(dateLayout, tc.Message().Text)
	if err != nil {
		return tc.Send("Дата должна иметь формат ДД.ММ.ГГГГ (например, 01.01.2001)")
	}
	state.Update(absenceEndKey, input)

	go state.Set(absenceConfirmState)
	return c.absenceCheckData(tc, state, "✅ Данные приняты.")
}

func (c *controller) absenceSkipEndHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(absenceConfirmState)
	return c.absenceCheckData(tc, state,
		`⚠️ Дата окончания пропущена.
Остальные данные приняты.`)
}

func (c *controller) absenceCheckData(tc tele.Context, state fsm.Context, msg string) error {
	a := absenceFromStateStorage(state)

	info, err := c.user.InfoWithSpecifiedID(context.Background(), a.UserID)
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send("Ошибка при поиске сотрудника в БД")
	}

	return tc.Send(fmt.Sprintf(
		`%s

<b>Проверьте данные:</b>
<i>Сотрудник</i>: %q
<i>Причина неявки</i>: %q
<i>Дата начала</i>: %v
<i>Дата окончания</i>: %v`,
		msg,
		info,
		a.Code,
		dateMessage(a.DateBegin),
		dateMessage(a.DateEnd),
	),
		replyMarkupForConfirmState())

}

func (c *controller) absenceConfirmHandler(tc tele.Context, state fsm.Context) error {
	defer state.Finish(true)

	// FIXME: Повтор (c.absenceCheckData)
	a := absenceFromStateStorage(state)
	info, err := c.user.InfoWithSpecifiedID(context.Background(), a.UserID)
	if err != nil {
		tc.Bot().OnError(err, tc)
		state.Finish(true)
		return tc.Send("Ошибка при поиске сотрудника в БД")
	}

	header := `<b>Данные приняты</b>\n`
	footer := ""

	if recordID := dataFromState[uint64](state, absenceRecordIDKey); recordID != 0 {
		err = c.absence.UpdateEndDate(context.Background(), recordID, a.DateEnd)
		if err != nil {
			tc.Bot().OnError(err, tc)
			state.Finish(true)
			return tc.Send("Ошибка при обновлении записи  в БД")
		}

		header = `<b>Данные обновлены</b>\n`
	} else {
		id, err := c.absence.Add(context.Background(), absenceFromStateStorage(state))
		if err != nil {
			tc.Bot().OnError(err, tc)
			return tc.Send(fmt.Sprintf("Ошибка сохранения: %v", err))
		}

		footer = fmt.Sprintf(`\n<u>ID записи: %d</u>`, id)
	}

	return tc.Send(fmt.Sprintf(
		`%s
<i>Сотрудник</i>: %q
<i>Причина неявки</i>: %q
<i>Дата начала</i>: %v
<i>Дата окончания</i>: %v
%s`,
		header,
		info,
		a.Code,
		dateMessage(a.DateBegin),
		dateMessage(a.DateEnd),
		footer,
	),
		tele.RemoveKeyboard)
}

func absenceResetHandler(tc tele.Context, state fsm.Context) error {
	go state.Set(absenceInputUserState)
	return tc.Send(`Начнём заново.
Введите фамилию сотрудника.
`)
}

func absenceFromStateStorage(state fsm.Context) model.Absence {
	return model.Absence{
		UserID:    dataFromState[uint64](state, absenceUserIDKey),
		Code:      dataFromState[string](state, absenceCodeKey),
		DateBegin: dataFromState[time.Time](state, absenceBeginKey),
		DateEnd:   dataFromState[time.Time](state, absenceEndKey),
	}
}
