package business

import (
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

var redisClient *redis.Client

func LoadRedisClient() *redis.Client {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "host.docker.internal:6379", // TODO This may need to change when on kube
		Password: "",
		DB:       0, // use default DB
	})
	return redisClient
}

func LoadRedisMockClient() redismock.ClientMock {
	var mock redismock.ClientMock
	redisClient, mock = redismock.NewClientMock()
	return mock
}
