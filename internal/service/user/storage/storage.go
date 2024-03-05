package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/PaulYakow/test-bot/internal/model"
	pg "github.com/PaulYakow/test-bot/pkg/postgresql"
)

type User struct {
	*pg.Pool
}

func New(p *pg.Pool) (*User, error) {
	return &User{
		Pool: p,
	}, nil
}

func (s *User) Create(ctx context.Context, mu model.User) (uint64, error) {
	const op = "user storage: create"

	u := convertModelUserToUser(&mu)

	row := s.Pool.QueryRow(ctx,
		`INSERT INTO users 
			(last_name, first_name, middle_name, birthday, "position", service_number)
		VALUES
			(@last_name, @first_name, @middle_name, @birthday, @position, @service_number)
			RETURNING id`,
		pgx.NamedArgs{
			"last_name":      u.LastName,
			"first_name":     u.FirstName,
			"middle_name":    u.MiddleName,
			"birthday":       u.Birthday,
			"position":       u.Position,
			"service_number": u.ServiceNumber,
		})

	if err := row.Scan(&u.ID); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return u.ID, nil
}

func (s *User) IDByLastName(ctx context.Context, lastName string) (uint64, error) {
	const op = "user storage: user id by last name"

	lastName += "%"
	log.Println(fmt.Sprintf("%s input: %s", op, lastName))

	row := s.Pool.QueryRow(ctx,
		`SELECT id
			FROM users
			WHERE last_name ILIKE $1;`, lastName)

	var id uint64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *User) ListByLastName(ctx context.Context, lastName string) ([]model.RecordInfo, error) {
	const op = "user storage: list users by last name"

	lastName += "%"
	log.Println(fmt.Sprintf("%s input: %s", op, lastName))

	rows, err := s.Pool.Query(ctx,
		`SELECT id,
       				format('%s %s.%s. (%s)', last_name, LEFT(first_name, 1), LEFT(middle_name, 1), service_number) AS description
			FROM users
			WHERE last_name ILIKE $1;`, lastName)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	uis, err := pgx.CollectRows[userInfo](rows, pgx.RowToStructByNameLax[userInfo])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	infos := make([]model.RecordInfo, len(uis))
	for i, ui := range uis {
		infos[i] = convertUserInfoToModel(ui)
	}

	return infos, nil
}

func (s *User) InfoByID(ctx context.Context, id uint64) (string, error) {
	const op = "user storage: read last name & initials by id"

	log.Println(fmt.Sprintf("%s input: %d", op, id))

	row := s.Pool.QueryRow(ctx,
		`SELECT format('%s %s.%s. (%s)', last_name, LEFT(first_name, 1), LEFT(middle_name, 1), service_number) AS description
			FROM users
			WHERE id=@user_id;`,
		pgx.NamedArgs{"user_id": id})

	var info string
	if err := row.Scan(&info); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return info, nil
}

func (s *User) InfoByAbsenceID(ctx context.Context, id uint64) (string, error) {
	const op = "user storage: read last name & initials by absence id"

	log.Println(fmt.Sprintf("%s input: %d", op, id))

	row := s.Pool.QueryRow(ctx,
		`SELECT format('%s %s.%s. (%s)', last_name, LEFT(first_name, 1), LEFT(middle_name, 1), service_number) AS description
			FROM users
				JOIN absences a on users.id = a.user_id
			WHERE a.id=@absence_id;`,
		pgx.NamedArgs{"absence_id": id})

	var info string
	if err := row.Scan(&info); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return info, nil
}
