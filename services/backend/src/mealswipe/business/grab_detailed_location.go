package business

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type FS_Venue_Detailed struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FS_Detailed_Venue_Stepper struct {
	Venue FS_Venue_Detailed `json:"venue"`
}

type FS_Venue_Detail_Result struct {
	Venue FS_Detailed_Venue_Stepper `json:"response"`
}

func tagDetailedVenueGrabs(venueId string, fresh bool) (err error) {
	pipe := redisClient.Pipeline()

	pipe.Incr(context.TODO(), "venue."+venueId+".impressions")
	if fresh {
		pipe.Incr(context.TODO(), "venue."+venueId+".detailedgrabs")
	}

	_, err = pipe.Exec(context.TODO())
	return
}

func getDetailedVenueIfCached(venueId string) (raw string, err error) {
	key := "venue." + venueId + ".cache"
	result := redisClient.Get(context.TODO(), key)
	return result.Val(), result.Err()
}

func cacheDetailedVenue(venueId string, data string) error {
	key := "venue." + venueId + ".cache"
	result := redisClient.Set(context.TODO(), key, data, time.Hour*24)
	return result.Err()
}

func getDetailedVenueFresh(venueId string) (raw string, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.foursquare.com/v2/venues/%s?client_id=UIEPSPWBZLULKZJQGT3KNRBX40O4GHBKA1SZ404HCMTUYCSN&client_secret=3QD0PJNSFOJTWWLZCGO3ERHCTQEVA4L11LSEFFDLAOKFSDVR&v=20210620", venueId))
	if err != nil {
		return
	}
	log.Println(resp)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return string(bodyBytes), nil

}

func GetDetailedLocation(venueId string) (venue FS_Venue_Detailed, err error) {
	rawJson, err := getDetailedVenueIfCached(venueId)
	if err != nil {
		log.Println("Cache missed, trying to get fresh")
		rawJson, err = getDetailedVenueFresh(venueId)
		if err != nil {
			return
		}
		log.Println("Caching the location")

		tagErr := tagDetailedVenueGrabs(venueId, true)
		if tagErr != nil {
			log.Println(tagErr) // TODO Non fatal log
		}

		// TODO This can be done async of the response, and shouldn't make response fails if it fails
		cacheErr := cacheDetailedVenue(venueId, rawJson)
		if cacheErr != nil {
			log.Println(tagErr) // TODO Non fatal log
		}
	} else {
		tagErr := tagDetailedVenueGrabs(venueId, false)
		if tagErr != nil {
			log.Println(tagErr) // TODO Non fatal log
		}
		log.Println("Cache hit")
	}

	log.Println("Got location detail, making json")

	var res FS_Venue_Detail_Result
	err = json.Unmarshal([]byte(rawJson), &res)
	return res.Venue.Venue, err
}
