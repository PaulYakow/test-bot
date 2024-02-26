package postgresql

import (
	"time"
)

// Option применяет заданную настройку к репозиторию (Pool).
type Option func(cfg *config)

// ConnAttempts задаёт количество попыток подключения.
func ConnAttempts(attempts int) Option {
	return func(cfg *config) {
		cfg.connAttempts = attempts
	}
}

// ConnTimeout задаёт таймаут между попытками подключения.
func ConnTimeout(timeout time.Duration) Option {
	return func(cfg *config) {
		cfg.connAttemptsTimeout = timeout
	}
}

// MaxOpenConn задаёт максимальное количество подключений к БД
func MaxOpenConn(size int) Option {
	return func(cfg *config) {
		cfg.MaxConns = int32(size)
	}
}

// MaxConnIdleTime задаёт время, после которого бездействующее соединение будет закрыто.
func MaxConnIdleTime(duration time.Duration) Option {
	return func(cfg *config) {
		cfg.MaxConnIdleTime = duration
	}
}

// MaxConnLifeTime задаёт время с момента создания, после которого соединение будет закрыто.
func MaxConnLifeTime(duration time.Duration) Option {
	return func(cfg *config) {
		cfg.MaxConnLifetime = duration
	}
}
