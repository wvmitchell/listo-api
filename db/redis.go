// Package db provides the database clients for the application.
package db

import (
	"context"
	"crypto/tls"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
	"time"
)

// RedisService is a struct that contains the Redis client.
type RedisService struct {
	Client *redis.Client
}

var ctx = context.Background()

// NewRedisService sets up a new Redis client and returns it.
func NewRedisService() (*RedisService, error) {
	addr := os.Getenv("REDIS_URL")
	password := os.Getenv("REDIS_PASSWORD")
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	tlsConfig := &tls.Config{}

	if addr == "localhost:6379" {
		tlsConfig = nil // Disable TLS for local development
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  password,
		DB:        db,
		TLSConfig: tlsConfig,
	})

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		println("Error connecting to Redis: " + err.Error())
		return nil, err
	}

	return &RedisService{
		Client: rdb,
	}, nil
}

// SetShortCodeWithJWT sets a sha256 truncated hash as the key and the original JWT as the value.
func (rs *RedisService) SetShortCodeWithJWT(shortCode string, jwt string) error {
	err := rs.Client.Set(ctx, shortCode, jwt, time.Duration(12*time.Hour))
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
