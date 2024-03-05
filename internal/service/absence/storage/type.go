package storage

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/PaulYakow/test-bot/internal/model"
)

type absence struct {
	ID        uint64       `db:"id"`
	UserID    uint64       `db:"user_id"`
	Type      string       `db:"type"`
	DateBegin time.Time    `db:"date_begin"`
	DateEnd   sql.NullTime `db:"date_end"`
}

func convertModelAbsenceToAbsence(ma *model.Absence) absence {
	a := absence{
		UserID:    ma.UserID,
		Type:      ma.Code,
		DateBegin: ma.DateBegin,
	}

	if !ma.DateEnd.IsZero() {
		a.DateEnd.Time = ma.DateEnd
	}

	return a
}

type absenceInfo struct {
	ID          uint64 `db:"id"`
	Description string `db:"description"`
}

func convertAbsenceInfoToModel(ui absenceInfo) model.RecordInfo {
	return model.RecordInfo{
		ID:          strconv.FormatUint(ui.ID, 10),
		Description: ui.Description,
	}
}
