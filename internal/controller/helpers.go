package controller

import (
	"time"

	"github.com/vitaliy-ukiru/fsm-telebot"
)

func dateMessage(d time.Time) string {
	if d.IsZero() {
		return "<u>Не указана</u>"
	}

	return d.Format(dateLayout)
}

func dataFromState[T any](state fsm.Context, key string) T {
	var data T
	state.MustGet(key, &data)

	return data
}
