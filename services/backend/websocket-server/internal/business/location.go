package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func DbLocationFromId(fsq_id string) (loc *mealswipepb.Location, err error) {
	hmget := GetRedisClient().HMGet(
		context.TODO(),
		BuildLocKey(fsq_id),
		"name",
		"photos",
		"latitude",
		"longitude",
		"chain_name",
		"address",
	)

	if err = hmget.Err(); err != nil {
		return
	}

	vals := hmget.Val()

	var photos []string
	var photo string
	json.Unmarshal([]byte(vals[1].(string)), &photos)
	if len(photos) > 0 {
		photo = photos[0]
	}

	loc = &mealswipepb.Location{
		Name:    fmt.Sprintf("%v", vals[0]),
		Photo:   fmt.Sprintf("%v", photo),
		Lat:     fmt.Sprintf("%v", vals[2]),
		Lng:     fmt.Sprintf("%v", vals[3]),
		Chain:   fmt.Sprintf("%v", vals[4]),
		Address: fmt.Sprintf("%v", vals[5]),
	}
	return
}

func DbLocationFromInd(sessionId string, index int64) (loc *mealswipepb.Location, err error) {
	get := GetRedisClient().LIndex(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), index)
	if err = get.Err(); err != nil {
		return
	}

	if len(get.Val()) == 0 {
		err = errors.New("couldn't find fsq id for index")
		return
	}

	return DbLocationFromId(get.Val())
}

func DbLocationIdsForLocation(lat float64, lng float64) (fsq_ids []string, distances []float64, err error) {
	// TODO Replace with GeoSearch when redis client supports it
	geoRad := GetRedisClient().GeoRadius(context.TODO(), BuildLocIndexKey("restaurants"), lng, lat, &redis.GeoRadiusQuery{
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
