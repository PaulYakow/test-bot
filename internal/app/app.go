package app

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/config"
	"github.com/PaulYakow/test-bot/internal/controller"
	"github.com/PaulYakow/test-bot/internal/service/user"
	ustore "github.com/PaulYakow/test-bot/internal/service/user/storage"
	"github.com/PaulYakow/test-bot/internal/storage"
)

func Run(ctx context.Context, cfg *config.Config) error {
	pgPool, err := storage.New(ctx, cfg.PG)
	if err != nil {
		return err
	}

	userStore, err := ustore.New(pgPool.Pool)
	if err != nil {
		return err
	}

	uService := user.New(userStore)

	ctrl, err := controller.New(cfg, uService)
	if err != nil {
		return err
	}

	ctrl.Start()

	return nil
}
