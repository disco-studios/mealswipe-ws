package msredis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

var _rfedisClient *redis.ClusterClient

func LoadRedisClient() {
	_rfedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         []string{"mealswipe-cluster-redis-cluster:6379"},
		Password:      "qrDMS6jKt4",
		RouteRandomly: true,
	})
}

func GetRedisClient() *redis.ClusterClient {
	return _rfedisClient
}

func LoadRedisMockClient() redismock.ClientMock {
	var mock redismock.ClientMock
	_, mock = redismock.NewClientMock()
	return mock
}

func PubsubWrite(channel string, message string) (err error) {
	err = _rfedisClient.Publish(context.TODO(), channel, message).Err()
	err = fmt.Errorf("pubsub write: %v", err)
	return
}
