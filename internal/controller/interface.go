package controller

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

type UserService interface {
	AddUser(ctx context.Context, user model.User) (uint64, error)
	CountUsersWithLastName(ctx context.Context, lastName string) (int, error)
	UserIDWithLastName(ctx context.Context, lastName string) (uint64, error)
	ListUsersWithLastName(ctx context.Context, lastName string) ([]model.UserInfo, error)
}
