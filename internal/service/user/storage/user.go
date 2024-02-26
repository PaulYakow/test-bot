package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/PaulYakow/test-bot/internal/model"
	"github.com/PaulYakow/test-bot/pkg/repoerr"
)

func (s *storage) Create(ctx context.Context, mu model.User) (uint64, error) {
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
		if strings.Contains(err.Error(), "23") { // Integrity Constraint Violation
			if strings.Contains(err.Error(), "department_id") {
				return 0, fmt.Errorf("the department does not exist: %w", repoerr.ErrConflict)
			}
			if strings.Contains(err.Error(), "passport_id") {
				return 0, fmt.Errorf("the position does not exist: %w", repoerr.ErrConflict)
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return u.ID, nil
}
