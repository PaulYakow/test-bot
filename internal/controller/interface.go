package controller

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

type UserService interface {
	Add(ctx context.Context, user model.User) (uint64, error)
	NumberWithSpecifiedLastName(ctx context.Context, lastName string) (int, error)
	IDWithSpecifiedLastName(ctx context.Context, lastName string) (uint64, error)
	ListWithSpecifiedLastName(ctx context.Context, lastName string) ([]model.UserInfo, error)
	InfoWithSpecifiedID(ctx context.Context, id uint64) (string, error)
}

type AbsenceService interface {
	Add(ctx context.Context, absence model.Absence) (uint64, error)
	ListCodes(ctx context.Context) ([]string, error)
}
