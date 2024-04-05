package rediskey

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
)

func (e *emptyKey) Del(ctx context.Context, keys ...interface{}) (int, error) {
	return redigo.Int(e.do(ctx, "DEL", keys...))
}

func (e *EntityKey) Del(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "DEL", e.key))
}

func (e *EntityKey) Dump(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "DUMP", e.key))
}

func (e *EntityKey) Exists(ctx context.Context) (bool, error) {
	return redigo.Bool(e.do(ctx, "EXISTS", e.key))
}

func (e *EntityKey) Expire(ctx context.Context, seconds int) (int, error) {
	return redigo.Int(e.do(ctx, "EXPIRE", e.key, seconds))
}

func (e *EntityKey) ExpireAt(ctx context.Context, timeStamp uint64) (int, error) {
	return redigo.Int(e.do(ctx, "EXPIREAT", e.key, timeStamp))
}
func (e *EntityKey) PExpire(ctx context.Context, seconds int) (int, error) {
	return redigo.Int(e.do(ctx, "PEXPIRE", e.key, seconds))
}

func (e *EntityKey) PExpireAt(ctx context.Context, timeStamp uint64) (int, error) {
	return redigo.Int(e.do(ctx, "PEXPIREAT", e.key, timeStamp))
}

func (e *EntityKey) TTL(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "TTL", e.key))
}

func (e *EntityKey) PTTL(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "PTTL", e.key))
}

func (e *emptyKey) Keys(ctx context.Context, pattern string) ([]string, error) {
	return redigo.Strings(e.do(ctx, "KEYS", pattern))
}

func (e *EntityKey) Move(ctx context.Context, db uint8) (int, error) {
	return redigo.Int(e.do(ctx, "MOVE", e.key, db))
}

func (e *EntityKey) ObjectEncoding(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "OBJECT", "ENCODING", e.key))
}

func (e *EntityKey) ObjectRefCount(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "OBJECT", "ENCODING", e.key))
}

func (e *EntityKey) ObjectIdleTime(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "OBJECT", "IDLETIME", e.key))
}

func (e *EntityKey) Persist(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "PERSIST", e.key))
}

func (e *emptyKey) RandomKey(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "RANDOMKEY"))
}

func (e *EntityKey) Rename(ctx context.Context, newName string) (int, error) {
	n, err := redigo.Int(e.do(ctx, "RENAME", e.key, newName))
	if err == nil {
		e.key = newName
	}
	return n, err
}
func (e *EntityKey) RenameNX(ctx context.Context, newName string) (int, error) {
	n, err := redigo.Int(e.do(ctx, "RENAMENX", e.key, newName))
	if err == nil {
		e.key = newName
	}
	return n, err
}

func (e *EntityKey) Restore(ctx context.Context, ttl int, value string) (string, error) {
	return redigo.String(e.do(ctx, "RESTORE", e.key, ttl, value))
}

func (e *EntityKey) Type(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "TYPE", e.key))
}

func (e *emptyKey) Scan(ctx context.Context, cursor int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SCAN", cursor))
}

func (e *emptyKey) ScanMatch(ctx context.Context, cursor int, pattern string) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SCAN", cursor, "MATCH", pattern))
}

func (e *emptyKey) ScanCount(ctx context.Context, cursor int, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SCAN", cursor, "COUNT", count))
}

func (e *emptyKey) ScanMatchCount(ctx context.Context, cursor int, pattern string, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SCAN", cursor, "MATCH", pattern, "COUNT", count))
}
