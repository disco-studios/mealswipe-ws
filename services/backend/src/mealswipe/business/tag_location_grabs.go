package business

import "context"

func TagLocationGrabs(venues []*FS_Venue) (err error) {
	pipe := redisClient.Pipeline()

	for _, venue := range venues {
		key := "venue." + venue.ID + ".grabs"
		pipe.Incr(context.TODO(), key)
	}

	_, err = pipe.Exec(context.TODO())
	return
}
