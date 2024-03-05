package storage

import (
	"context"
	"fmt"
	"time"

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

func (s *Absence) ListByNullEndDate(ctx context.Context) ([]model.AbsenceInfo, error) {
	const op = "absence storage: list absences by null end date"

	rows, err := s.Pool.Query(ctx,
		`SELECT a.id,
       				format('%s %s.%s. - %s (%s - н.в.)',
							last_name,
              				LEFT(first_name, 1),
              				LEFT(middle_name, 1),
              				"type",
              				to_char(date_begin, 'DD.MM.YYYY')) AS description
			FROM users
         			JOIN absences a on users.id = a.user_id
			WHERE date_end IS NULL
			ORDER BY a.id;`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	absenceList, err := pgx.CollectRows[absenceInfo](rows, pgx.RowToStructByNameLax[absenceInfo])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	infos := make([]model.AbsenceInfo, len(absenceList))
	for i, ai := range absenceList {
		infos[i] = convertAbsenceInfoToModel(ai)
	}

	return infos, nil
}

func (s *Absence) UpdateEndDate(ctx context.Context, id uint64, date time.Time) error {
	const op = "absence storage: update end date"

	_, err := s.Pool.Exec(ctx,
		`UPDATE absences SET date_end = @date_end
				WHERE id = @id;`,
		pgx.NamedArgs{
			"id":       id,
			"date_end": date,
		})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
