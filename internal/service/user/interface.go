package user

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

type userStorage interface {
	Create(ctx context.Context, user model.User) (uint64, error)
	IDByLastName(ctx context.Context, lastName string) (uint64, error)
	ListByLastName(ctx context.Context, lastName string) ([]model.RecordInfo, error)
	InfoByID(ctx context.Context, userID uint64) (string, error)
	InfoByAbsenceID(ctx context.Context, id uint64) (string, error)
	//Exist(ctx context.Context, userID uint64) (bool, error)
	//Read(ctx context.Context, userID uint64) (*model.User, error)
	//Update(ctx context.Context, user model.User) error
	//Delete(ctx context.Context, userID uint64) error
}
