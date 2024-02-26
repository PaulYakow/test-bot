package storage

import pg "github.com/PaulYakow/test-bot/pkg/postgresql"

type storage struct {
	*pg.Pool
}

func New(p *pg.Pool) (*storage, error) {
	return &storage{
		Pool: p,
	}, nil
}
