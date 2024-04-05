package rediskey

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
)

func (e *EntityKey) SAdd(ctx context.Context, members ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, e.key)
	args = append(args, members...)
	return redigo.Int(e.do(ctx, "SADD", args...))
}

func (e *EntityKey) SCard(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "SCARD", e.key))
}

func (e *emptyKey) SDiff(ctx context.Context, keys ...interface{}) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SDIFF", keys...))
}

func (e *emptyKey) SDiffStore(ctx context.Context, destination string, keys ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)
	return redigo.Int(e.do(ctx, "SDIFFSTORE", args...))
}
func (e *emptyKey) SInter(ctx context.Context, keys ...interface{}) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SINTER", keys...))
}

func (e *emptyKey) SInterStore(ctx context.Context, destination string, keys ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)
	return redigo.Int(e.do(ctx, "SINTERSTORE", args...))
}

func (e *emptyKey) SUnion(ctx context.Context, keys ...interface{}) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SUNION", keys...))
}

func (e *emptyKey) SUnionStore(ctx context.Context, destination string, keys ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, destination)
	args = append(args, keys...)
	return redigo.Int(e.do(ctx, "SUNIONSTORE", args...))
}

func (e *EntityKey) SIsMember(ctx context.Context, member string) (int, error) {
	return redigo.Int(e.do(ctx, "SISMEMBER", e.key, member))
}

func (e *EntityKey) SMembers(ctx context.Context) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SMEMBERS", e.key))
}

func (e *emptyKey) SMove(ctx context.Context, source, destination, member string) (int, error) {
	return redigo.Int(e.do(ctx, "SMOVE", source, destination, member))
}

func (e *EntityKey) SPop(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "SPOP", e.key))
}

func (e *EntityKey) SRandomMember(ctx context.Context) (string, error) {
	return redigo.String(e.do(ctx, "SRANDOMMEMBER", e.key))
}

func (e *EntityKey) SRandomMemberCount(ctx context.Context, count int) (string, error) {
	return redigo.String(e.do(ctx, "SRANDOMMEMBER", e.key, count))
}

func (e *EntityKey) SRem(ctx context.Context, members ...interface{}) (int, error) {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, e.key)
	args = append(args, members...)
	return redigo.Int(e.do(ctx, "SREM", args...))
}

func (e *EntityKey) SScan(ctx context.Context, cursor int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SSCAN", e.key, cursor))
}

func (e *EntityKey) SScanMatch(ctx context.Context, cursor int, pattern string) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SSCAN", e.key, cursor, "MATCH", pattern))
}

func (e *EntityKey) SScanCount(ctx context.Context, cursor int, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SSCAN", e.key, cursor, "COUNT", count))
}

func (e *EntityKey) SScanMatchCount(ctx context.Context, cursor int, pattern string, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "SSCAN", e.key, cursor, "MATCH", pattern, "COUNT", count))
}
