package msredis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
	apmgoredis "go.elastic.co/apm/module/apmgoredisv8"
)

const TRACE_REDIS bool = true

var _rfedisClient *redis.ClusterClient

func LoadRedisClient() {
	_rfedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         []string{"mealswipe-cluster-redis-cluster:6379"},
		Password:      "qrDMS6jKt4",
		RouteRandomly: true,
	})
	if TRACE_REDIS {
		_rfedisClient.AddHook(apmgoredis.NewHook())
	}
}

func GetRedisClient() *redis.ClusterClient {
	return _rfedisClient
}

func LoadRedisMockClient() redismock.ClientMock {
	var mock redismock.ClientMock
	_, mock = redismock.NewClientMock()
	return mock
}

func PubsubWrite(ctx context.Context, channel string, message string) (err error) {
	err = _rfedisClient.Publish(ctx, channel, message).Err()
	if err != nil {
		err = fmt.Errorf("pubsub write: %v", err)
	}
	return
}
