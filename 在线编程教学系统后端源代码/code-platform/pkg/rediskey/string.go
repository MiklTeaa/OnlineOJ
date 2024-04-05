package rediskey

import (
	"context"
	"errors"

	redigo "github.com/gomodule/redigo/redis"
)

func (e *EntityKey) Get(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "GET", e.key))
}

func (e *EntityKey) GetBytes(ctx context.Context) ([]byte, error) {
	return redigo.Bytes(e.do(ctx, "GET", e.key))
}

func (e *EntityKey) GetInt(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "GET", e.key))
}

func (e *EntityKey) GetUint64(ctx context.Context) (uint64, error) {
	return redigo.Uint64(e.do(ctx, "GET", e.key))
}

func (e *EntityKey) Set(ctx context.Context, value interface{}) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value))
}

func (e *EntityKey) SetEX(ctx context.Context, value interface{}, seconds int) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "EX", seconds))
}

func (e *EntityKey) SetPX(ctx context.Context, value interface{}, milliseconds int) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "PX", milliseconds))
}

func (e *EntityKey) SetNX(ctx context.Context, value interface{}) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "NX"))
}

func (e *EntityKey) SetXX(ctx context.Context, value interface{}) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "XX"))
}

func (e *EntityKey) SetEXNX(ctx context.Context, value interface{}, seconds int) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "EX", seconds, "NX"))
}

func (e *EntityKey) SetPXNX(ctx context.Context, value interface{}, milliseconds int) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "EX", milliseconds, "PX"))
}

func (e *EntityKey) SetEXXX(ctx context.Context, value interface{}, seconds int) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "EX", seconds, "XX"))
}

func (e *EntityKey) SetPXXX(ctx context.Context, value interface{}, milliseconds int) (string, error) {
	return redigo.String(e.do(ctx, "SET", e.key, value, "PX", milliseconds, "XX"))
}

func (e *EntityKey) StrLen(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "STRLEN", e.key))
}

func (e *EntityKey) Append(ctx context.Context, extra string) (int, error) {
	return redigo.Int(e.do(ctx, "APPEND", e.key, extra))
}

func (e *EntityKey) BitCount(ctx context.Context, marks ...int) (int, error) {
	if len(marks) != 2 && len(marks) != 0 {
		return 0, errors.New("the parameters of bitcount should be 0 or 2")
	}

	if len(marks) == 0 {
		return redigo.Int(e.do(ctx, "BITCOUNT", e.key))
	}
	return redigo.Int(e.do(ctx, "BITCOUNT", e.key, marks[0], marks[1]))
}

type bitOP string

const (
	NOT bitOP = "NOT"
	XOR bitOP = "XOR"
	OR  bitOP = "OR"
	AND bitOP = "AND"
)

func (e *emptyKey) BitOP(ctx context.Context, op bitOP, destinationKey string, keys ...interface{}) (int, error) {
	args := append(make([]interface{}, 0, len(keys)+2), string(op), destinationKey)
	args = append(args, keys...)
	return redigo.Int(e.do(ctx, "BITOP", args...))
}

func (e *EntityKey) Decr(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "DECR", e.key))
}

func (e *EntityKey) DecrBy(ctx context.Context, step int) (int, error) {
	return redigo.Int(e.do(ctx, "DECRBY", e.key, step))
}

func (e *EntityKey) GetBit(ctx context.Context, offset int) (int, error) {
	return redigo.Int(e.do(ctx, "GETBIT", e.key, offset))
}

func (e *EntityKey) GetRange(ctx context.Context, start, end int) (string, error) {
	return redigo.String(e.do(ctx, "GETRANGE", e.key, start, end))
}

func (e *EntityKey) GetSet(ctx context.Context, value string) (string, error) {
	return redigo.String(e.do(ctx, "GETSET", e.key, value))
}

func (e *EntityKey) Incr(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "INCR", e.key))
}

func (e *EntityKey) IncrBy(ctx context.Context, step int) (int, error) {
	return redigo.Int(e.do(ctx, "INCRBY", e.key, step))
}

func (e *EntityKey) IncrByFloat(ctx context.Context, step float64) (float64, error) {
	return redigo.Float64(e.do(ctx, "INCRBYFLOAT", e.key, step))
}

func (e *emptyKey) MGet(ctx context.Context, keys ...interface{}) ([][]byte, error) {
	return redigo.ByteSlices(e.do(ctx, "MGET", keys...))
}

func (e *emptyKey) MSet(ctx context.Context, keyValues ...interface{}) (string, error) {
	return redigo.String(e.do(ctx, "MSET", keyValues...))
}

func (e *emptyKey) MSetNX(ctx context.Context, keyValues ...interface{}) (int, error) {
	return redigo.Int(e.do(ctx, "MSETNX", keyValues...))
}

func (e *EntityKey) SetBit(ctx context.Context, offset int, value int) (int, error) {
	return redigo.Int(e.do(ctx, "SETBIT", e.key, offset, value))
}

func (e *EntityKey) SetRange(ctx context.Context, offset int, value string) (int, error) {
	return redigo.Int(e.do(ctx, "SETRANGE", e.key, offset, value))
}
