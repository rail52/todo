package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Address            string        `yaml:"address" env-default:"0.0.0.0:8080"`
	AuthServiceAddress string        `yaml:"auth_service_address", env-required:"true"`
	Env                string        `yaml:"env" env-default:"local"`
	Timeout            time.Duration `yaml:"timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout"`
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
