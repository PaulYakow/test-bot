package user

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

func (s *service) AddUser(ctx context.Context, u model.User) (uint64, error) {
	return s.userStorage.Create(ctx, u)
}
