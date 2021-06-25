package business

import "context"

func DbGetSessionIdFromCode(code string) (sessionId string, err error) {
	key := "code." + code
	result := redisClient.Get(context.TODO(), key)
	return result.Val(), result.Err()
}
