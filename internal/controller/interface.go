package controller

import (
	"context"
	"time"

	"github.com/PaulYakow/test-bot/internal/model"
)

type UserService interface {
	Add(ctx context.Context, user model.User) (uint64, error)
	IDWithSpecifiedLastName(ctx context.Context, lastName string) (uint64, error)
	ListWithSpecifiedLastName(ctx context.Context, lastName string) ([]model.RecordInfo, error)
	InfoWithSpecifiedID(ctx context.Context, id uint64) (string, error)
	InfoWithSpecifiedAbsenceID(ctx context.Context, id uint64) (string, error)
}

type AbsenceService interface {
	Add(ctx context.Context, absence model.Absence) (uint64, error)
	ListCodes(ctx context.Context) ([]string, error)
	ListWithNullEndDate(ctx context.Context) ([]model.RecordInfo, error)
	UpdateEndDate(ctx context.Context, id uint64, date time.Time) error
}
