package user

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
	"github.com/PaulYakow/test-bot/internal/service/user/storage"
)

var (
	_ userStorage = &storage.User{}
)

type Service struct {
	storage userStorage
}

func New(us userStorage) *Service {
	return &Service{
		storage: us,
	}
}

func (s *Service) Add(ctx context.Context, u model.User) (uint64, error) {
	return s.storage.Create(ctx, u)
}

func (s *Service) IDWithSpecifiedLastName(ctx context.Context, lastName string) (uint64, error) {
	return s.storage.IDByLastName(ctx, lastName)
}

func (s *Service) ListWithSpecifiedLastName(ctx context.Context, lastName string) ([]model.RecordInfo, error) {
	return s.storage.ListByLastName(ctx, lastName)
}

func (s *Service) InfoWithSpecifiedID(ctx context.Context, id uint64) (string, error) {
	return s.storage.InfoByID(ctx, id)
}

func (s *Service) InfoWithSpecifiedAbsenceID(ctx context.Context, id uint64) (string, error) {
	return s.storage.InfoByAbsenceID(ctx, id)
}
