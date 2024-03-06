package app

import (
	"context"
	"fmt"

	"github.com/PaulYakow/test-bot/internal/config"
	"github.com/PaulYakow/test-bot/internal/controller"
	aService "github.com/PaulYakow/test-bot/internal/service/absence"
	aStore "github.com/PaulYakow/test-bot/internal/service/absence/storage"
	uService "github.com/PaulYakow/test-bot/internal/service/user"
	uStore "github.com/PaulYakow/test-bot/internal/service/user/storage"
	"github.com/PaulYakow/test-bot/internal/storage"
)

// Check interface implementation
var (
	_ controller.UserService    = &uService.Service{}
	_ controller.AbsenceService = &aService.Service{}
)

func Run(ctx context.Context, cfg *config.Config) error {
	const op = "app run"

	pgPool, err := storage.New(ctx, cfg.PG)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Create user set
	userStore, err := uStore.New(pgPool.Pool)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userService := uService.New(userStore)

	// Create absence set
	absenceStore, err := aStore.New(pgPool.Pool)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	absenceService := aService.New(absenceStore)

	// Create main set of services
	set := controller.Set{
		User:    controller.NewUser(userService, nil),
		Absence: controller.NewAbsence(absenceService, nil),
	}

	ctrl, err := controller.New(cfg, set)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ctrl.Start()

	return nil
}
