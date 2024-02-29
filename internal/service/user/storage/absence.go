package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func (s *Storage) ListAbsenceCode(ctx context.Context) ([]string, error) {
	const op = "user storage: list absence type"

	rows, err := s.Pool.Query(ctx,
		`SELECT unnest(enum_range(null::absence_code)) AS code`,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	codes, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return codes, nil
}
