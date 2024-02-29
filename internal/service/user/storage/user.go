package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"github.com/PaulYakow/test-bot/internal/model"
)

func (s *Storage) Create(ctx context.Context, mu model.User) (uint64, error) {
	const op = "user storage: create user"

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

func (s *Storage) CountUsersByLastName(ctx context.Context, lastName string) (int, error) {
	const op = "user storage: count users by last name"

	lastName += "%"
	log.Println(fmt.Sprintf("%s input: %s", op, lastName))
	row := s.Pool.QueryRow(ctx,
		`SELECT COUNT(*)
			FROM users
			WHERE last_name ILIKE $1;`, lastName)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

func (s *Storage) UserIDByLastName(ctx context.Context, lastName string) (uint64, error) {
	const op = "user storage: user id by last name"

	lastName += "%"
	row := s.Pool.QueryRow(ctx,
		`SELECT id
			FROM users
			WHERE last_name ILIKE '@last_name';`,
		pgx.NamedArgs{"last_name": lastName},
	)

	var id uint64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) ListUsersByLastName(ctx context.Context, lastName string) ([]model.UserInfo, error) {
	const op = "user storage: list users by last name"

	lastName += "%"
	rows, err := s.Pool.Query(ctx,
		`SELECT id,
       				format('%s %s.%s. (%s)', last_name, LEFT(first_name, 1), LEFT(middle_name, 1), service_number) AS description
			FROM users
			WHERE last_name ILIKE '@last_name';`,
		pgx.NamedArgs{"last_name": lastName},
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	uis, err := pgx.CollectRows[userInfo](rows, pgx.RowToStructByNameLax[userInfo])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	infos := make([]model.UserInfo, len(uis))
	for i, ui := range uis {
		infos[i] = convertUserInfoToModel(ui)
	}

	return infos, nil
}
