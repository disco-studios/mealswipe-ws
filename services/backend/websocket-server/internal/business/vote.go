package business

import "context"

func DbVote(userId string, index int64, state bool) (err error) {
	voteBit := 0
	if state {
		voteBit = 1
	}

	return redisClient.SetBit(context.TODO(), "user."+userId+".votes", index, voteBit).Err()
}
