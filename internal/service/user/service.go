package user

type service struct {
	userStorage userStorage
}

func New(us userStorage) *service {
	return &service{
		userStorage: us,
	}
}
