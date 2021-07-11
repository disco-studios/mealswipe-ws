package business

import (
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

var _rfedisClient *redis.ClusterClient

func LoadRedisClient() {
	_rfedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:         []string{"mealswipe-cluster-redis-cluster:6379"},
		Password:      "gvSKYRLzac",
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
