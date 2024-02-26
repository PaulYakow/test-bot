package config

import (
	"github.com/ilyakaznacheev/cleanenv"

	"github.com/PaulYakow/test-bot/internal/storage"
)

type Config struct {
	Token       string         `env:"TG_TOKEN" env-required:"true"`
	WebhookURL  string         `env:"WEBHOOK_URL" env-required:"true"`
	WebhookPort string         `env:"WEBHOOK_PORT" env-required:"true"`
	PG          storage.Config `env-prefix:"PG_"`
}

// New создаёт объект Config.
func New() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
