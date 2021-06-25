package business

import "context"

func PubsubWrite(channel string, message string) (err error) {
	return redisClient.Publish(context.TODO(), channel, message).Err()
}
