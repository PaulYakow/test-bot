package user

import "github.com/PaulYakow/test-bot/internal/service/user/storage"

var (
	_ userStorage = &storage.Storage{}
)

type Service struct {
	userStorage userStorage
}

func New(us userStorage) *Service {
	return &Service{
		userStorage: us,
	}
}
