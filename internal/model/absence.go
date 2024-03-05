package model

import "time"

type Absence struct {
	UserID    uint64
	Code      string
	DateBegin time.Time
	DateEnd   time.Time
}
