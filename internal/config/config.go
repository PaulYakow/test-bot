package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"

	"github.com/PaulYakow/test-bot/internal/storage"
)

type Config struct {
	Token       string         `env:"TG_TOKEN" env-required:"true"`
	SuperuserID string         `env:"TG_SUPERUSER_ID"`
	WebhookURL  string         `env:"WEBHOOK_URL" env-required:"true"`
	WebhookPort string         `env:"WEBHOOK_PORT" env-required:"true"`
	PG          storage.Config `env-prefix:"PG_"`
}

// New создаёт объект Config.
func New() (*Config, error) {
	const op = "config new"

	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return cfg, nil
}
