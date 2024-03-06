package controller

import (
	"fmt"

	tb "gopkg.in/telebot.v3"
)

func testHandler(tc tb.Context) error {
	msg := `❕<b>Выберите действие</b>❕`

	tc.Bot().Handle("\ftest_btn", testBtnHandler)
	tc.Bot().Handle("\faction", actionBtnHandler)
	tc.Bot().Handle("/form", formHandler(selectActionForm()))

	testInfo := []struct {
		ID          string
		Description string
	}{
		{
			ID:          "1",
			Description: "number one",
		},
		{
			ID:          "2",
			Description: "number two",
		},
		{
			ID:          "3",
			Description: "number three",
		},
	}

	inline := &tb.ReplyMarkup{}
	btn := tb.Btn{
		Unique: "test_btn",
	}

	rows := make([]tb.Row, len(testInfo))
	for i, info := range testInfo {
		btn.Text = info.Description
		btn.Data = info.ID
		rows[i] = inline.Row(btn)
	}

	inline.Inline(rows...)
	inline.ResizeKeyboard = true

	return tc.Send(msg, inline)
}

func testBtnHandler(tc tb.Context) error {
	msg := fmt.Sprintf(`Callback info:
Unique: %s
Data: %s`,
		tc.Callback().Unique,
		tc.Callback().Data)

	return tc.Send(msg)
}

type Form struct {
	Message     string
	ReplyMarkup *tb.ReplyMarkup
}

func selectActionForm() Form {
	rm := &tb.ReplyMarkup{}
	rm.Inline(
		rm.Row(tb.Btn{Text: "Action #1", Unique: "action", Data: "action1_data"}),
		rm.Row(tb.Btn{Text: "Action #2", Unique: "action", Data: "action2_data"}),
	)

	rm.ResizeKeyboard = true
	rm.OneTimeKeyboard = true

	return Form{
		Message:     "❕<b>Выберите действие</b>❕",
		ReplyMarkup: rm,
	}
}

func formHandler(f Form) tb.HandlerFunc {
	return func(tc tb.Context) error {
		return tc.Send(f.Message, f.ReplyMarkup)
	}
}

func actionBtnHandler(tc tb.Context) error {
	msg := fmt.Sprintf(`Action callback info:
Unique: %s
Data: %s`,
		tc.Callback().Unique,
		tc.Callback().Data)

	return tc.Send(msg)
}
