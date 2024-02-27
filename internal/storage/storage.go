package storage

import (
	"context"
	"embed"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	pg "github.com/PaulYakow/test-bot/pkg/postgresql"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type storage struct {
	*pg.Pool
}

func New(ctx context.Context, cfg Config) (*storage, error) {
	const op = "storage new"

	pool, err := pg.New(ctx, cfg.DSN,
		pg.ConnAttempts(cfg.ConnAttempts),
		pg.ConnTimeout(cfg.ConnTimeout),
		pg.MaxOpenConn(cfg.MaxOpenConn),
		pg.MaxConnIdleTime(cfg.MaxConnIdleTime),
		pg.MaxConnLifeTime(cfg.MaxConnLifeTime))
	if err != nil {
		return nil, fmt.Errorf("%s (create pool): %w", op, err)
	}

	goose.SetBaseFS(embedMigrations)

	sql, err := goose.OpenDBWithDriver("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%s (goose create connection): %w", op, err)
	}
	defer sql.Close()

	if err = goose.Up(sql, "migrations"); err != nil {
		return nil, fmt.Errorf("%s (goose up): %w", op, err)
	}

	return &storage{
		Pool: pool,
	}, nil
}
