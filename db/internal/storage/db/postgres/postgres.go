package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	UserName string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`
}

func NewStorage(PostgresCfg Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s  user=%s password=%s dbname=%s sslmode=disable",
		PostgresCfg.Host,
		PostgresCfg.Port,
		PostgresCfg.UserName,
		PostgresCfg.Password,
		PostgresCfg.DBName,
	)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("can't create pgxpool.Pool (maybe DB(docker-container) isn't up)")
	}
	if pool.Ping(context.Background()) != nil {
		return nil, fmt.Errorf("can't connect to DB (maybe DB(docker-container) isn't up)")
	}
	return pool, nil
}
