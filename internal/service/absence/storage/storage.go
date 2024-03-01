package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/PaulYakow/test-bot/internal/model"
	pg "github.com/PaulYakow/test-bot/pkg/postgresql"
)

type Absence struct {
	*pg.Pool
}

func New(p *pg.Pool) (*Absence, error) {
	return &Absence{
		Pool: p,
	}, nil
}

func (s *Absence) Create(ctx context.Context, ma model.Absence) (uint64, error) {
	const op = "absence storage: create"

	a := convertModelAbsenceToAbsence(&ma)

	row := s.Pool.QueryRow(ctx,
		`INSERT INTO absences
			(user_id, "type", date_begin, date_end)
		VALUES
			(@user_id, @type, @date_begin, @date_end)
			RETURNING id`,
		pgx.NamedArgs{
			"user_id":    a.UserID,
			"type":       a.Type,
			"date_begin": a.DateBegin,
			"date_end":   a.DateEnd,
		})

	if err := row.Scan(&a.ID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return a.ID, nil
}

func (s *Absence) ListCodes(ctx context.Context) ([]string, error) {
	const op = "absence storage: list supported codes"

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
