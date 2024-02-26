package controller

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

type UserService interface {
	AddUser(ctx context.Context, user model.User) (uint64, error)
}
