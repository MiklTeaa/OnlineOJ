package testx

import (
	"context"
	"fmt"

	"code-platform/storage"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

const redisTestDBNumber = 4

func mustInitRedisTestingClient(db int, redisConfig *viper.Viper) storage.RedisClient {
	host := redisConfig.GetString("host")
	port := redisConfig.GetString("port")
	password := redisConfig.GetString("password")

	address := fmt.Sprintf("%s:%s", host, port)
	pool := &redigo.Pool{
		Wait: true,
		DialContext: func(ctx context.Context) (redigo.Conn, error) {
			conn, err := redigo.Dial("tcp", address,
				redigo.DialDatabase(db),
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

	return storage.RedisClient{Pool: pool}
}

func MustFlushDB(ctx context.Context, pool *redigo.Pool) {
	conn, err := pool.GetContext(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	if _, err := redigo.DoContext(conn, ctx, "SELECT", redisTestDBNumber); err != nil {
		panic(err)
	}

	if _, err := redigo.DoContext(conn, ctx, "FLUSHDB"); err != nil {
		panic(err)
	}
}
