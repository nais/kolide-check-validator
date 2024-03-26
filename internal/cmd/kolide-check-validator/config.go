package kolide_check_validator

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	KolideApiToken string `env:"KOLIDE_API_TOKEN"`
	SlackWebhook   string `env:"SLACK_WEBHOOK"`
	LogFormat      string `env:"LOG_FORMAT,default=json"`
	LogLevel       string `env:"LOG_LEVEL,default=info"`
}

func newConfig(ctx context.Context, lookuper envconfig.Lookuper) (*Config, error) {
	cfg := &Config{}
	err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target:   cfg,
		Lookuper: lookuper,
	})
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
