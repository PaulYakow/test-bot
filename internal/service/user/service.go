package user

type Service struct {
	userStorage userStorage
}

func New(us userStorage) *Service {
	return &Service{
		userStorage: us,
	}
}
