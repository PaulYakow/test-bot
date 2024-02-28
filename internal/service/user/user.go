package user

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

func (s *Service) AddUser(ctx context.Context, u model.User) (uint64, error) {
	return s.userStorage.Create(ctx, u)
}

func (s *Service) CountUsersWithLastName(ctx context.Context, lastName string) (int, error) {
	return s.userStorage.CountUsersByLastName(ctx, lastName)
}

func (s *Service) UserIDWithLastName(ctx context.Context, lastName string) (uint64, error) {
	return s.userStorage.UserIDByLastName(ctx, lastName)
}

func (s *Service) ListUsersWithLastName(ctx context.Context, lastName string) ([]model.UserInfo, error) {
	return s.userStorage.ListUsersByLastName(ctx, lastName)
}
