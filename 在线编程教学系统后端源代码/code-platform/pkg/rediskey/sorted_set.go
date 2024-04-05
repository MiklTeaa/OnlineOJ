package rediskey

import (
	"context"

	redigo "github.com/gomodule/redigo/redis"
)

func (e *EntityKey) ZAdd(ctx context.Context, scoreMembers ...interface{}) (int, error) {
	args := append(make([]interface{}, 0, len(scoreMembers)+1), e.key)
	args = append(args, scoreMembers...)
	return redigo.Int(e.do(ctx, "ZADD", args...))
}

func (e *EntityKey) ZCard(ctx context.Context) (int, error) {
	return redigo.Int(e.do(ctx, "ZCARD", e.key))
}

func (e *EntityKey) ZCount(ctx context.Context, min, max int) (int, error) {
	return redigo.Int(e.do(ctx, "ZCOUNT", e.key, min, max))
}

func (e *EntityKey) ZIncrBy(ctx context.Context, increment int, member string) (string, error) {
	return redigo.String(e.do(ctx, "ZINCRBY", e.key, increment, member))
}

func (e *EntityKey) ZRange(ctx context.Context, start, stop int, withScores bool) ([]string, error) {
	if withScores {
		return redigo.Strings(e.do(ctx, "ZRANGE", e.key, start, stop, "WITHSCORES"))
	}
	return redigo.Strings(e.do(ctx, "ZRANGE", e.key, start, stop))
}

func (e *EntityKey) ZRangeByScore(ctx context.Context, min, max int, withScores bool, limitParameters ...int) ([]string, error) {
	if withScores {
		if len(limitParameters) == 2 {
			offset, count := limitParameters[0], limitParameters[1]
			return redigo.Strings(e.do(ctx, "ZRANGEBYSCORE", e.key, min, max, "WITHSCORE", "LIMIT", offset, count))
		}
		return redigo.Strings(e.do(ctx, "ZRANGEBYSCORE", e.key, min, max, "WITHSCORE"))
	}

	if len(limitParameters) == 2 {
		offset, count := limitParameters[0], limitParameters[1]
		return redigo.Strings(e.do(ctx, "ZRANGEBYSCORE", e.key, min, max, "LIMIT", offset, count))
	}
	return redigo.Strings(e.do(ctx, "ZRANGEBYSCORE", e.key, min, max))
}

func (e *EntityKey) ZRank(ctx context.Context, member string) (int, error) {
	return redigo.Int(e.do(ctx, "ZRANK", e.key, member))
}

func (e *EntityKey) ZRem(ctx context.Context, members ...interface{}) (int, error) {
	args := append(make(redigo.Args, 0, len(members)+1), e.key)
	args = append(args, members...)
	return redigo.Int(e.do(ctx, "ZREM", args...))
}

func (e *EntityKey) ZRemRangeByRank(ctx context.Context, start, stop int) (int, error) {
	return redigo.Int(e.do(ctx, "ZREMRANGEBYRANK", e.key, start, stop))
}

func (e *EntityKey) ZRemRangeByScore(ctx context.Context, min, max int) (int, error) {
	return redigo.Int(e.do(ctx, "ZREMRANGEBYSCORE", e.key, min, max))
}

func (e *EntityKey) ZRevRange(ctx context.Context, start, stop int, withScores bool) ([]string, error) {
	if withScores {
		return redigo.Strings(e.do(ctx, "ZREVRANGE", e.key, start, stop, "WITHSCORES"))
	}
	return redigo.Strings(e.do(ctx, "ZREVRANGE", e.key, start, stop))
}

func (e *EntityKey) ZRevRangeByScore(ctx context.Context, max, min int, withScores bool, limitParameters ...int) ([]string, error) {
	if withScores {
		if len(limitParameters) == 2 {
			offset, count := limitParameters[0], limitParameters[1]
			return redigo.Strings(e.do(ctx, "ZREVRANGEBYSCORE", e.key, max, min, "WITHSCORE", "LIMIT", offset, count))
		}
		return redigo.Strings(e.do(ctx, "ZREVRANGEBYSCORE", e.key, max, min, "WITHSCORE"))
	}

	if len(limitParameters) == 2 {
		offset, count := limitParameters[0], limitParameters[1]
		return redigo.Strings(e.do(ctx, "ZREVRANGEBYSCORE", e.key, max, min, "LIMIT", offset, count))
	}
	return redigo.Strings(e.do(ctx, "ZREVRANGEBYSCORE", e.key, max, min))
}

func (e *EntityKey) ZRevRank(ctx context.Context, member string) (int, error) {
	return redigo.Int(e.do(ctx, "ZREVRANK", e.key, member))
}

func (e *EntityKey) ZScore(ctx context.Context, member string) (int, error) {
	return redigo.Int(e.do(ctx, "ZSCORE", e.key, member))
}

type AGGREGATE string

const (
	Default AGGREGATE = ""
	Sum     AGGREGATE = "SUM"
	Min     AGGREGATE = "MIN"
	Max     AGGREGATE = "MAX"
)

func dealWithArgsWhenUseZSetStore(destination string, numKeys int, keys []interface{}, weights []interface{}, aggregate AGGREGATE) []interface{} {
	args := make([]interface{}, 0, len(keys)+len(weights)+3)
	args = append(args, destination, numKeys)
	args = append(args, keys...)
	if len(weights) != 0 {
		args = append(args, "WEIGHT")
		args = append(args, weights...)
	}

	if aggregate == Default {
		aggregate = Sum
	}
	args = append(args, string(aggregate))
	return args
}

func (e *emptyKey) ZUnionStore(ctx context.Context, destination string, numKeys int, keys []interface{}, weights []interface{}, aggregate AGGREGATE) (int, error) {
	return redigo.Int(e.do(ctx, "ZUNIONSTORE", dealWithArgsWhenUseZSetStore(destination, numKeys, keys, weights, aggregate)...))
}

func (e *emptyKey) ZInterStore(ctx context.Context, destination string, numKeys int, keys []interface{}, weights []interface{}, aggregate AGGREGATE) (int, error) {
	return redigo.Int(e.do(ctx, "ZINTERSTORE", dealWithArgsWhenUseZSetStore(destination, numKeys, keys, weights, aggregate)...))

}

func (e *EntityKey) ZScan(ctx context.Context, cursor int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "ZSCAN", e.key, cursor))
}

func (e *EntityKey) ZScanMatch(ctx context.Context, cursor int, pattern string) ([]string, error) {
	return redigo.Strings(e.do(ctx, "ZSCAN", e.key, cursor, "MATCH", pattern))
}

func (e *EntityKey) ZScanCount(ctx context.Context, cursor int, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "ZSCAN", e.key, cursor, "COUNT", count))
}

func (e *EntityKey) ZScanMatchCount(ctx context.Context, cursor int, pattern string, count int) ([]string, error) {
	return redigo.Strings(e.do(ctx, "ZSCAN", e.key, cursor, "MATCH", pattern, "COUNT", count))
}
