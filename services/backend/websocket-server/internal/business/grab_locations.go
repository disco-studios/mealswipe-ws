package business

import (
	"context"

	"github.com/go-redis/redis/v8"
)

func GrabLocations(lat float64, lng float64) (fsq_ids []string, distances []float64, err error) {
	// TODO Replace with GeoSearch when redis client supports it
	geoRad := redisClient.GeoRadius(context.TODO(), "locindex.restaurants", lng, lat, &redis.GeoRadiusQuery{
		Radius:   2,
		Unit:     "mi",
		WithDist: true,
	})

	if err = geoRad.Err(); err != nil {
		return
	}

	for _, loc := range geoRad.Val() {
		fsq_ids = append(fsq_ids, loc.Name)
		distances = append(distances, loc.Dist)
	}
	return
}
