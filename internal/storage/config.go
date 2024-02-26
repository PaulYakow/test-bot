package storage

import "time"

type Config struct {
	DSN             string        `env:"DSN" env-required:"true"`
	ConnAttempts    int           `env:"CONNECTION_ATTEMPTS" env-default:"5"`
	ConnTimeout     time.Duration `env:"CONNECTION_TIMEOUT" env-default:"1s"`
	MaxOpenConn     int           `env:"MAX_OPEN_CONNECTIONS" env-default:"4"`
	MaxConnIdleTime time.Duration `env:"MAX_CONNECTION_IDLE_TIME" env-default:"15m"`
	MaxConnLifeTime time.Duration `env:"MAX_CONNECTION_LIFE_TIME" env-default:"10m"`
}
