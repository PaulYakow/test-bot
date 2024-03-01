package absence

import (
	"context"

	"github.com/PaulYakow/test-bot/internal/model"
)

type absenceStorage interface {
	Create(ctx context.Context, absence model.Absence) (uint64, error)
	ListCodes(ctx context.Context) ([]string, error)
}
