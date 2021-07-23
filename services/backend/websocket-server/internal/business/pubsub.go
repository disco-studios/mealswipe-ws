package business

import "context"

func DbPubsubWrite(channel string, message string) (err error) {
	return GetRedisClient().Publish(context.TODO(), channel, message).Err()
}
