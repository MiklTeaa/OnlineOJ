package rediskey

import (
	"context"
	"fmt"
	"sync"

	redigo "github.com/gomodule/redigo/redis"
)

type EntityKey struct {
	pool *redigo.Pool
	key  string
}

func NewkeyFormat(format string, args ...interface{}) *EntityKey {
	return &EntityKey{key: fmt.Sprintf(format, args...)}
}

func Newkey(key string) *EntityKey {
	return &EntityKey{key: key}
}

func (e *EntityKey) Replace(format string, args ...interface{}) *EntityKey {
	if len(args) == 0 {
		e.key = format
	} else {
		e.key = fmt.Sprintf(format, args...)
	}
	return e
}

func (e *EntityKey) Clear() {
	e.key = ""
	e.pool = nil
}

func (e *EntityKey) String() string {
	return e.key
}

func (e *EntityKey) do(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := e.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return redigo.DoContext(conn, ctx, commandName, args...)
}

func (e *EntityKey) Pool(pool *redigo.Pool) *EntityKey {
	e.pool = pool
	return e
}

// use for reuse emptyKey object
var emptyKeyPool = &sync.Pool{
	New: func() interface{} {
		return &emptyKey{}
	},
}

type emptyKey struct {
	pool *redigo.Pool
}

func NewEmptyKey() *emptyKey {
	return emptyKeyPool.Get().(*emptyKey)
}

func (e *emptyKey) do(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := e.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return redigo.DoContext(conn, ctx, commandName, args...)
}

func (e *emptyKey) Pool(pool *redigo.Pool) *emptyKey {
	e.pool = pool
	return e
}

func (e *emptyKey) Release() {
	e.pool = nil
	emptyKeyPool.Put(e)
}
