// Package db provides the database clients for the application.
package db

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	Client *redis.Client
}

var ctx = context.Background()

// NewRedisService sets up a new Redis client and returns it.
func NewRedisService() (*RedisService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &RedisService{
		Client: rdb,
	}, nil
}

// SetShortCodeWithJWT sets a sha256 truncated hash as the key and the original JWT as the value.
func (rs *RedisService) SetShortCodeWithJWT(shortCode string, jwt string) error {
	err := rs.Client.Set(ctx, shortCode, jwt, 12*60*60)
	if err.Err() != nil {
		return err.Err()
	}

	return nil
}

// GetJWTFromShortCode retrieves the JWT from the Redis store using the short code.
func (rs *RedisService) GetJWTFromShortCode(shortCode string) (string, error) {
	val, err := rs.Client.Get(ctx, shortCode).Result()
	if err != nil {
		return val, err
	}

	return val, nil
}
