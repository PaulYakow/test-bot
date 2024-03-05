package model

import "time"

type Absence struct {
	UserID    uint64
	Code      string
	DateBegin time.Time
	DateEnd   time.Time
}

// FIXME: Напрашивается объединение в один тип с UserInfo (например, RecordInfo)
type AbsenceInfo struct {
	ID          string
	Description string
}
