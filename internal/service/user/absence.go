package user

import "context"

func (s *Service) ListAbsenceCode(ctx context.Context) ([]string, error) {
	return s.userStorage.ListAbsenceCode(ctx)
}
