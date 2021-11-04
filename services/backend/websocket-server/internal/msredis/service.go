package msredis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	"go.elastic.co/apm"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8"
)

const TRACE_REDIS bool = true
const CLUSTER bool = false

var _redisClusterClient *redis.ClusterClient
var _redisClient *redis.Client

func LoadRedisClient() {
	if CLUSTER {
		_redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:         []string{"mealswipe-cluster-redis-cluster:6379"},
			Password:      "qrDMS6jKt4",
			RouteRandomly: true,
		})
		if TRACE_REDIS {
			_redisClusterClient.AddHook(apmgoredis.NewHook())
		}
	} else {
		_redisClient = redis.NewClient(&redis.Options{
			Addr: "ms-redis-service:6379",
		})
	}
}

func LoadRedisMockClient() redismock.ClientMock {
	var mock redismock.ClientMock
	_, mock = redismock.NewClientMock()
	return mock
}

func PubsubWrite(ctx context.Context, channel string, message string) (err error) {
	span, ctx := apm.StartSpan(ctx, "PubsubWrite", "msredis")
	defer span.End()

	if CLUSTER {
		err = _redisClusterClient.Publish(ctx, channel, message).Err()
	} else {
		err = _redisClient.Publish(ctx, channel, message).Err()
	}

	if err != nil {
		err = fmt.Errorf("pubsub write: %v", err)
	}
	return
}

func SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if CLUSTER {
		return _redisClusterClient.SetEX(ctx, key, value, expiration)
	} else {
		return _redisClient.SetEX(ctx, key, value, expiration)
	}
}

func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	if CLUSTER {
		return _redisClusterClient.SetNX(ctx, key, value, expiration)
	} else {
		return _redisClient.SetNX(ctx, key, value, expiration)
	}
}

func Pipeline() redis.Pipeliner {
	if CLUSTER {
		return _redisClusterClient.Pipeline()
	} else {
		return _redisClient.Pipeline()
	}
}

func Get(ctx context.Context, key string) *redis.StringCmd {
	if CLUSTER {
		return _redisClusterClient.Get(ctx, key)
	} else {
		return _redisClient.Get(ctx, key)
	}
}

func Del(ctx context.Context, key string) *redis.IntCmd {
	if CLUSTER {
		return _redisClusterClient.Del(ctx, key)
	} else {
		return _redisClient.Del(ctx, key)
	}
}

func Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	if CLUSTER {
		return _redisClusterClient.Scan(ctx, cursor, match, count)
	} else {
		return _redisClient.Scan(ctx, cursor, match, count)
	}
}

func Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if CLUSTER {
		return _redisClusterClient.Subscribe(ctx, channels...)
	} else {
		return _redisClient.Subscribe(ctx, channels...)
	}
}

func HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	if CLUSTER {
		return _redisClusterClient.HGetAll(ctx, key)
	} else {
		return _redisClient.HGetAll(ctx, key)
	}
}

func SetBit(ctx context.Context, key string, offset int64, value int) *redis.IntCmd {
	if CLUSTER {
		return _redisClusterClient.SetBit(ctx, key, offset, value)
	} else {
		return _redisClient.SetBit(ctx, key, offset, value)
	}
}

func HGet(ctx context.Context, key string, field string) *redis.StringCmd {
	if CLUSTER {
		return _redisClusterClient.HGet(ctx, key, field)
	} else {
		return _redisClient.HGet(ctx, key, field)
	}
}

func HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if CLUSTER {
		return _redisClusterClient.HSet(ctx, key, values...)
	} else {
		return _redisClient.HSet(ctx, key, values...)
	}
}

func SIsMember(ctx context.Context, key string, member interface{}) *redis.BoolCmd {
	if CLUSTER {
		return _redisClusterClient.SIsMember(ctx, key, member)
	} else {
		return _redisClient.SIsMember(ctx, key, member)
	}
}
