package model

import "time"

type User struct {
	ID            uint64
	LastName      string
	FirstName     string
	MiddleName    string
	Birthday      time.Time
	Position      string
	ServiceNumber int
}

type UserInfo struct {
	ID          string
	Description string
}
