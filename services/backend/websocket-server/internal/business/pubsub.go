package business

import (
	"context"

	"mealswipe.app/mealswipe/internal/msredis"
)

func DbPubsubWrite(channel string, message string) (err error) {
	return msredis.GetRedisClient().Publish(context.TODO(), channel, message).Err()
}
