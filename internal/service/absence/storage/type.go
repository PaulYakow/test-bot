package storage

import (
	"time"

	"github.com/PaulYakow/test-bot/internal/model"
)

type absence struct {
	ID        uint64    `db:"id"`
	UserID    uint64    `db:"user_id"`
	Type      string    `db:"type"`
	DateBegin time.Time `db:"date_begin"`
	DateEnd   time.Time `db:"date_end"`
}

func convertModelAbsenceToAbsence(ma *model.Absence) absence {
	return absence{
		UserID:    ma.UserID,
		Type:      ma.Code,
		DateBegin: ma.DateBegin,
		DateEnd:   ma.DateEnd,
	}
}
