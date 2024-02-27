package storage

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/stdlib"
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

	if err = goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("%s (goose set dialect): %w", op, err)
	}

	db := stdlib.OpenDBFromPool(pool.Pool)
	status := "up"
	if err = db.PingContext(ctx); err != nil {
		status = "down"
	}
	log.Println(fmt.Sprintf("%s (open db from pool): %s", op, status))

	if err = goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("%s (goose up): %w", op, err)
	}

	if err = db.Close(); err != nil {
		return nil, fmt.Errorf("%s (db close): %w", op, err)
	}

	return &storage{
		Pool: pool,
	}, nil
}
