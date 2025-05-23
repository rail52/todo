package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env              string        `yaml:"env" env-default:"local"`
	Address          string        `yaml:"address"`
	DBServiceAddress string        `yaml:"db-service_address" env-required:"true"`
	Timeout          time.Duration `yaml:"timeout"`
	IdleTimeout      time.Duration `yaml:"idle_timeout"`
	Kafka            `yaml:"kafka"`
}

type Kafka struct {
	Brokers []string `yaml:"kafka_brokers"`
	Topic   string   `yaml:"kafka_topic"`
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
