package app

import (
	"context"
	"fmt"

	"github.com/PaulYakow/test-bot/internal/config"
	"github.com/PaulYakow/test-bot/internal/controller"
	"github.com/PaulYakow/test-bot/internal/service/absence"
	astore "github.com/PaulYakow/test-bot/internal/service/absence/storage"
	"github.com/PaulYakow/test-bot/internal/service/user"
	ustore "github.com/PaulYakow/test-bot/internal/service/user/storage"
	"github.com/PaulYakow/test-bot/internal/storage"
)

// Check interface implementation
var (
	_ controller.UserService    = &user.Service{}
	_ controller.AbsenceService = &absence.Service{}
)

func Run(ctx context.Context, cfg *config.Config) error {
	const op = "app run"

	pgPool, err := storage.New(ctx, cfg.PG)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Create user set
	userStore, err := ustore.New(pgPool.Pool)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	userService := user.New(userStore)

	// Create absence set
	absenceStore, err := astore.New(pgPool.Pool)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	absenceService := absence.New(absenceStore)

	// Create main set of services
	serviceSet := controller.ServiceSet{
		User:    userService,
		Absence: absenceService,
	}

	ctrl, err := controller.New(cfg, serviceSet)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ctrl.Start()

	return nil
}
