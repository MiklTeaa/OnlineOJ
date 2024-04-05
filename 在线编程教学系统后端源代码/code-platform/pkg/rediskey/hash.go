package rediskey

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
)

func (e *EntityKey) HDel(ctx context.Context, fields ...interface{}) (int, error) {
	args := append(make([]interface{}, 0, len(fields)+1), e.key)
	args = append(args, fields...)
	return redigo.Int(e.do(ctx, "HDEL", args...))
}

func (e *EntityKey) HExists(ctx context.Context, field string) (int, error) {
	return redigo.Int(e.do(ctx, "HEXISTS", e.key, field))
}

func (e *EntityKey) HGet(ctx context.Context, field string) (string, error) {
	return redigo.String(e.do(ctx, "HGET", e.key, field))
}

func (e *EntityKey) HGetAll(ctx context.Context) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HGETALL", e.key))
}

func (e *EntityKey) HIncrBy(ctx context.Context, field string, step int) (int, error) {
	return redigo.Int(e.do(ctx, "HINCRBY", e.key, field, step))
}

func (e *EntityKey) HIncrByFloat(ctx context.Context, field string, step float64) (float64, error) {
	return redigo.Float64(e.do(ctx, "HINCRBYFLOAT", e.key, field, step))
}

func (e *EntityKey) HKeys(ctx context.Context) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HKEYS", e.key))
}

func (e *EntityKey) HVals(ctx context.Context) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HVALS", e.key))
}

func (e *EntityKey) HLen(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "HLEN", e.key))
}

func (e *EntityKey) HMGet(ctx context.Context, fields ...interface{}) ([]string, error) {
	args := append(make([]interface{}, 0, len(fields)+1), e.key)
	args = append(args, fields...)
	return redigo.Strings(e.do(ctx, "HMGET", args...))
}

func (e *EntityKey) HSet(ctx context.Context, fieldValues ...interface{}) (int, error) {
	args := append(make([]interface{}, 0, len(fieldValues)+1), e.key)
	args = append(args, fieldValues...)
	return redigo.Int(e.do(ctx, "HSET", args...))
}

func (e *EntityKey) HSetNX(ctx context.Context, field string, value interface{}) (int, error) {
	return redigo.Int(e.do(ctx, "HSETNX", e.key, field, value))
}

func (e *EntityKey) HScan(ctx context.Context, cursor int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HSCAN", e.key, cursor))
}

func (e *EntityKey) HScanMatch(ctx context.Context, cursor int, pattern string) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HSCAN", e.key, cursor, "MATCH", pattern))
}

func (e *EntityKey) HScanCount(ctx context.Context, cursor int, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HSCAN", e.key, cursor, "COUNT", count))
}

func (e *EntityKey) HScanMatchCount(ctx context.Context, cursor int, pattern string, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "HSCAN", e.key, cursor, "MATCH", pattern, "COUNT", count))
}
