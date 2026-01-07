package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServiceName        string `envconfig:"SERVICE_NAME" default:"identity-service"`
	Port               string `envconfig:"PORT" default:"4001"`
	InternalAPIKey     string `envconfig:"INTERNAL_API_KEY" required:"true"`
	DatabaseURL        string `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret          string `envconfig:"JWT_SECRET" required:"true"`
	GoogleClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	JWTExpiresIn       string `envconfig:"JWT_EXPIRES_IN" default:"15m"`
    RefreshTokenExpiry string `envconfig:"REFRESH_TOKEN_EXPIRY" default:"168h"` // 7 days
}

func Load() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
