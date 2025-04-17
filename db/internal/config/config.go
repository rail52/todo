package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"db/internal/storage/db/postgres"
	"db/internal/storage/cache/redis"
	"time"
)

type Config struct {
	Postgres postgres.Config
	Redis redis.Config
	Env                   string        `yaml:"env" env-default:"local"`
	Address               string        `yaml:"address"`
	Timeout               time.Duration `yaml:"timeout"`
	IdleTimeout           time.Duration `yaml:"idle_timeout"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("invalid CONFIG_PATH")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("%v Is Not Exist", configPath)
	}
	var config Config

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		log.Fatalf("can't read config: %v", err)
	}

	return &config
}
