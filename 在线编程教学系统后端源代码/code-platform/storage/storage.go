package storage

import (
	"context"
	"database/sql"

	"code-platform/log"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
)

// Storage
type Storage struct {
	RDB      RDBClient
	Redis    RedisClient
	RedisLRU RedisClient
	Minio    MinioClient
}

// NewStorage return a storage
func NewStorage() *Storage {
	return &Storage{
		RDB:      MustInitMysqlClient(),
		Redis:    MustInitRedisClient(),
		RedisLRU: MustInitRedisLRUClient(),
		Minio:    MustInitMinioClient(),
	}
}

func (s *Storage) Pool() *redigo.Pool {
	return s.Redis.Pool
}

func (s *Storage) LRUPool() *redigo.Pool {
	return s.RedisLRU.Pool
}

func (s *Storage) RDBTransaction(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	db := s.RDB.(*sqlx.DB)
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *Storage) Close() {
	DB := s.RDB.(*sqlx.DB)
	if err := DB.Close(); err != nil {
		log.Error(err, "close db failed")
	}
	if err := s.Redis.Close(); err != nil {
		log.Error(err, "close redis client failed")
	}

	if err := s.RedisLRU.Close(); err != nil {
		log.Error(err, "close redis_lru client failed")
	}
}
