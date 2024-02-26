package user

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

type userStorage interface {
	//Exist(ctx context.Context, userID uint64) (bool, error)
	Create(ctx context.Context, user model.User) (uint64, error)
	//Read(ctx context.Context, userID uint64) (*model.User, error)
	//Update(ctx context.Context, user model.User) error
	//Delete(ctx context.Context, userID uint64) error
}
