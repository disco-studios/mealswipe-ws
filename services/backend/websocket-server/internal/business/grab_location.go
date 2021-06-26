package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func DbGrabLocation(fsq_id string) (loc *mealswipepb.Location, err error) {
	hmget := redisClient.HMGet(
		context.TODO(),
		"loc."+fsq_id,
		"name",
		"photos",
		"latitude",
		"longitude",
		"chain_name",
		"address",
	)

	if err = hmget.Err(); err != nil {
		log.Println("No loc", "loc."+fsq_id)
		return
	}

	vals := hmget.Val()

	var photos []string
	var photo string
	log.Println(vals[1].(string))
	json.Unmarshal([]byte(vals[1].(string)), &photos)
	log.Println(photos)
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

func DbGrabLocationFromInd(sessionId string, index int64) (loc *mealswipepb.Location, err error) {
	get := redisClient.LIndex(context.TODO(), "session."+sessionId+".locations", index)
	if err = get.Err(); err != nil {
		log.Println("Couldn't get loc for ind", index, "sessions."+sessionId+".locations")
		return
	}

	if len(get.Val()) == 0 {
		err = errors.New("couldn't find fsq id for index")
		return
	}

	return DbGrabLocation(get.Val())
}
