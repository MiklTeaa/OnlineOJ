package rediskey

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
)

func (e *emptyKey) BLPop(ctx context.Context, timeout int, keys ...interface{}) ([]string, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, keys...)
	args = append(args, timeout)
	return redigo.Strings(e.do(ctx, "BLPOP", args...))
}

func (e *emptyKey) BRPop(ctx context.Context, timeout int, keys ...interface{}) ([]string, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, keys...)
	args = append(args, timeout)
	return redigo.Strings(e.do(ctx, "BRPOP", args...))
}

func (e *emptyKey) BRPopLPush(ctx context.Context, source, destination string, timeout int) (string, error) {
	return redigo.String(e.do(ctx, "BRPOPLPUSH", source, destination, timeout))
}

func (e *EntityKey) LIndex(ctx context.Context, index int) (string, error) {
	return redigo.String(e.do(ctx, "LINDEX", e.key, index))
}

func (e *EntityKey) LInsert(ctx context.Context, isAfter bool, pivot int, value interface{}) (int, error) {
	loc := "BEFORE"
	if isAfter {
		loc = "AFTER"
	}
	return redigo.Int(e.do(ctx, "LINSERT", e.key, loc, pivot, value))
}

func (e *EntityKey) LLen(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "LLEN", e.key))
}

func (e *EntityKey) LPop(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "LPOP", e.key))
}

func (e *EntityKey) LPush(ctx context.Context, values ...interface{}) (int, error) {
	args := append(make([]interface{}, 0, len(values)+1), e.key)
	args = append(args, values...)
	return redigo.Int(e.do(ctx, "LPUSH", args...))
}

func (e *EntityKey) LPushX(ctx context.Context, value interface{}) (int, error) {
	return redigo.Int(e.do(ctx, "LPUSHX", e.key, value))
}

func (e *EntityKey) LRange(ctx context.Context, start, end int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "LRANGE", e.key, start, end))
}

func (e *EntityKey) LRem(ctx context.Context, value interface{}, count int) (int, error) {
	return redigo.Int(e.do(ctx, "LREM", e.key, count, value))
}

func (e *EntityKey) LSet(ctx context.Context, index int, value interface{}) (string, error) {
	return redigo.String(e.do(ctx, "LSET", e.key, index, value))
}

func (e *EntityKey) LTrim(ctx context.Context, start, end int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "LTRIM", e.key, start, end))
}

func (e *EntityKey) RPop(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "RPOP", e.key))
}

func (e *EntityKey) RPush(ctx context.Context, values ...interface{}) (int, error) {
	args := append(make([]interface{}, 0, len(values)+1), e.key)
	args = append(args, values...)
	return redigo.Int(e.do(ctx, "RPUSH", args...))
}

func (e *EntityKey) RPushX(ctx context.Context, value interface{}) (int, error) {
	return redigo.Int(e.do(ctx, "RPUSHX", e.key, value))
}

func (e *emptyKey) RPopLPush(ctx context.Context, source, destination string) (string, error) {
	return redigo.String(e.do(ctx, "RPOPLPUSH", source, destination))
}
