package absence

import (
	"context"
	"time"

	"github.com/PaulYakow/test-bot/internal/model"
)

type absenceStorage interface {
	Create(ctx context.Context, absence model.Absence) (uint64, error)
	ListCodes(ctx context.Context) ([]string, error)
	ListByNullEndDate(ctx context.Context) ([]model.AbsenceInfo, error)
	UpdateEndDate(ctx context.Context, id uint64, date time.Time) error
}
