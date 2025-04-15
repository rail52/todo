package cache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" env-required:"true"`
	Port     string `env:"REDIS_PORT" env-required:"true"`
	Password string `env:"REDIS_PASSWORD" env-required:"true"`
}

func New(cfg RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
	})
	if err := client.Ping().Err(); err != nil {
		panic(fmt.Errorf("failed to connect Redis:%w", err))
	}
	return client
}

type RedisRepository struct {
	Client *redis.Client
}

func NewRedisRepository(redisClient *redis.Client) *RedisRepository {
	return &RedisRepository{Client: redisClient}
}

func (r *RedisRepository) AddToBlackList(token string, exp int64) error {
	const op = "storage.cache.addToBlackList"
	ttl := exp - time.Now().Unix()
	if ttl < 0 {
		ttl = 60
	}
	err := r.Client.Set(token, "", time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}
	return nil
}

func (r *RedisRepository) IsBlackListed(token string) (bool, error) {
	const op = "storage.cache.isBlackListed"
	_, err := r.Client.Get(token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return true, nil
}
