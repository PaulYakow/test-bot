package app

import (
	"context"
	"fmt"

	"github.com/PaulYakow/test-bot/internal/config"
	"github.com/PaulYakow/test-bot/internal/controller"
	"github.com/PaulYakow/test-bot/internal/service/user"
	ustore "github.com/PaulYakow/test-bot/internal/service/user/storage"
	"github.com/PaulYakow/test-bot/internal/storage"
)

var (
	_ controller.UserService = &user.Service{}
)

func Run(ctx context.Context, cfg *config.Config) error {
	const op = "app run"

	pgPool, err := storage.New(ctx, cfg.PG)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userStore, err := ustore.New(pgPool.Pool)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	uService := user.New(userStore)

	ctrl, err := controller.New(cfg, uService)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ctrl.Start()

	return nil
}
