package storage

import pg "github.com/PaulYakow/test-bot/pkg/postgresql"

type Storage struct {
	*pg.Pool
}

func New(p *pg.Pool) (*Storage, error) {
	return &Storage{
		Pool: p,
	}, nil
}
