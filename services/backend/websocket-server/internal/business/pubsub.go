package business

import "context"

func DbPubsubWrite(channel string, message string) (err error) {
	return redisClient.Publish(context.TODO(), channel, message).Err()
}
