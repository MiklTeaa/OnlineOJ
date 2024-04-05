package storage

import (
	"context"

	"code-platform/config"

	redigo "github.com/gomodule/redigo/redis"
)

type RedisClient struct {
	Pool *redigo.Pool
}

// MustInitRedisClient return a redisClient or panic
func MustInitRedisClient() RedisClient {
	host := config.Redis.GetString("host")
	port := config.Redis.GetString("port")
	password := config.Redis.GetString("password")

	address := host + ":" + port
	pool := &redigo.Pool{
		Wait: true,
		DialContext: func(ctx context.Context) (redigo.Conn, error) {
			conn, err := redigo.DialContext(ctx, "tcp", address,
				redigo.DialPassword(password),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		MaxIdle:         1000,
		MaxActive:       2000,
		IdleTimeout:     30,
		MaxConnLifetime: 7190,
	}

	return RedisClient{Pool: pool}
}

func MustInitRedisLRUClient() RedisClient {
	host := config.RedisLRU.GetString("host")
	port := config.RedisLRU.GetString("port")
	password := config.RedisLRU.GetString("password")

	address := host + ":" + port
	pool := &redigo.Pool{
		DialContext: func(ctx context.Context) (redigo.Conn, error) {
			conn, err := redigo.DialContext(ctx, "tcp", address,
				redigo.DialPassword(password),
			)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		MaxIdle:         1000,
		MaxActive:       2000,
		IdleTimeout:     30,
		MaxConnLifetime: 7190,
	}

	return RedisClient{Pool: pool}
}

func (r *RedisClient) Close() error {
	return r.Pool.Close()
}
