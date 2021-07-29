package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"mealswipe.app/mealswipe/internal/common/foursquare"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

const LOCATION_MODE_API = true

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
	if LOCATION_MODE_API && vals[0] == nil {
		// If the location wasn't in the DB, fetch it then mock a correct response
		log.Println("> Cache miss", fsq_id)
		venue, err := dbLocationGrabFreshAPI(fsq_id)
		if err != nil {
			return nil, err
		}

		encodedPhotos, err := locationPhotoJsonFromVenue(venue)
		if err != nil {
			return nil, err
		}

		// Map to expected DB returns
		vals[0] = venue.Name             // name
		vals[1] = encodedPhotos          // photos (json string list)
		vals[2] = venue.Location.Lat     // lat
		vals[3] = venue.Location.Lng     // lng
		vals[4] = ""                     // chain // TODO see if we can get from API
		vals[5] = venue.Location.Address // Address
	} else {
		log.Println("> Cache hit", fsq_id)
	}

	var photo string
	var photos []string
	log.Print(vals[1])
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
	if LOCATION_MODE_API {
		return dbLocationIdsForLocationAPI(lat, lng)
	} else {
		return dbLocationIdsForLocationFlat(lat, lng)
	}
}

/*
** Flat file implementation
 */

func dbLocationIdsForLocationFlat(lat float64, lng float64) (fsq_ids []string, distances []float64, err error) {
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

/*
** API implementation
 */
func locationPhotoJsonFromVenue(venue foursquare.Venue) (encoded string, err error) {
	// Build list of photo URLs
	var photos []string
	for _, photo := range venue.Photos.Groups[0].Items {
		photos = append(photos, fmt.Sprintf(
			"%soriginal%s",
			photo.Prefix,
			photo.Suffix,
		))
	}

	// Encode photo URLs into JSON to match DB format
	photoBytes, err := json.Marshal(venue.Location.Lat)
	if err != nil {
		log.Println("Failed to marshal photos into json", photos)
		return
	}

	encoded = string(photoBytes)
	return
}

func dbLocationIdsForLocationAPI(lat float64, lng float64) (fsq_ids []string, distances []float64, err error) {
	requestUrl := fmt.Sprintf(
		"https://api.foursquare.com/v2/venues/search?client_id=%s&client_secret=%s&v=%s&ll=%f,%f&intent=browse&radius=%d&limit=50&categoryId=%s",
		"UIEPSPWBZLULKZJQGT3KNRBX40O4GHBKA1SZ404HCMTUYCSN", // client id
		"3QD0PJNSFOJTWWLZCGO3ERHCTQEVA4L11LSEFFDLAOKFSDVR", // client secret
		"20210726",                 // version
		lat,                        // lat
		lng,                        // lng
		2000,                       // radius (m)
		"4d4b7105d754a06374d81259", // category id (4d4b7105d754a06374d81259 food, 4bf58dd8d48988d14c941735 fast food)
	)

	// Make the request
	resp, err := http.Get(requestUrl)
	if err != nil {
		return
	}

	// Read the bytes in from the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Turn the response into a struct
	respObj := &foursquare.LocationRequestResponse{}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	// Turn the response into an array of IDs and distances
	var locArray []string
	var distArray []float64
	for _, venue := range respObj.Response.Venues {
		locArray = append(locArray, venue.Id)
		distArray = append(distArray, float64(venue.Location.Distance))
	}

	log.Print("Got locs", locArray)
	return locArray, distArray, nil
}

func dbLocationWriteVenue(fsq_id string, venue foursquare.Venue) (err error) {
	encodedPhotos, err := locationPhotoJsonFromVenue(venue)
	if err != nil {
		return
	}

	pipe := GetRedisClient().Pipeline()

	pipe.HSet(context.TODO(), BuildLocKey(fsq_id), map[string]interface{}{
		"name":      venue.Name,
		"photos":    encodedPhotos,
		"latitude":  venue.Location.Lat,
		"longitude": venue.Location.Lng,
		// "chain_name": "", // TODO We can maybe get this
		"address": venue.Location.Address,
	})
	pipe.Expire(context.TODO(), BuildLocKey(fsq_id), time.Hour*24) // We can only hold API data for 24 hours

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Print("can't save API result")
	}
	return
}

func dbLocationGrabFreshAPI(fsq_id string) (venue foursquare.Venue, err error) {

	requestUrl := fmt.Sprintf(
		"https://api.foursquare.com/v2/venues/%s?client_id=%s&client_secret=%s&v=%s",
		fsq_id, // venue ID
		"UIEPSPWBZLULKZJQGT3KNRBX40O4GHBKA1SZ404HCMTUYCSN", // client id
		"3QD0PJNSFOJTWWLZCGO3ERHCTQEVA4L11LSEFFDLAOKFSDVR", // client secret
		"20210726", // version
	)

	// Make the request
	resp, err := http.Get(requestUrl)
	if err != nil {
		return
	}

	// Read the bytes in from the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Turn the response into a struct
	respObj := &foursquare.VenueRequestResponse{}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	venue = respObj.Response.Venue

	// Save the result and return a venue
	return venue, dbLocationWriteVenue(fsq_id, venue)
}
