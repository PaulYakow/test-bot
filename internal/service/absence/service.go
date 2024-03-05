package absence

import (
	"context"
	"time"

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

func (s *Service) ListWithNullEndDate(ctx context.Context) ([]model.AbsenceInfo, error) {
	return s.storage.ListByNullEndDate(ctx)
}

func (s *Service) UpdateEndDate(ctx context.Context, id uint64, date time.Time) error {
	return s.storage.UpdateEndDate(ctx, id, date)
}
