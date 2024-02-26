package postgresql

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultConnAttempts        = 5
	defaultConnAttemptsTimeout = time.Second
)

type config struct {
	*pgxpool.Config
	connAttempts        int
	connAttemptsTimeout time.Duration
}

// Pool структура с настройками подключения к БД и доступом к текущему соединению.
type Pool struct {
	*pgxpool.Pool
	cfg config
}

// New создаёт объект Pool с заданными параметрами и подключается к БД.
func New(ctx context.Context, dsn string, opts ...Option) (*Pool, error) {
	const op = "postgresql: new db"

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db := &Pool{
		cfg: config{
			Config:              cfg,
			connAttempts:        defaultConnAttempts,
			connAttemptsTimeout: defaultConnAttemptsTimeout,
		},
	}

	for _, opt := range opts {
		opt(&db.cfg)
	}

	for db.cfg.connAttempts > 0 {
		if db.Pool, err = pgxpool.NewWithConfig(ctx, db.cfg.Config); err == nil {
			break
		}

		slog.Info(fmt.Sprintf("trying to connect: attempts left %v", db.cfg.connAttempts))

		time.Sleep(db.cfg.connAttemptsTimeout)

		db.cfg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return db, nil
}

// Close дожидается завершения запросов и закрывает все открытые соединения.
func (db *Pool) Close() {
	db.Pool.Close()
}
