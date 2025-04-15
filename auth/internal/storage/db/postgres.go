package db

import (
	"auth/internal/domain/model"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var ErrUserNotFound = errors.New("user not found")

type Config struct {
	UserName string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     string `env:"POSTGRES_PORT" env-required:"true"`
	DBName   string `env:"POSTGRES_DB" env-required:"true"`
}

type Storage struct {
	db *sqlx.DB
}

func NewStorage(cfg Config) (*Storage, error) {
	dsn := fmt.Sprintf("host=%s port=%s  user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.UserName, cfg.Password, cfg.DBName)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error with connect to database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) CreateUser(u *model.User) error {
	const op = "stoage.db.CreateUser"
	query := `INSERT INTO users (email, password)
			  VALUES ($1, $2)
			  RETURNING id`
	var newID int64
	err := s.db.QueryRow(query, u.Email, u.Password).Scan(&newID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	u.ID = newID
	return nil
}

func (s *Storage) GetByEmail(email string) (*model.User, error) {
	const op = "storage.db.GetByEmail"
	query := `SELECT id, email, password FROM users WHERE email=$1`
	user := &model.User{}
	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}
	return user, nil
}
