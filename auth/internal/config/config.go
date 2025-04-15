package config

import (
	"auth/internal/storage/cache"
	postgres "auth/internal/storage/db"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	postgres.Config
	cache.RedisConfig

	Address string `env:"ADDRESS" env-required:"true"`

	MentorServiceAddress string `env:"MENTOR_SERVICE_ADDRESS" env-required:"true"`

	Env string `env:"ENV" env-required:"true"`

	AccessTokenTTL  int64 `env:"ACCES_TOKEN_TTL" env-default:"1800"`
	RefreshTokenTTL int64 `env:"RESRESH_TOKEN_TTL" env-default:"604800"`

	PrivateKeyPath string `env:"JWT_PRIVATE_KEY_PATH"`
	PublicKeyPath  string `env:"JWT_PUBLIC_KEY_PATH"`

	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

func LoadConfig() *Config {
	cfg := &Config{}
	err := cleanenv.ReadConfig("./.env", cfg)
	if err != nil {
		log.Fatalf("error reading config: %s", err.Error())
	}
	return cfg
}
