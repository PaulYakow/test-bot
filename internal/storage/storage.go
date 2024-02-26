package storage

import (
	"context"
	"fmt"

	pg "github.com/PaulYakow/test-bot/pkg/postgresql"
)

type storage struct {
	*pg.Pool
}

func New(ctx context.Context, cfg Config) (*storage, error) {
	const op = "storage: new"

	pool, err := pg.New(ctx, cfg.DSN,
		pg.ConnAttempts(cfg.ConnAttempts),
		pg.ConnTimeout(cfg.ConnTimeout),
		pg.MaxOpenConn(cfg.MaxOpenConn),
		pg.MaxConnIdleTime(cfg.MaxConnIdleTime),
		pg.MaxConnLifeTime(cfg.MaxConnLifeTime))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO: add migration

	return &storage{
		Pool: pool,
	}, nil
}
