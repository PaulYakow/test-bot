package storage

import "github.com/PaulYakow/test-bot/internal/model"

type user struct {
	ID            uint64 `db:"id"`
	LastName      string `db:"last_name"`
	FirstName     string `db:"first_name"`
	MiddleName    string `db:"middle_name"`
	Birthday      string `db:"birthday"`
	Position      string `db:"position"`
	ServiceNumber int    `db:"service_number"`
}

func convertModelUserToUser(mu *model.User) user {
	return user{
		ID:            mu.ID,
		LastName:      mu.LastName,
		FirstName:     mu.FirstName,
		MiddleName:    mu.MiddleName,
		Birthday:      mu.Birthday,
		Position:      mu.Position,
		ServiceNumber: mu.ServiceNumber,
	}
}
