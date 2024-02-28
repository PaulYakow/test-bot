package controller

import (
	"fmt"

	tele "gopkg.in/telebot.v3"
)

func testHandler(tc tele.Context) error {
	msg := `❕<b>Выберите действие</b>❕`

	tc.Bot().Handle("\ftest_btn", testBtnHandler)

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

	inline := &tele.ReplyMarkup{}
	for _, info := range testInfo {
		inline.Inline(
			inline.Row(tele.Btn{
				Text:   info.Description,
				Data:   info.ID,
				Unique: "test_btn",
			}),
		)
	}

	inline.ResizeKeyboard = true

	return tc.Send(msg, inline)
}

func testBtnHandler(tc tele.Context) error {
	msg := fmt.Sprintf(`Callback info:
Unique: %s
Data: %s`,
		tc.Callback().Unique,
		tc.Callback().Data)

	return tc.Send(msg)
}
