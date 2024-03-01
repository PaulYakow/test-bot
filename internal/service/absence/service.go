package absence

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
	"github.com/PaulYakow/test-bot/internal/service/absence/storage"
)

var (
	_ absenceStorage = &storage.Absence{}
)

type Service struct {
	storage absenceStorage
}

func New(as absenceStorage) *Service {
	return &Service{
		storage: as,
	}
}

func (s *Service) ListCodes(ctx context.Context) ([]string, error) {
	return s.storage.ListCodes(ctx)
}

func (s *Service) Add(ctx context.Context, absence model.Absence) (uint64, error) {
	return s.storage.Create(ctx, absence)
}
