package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"mealswipe.app/mealswipe/internal/common/foursquare"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

const LOCATION_MODE_API = true
const HITS_BEFORE_MISS = 4 + 1 // show 4 hits until show miss

func DbLocationFromId(loc_id string) (loc *mealswipepb.Location, err error) {
	hmget := GetRedisClient().HMGet(
		context.TODO(),
		BuildLocKey(loc_id),
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
		log.Println("> Cache miss", loc_id)
		venue, err := dbLocationGrabFreshAPI(loc_id)
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
		log.Println("> Cache hit", loc_id)
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
		err = errors.New("couldn't find loc id for index")
		return
	}

	return DbLocationFromId(get.Val())
}

func DbLocationIdsForLocation(lat float64, lng float64, radius int32) (loc_id []string, distances []float64, err error) {
	if LOCATION_MODE_API {
		return dbLocationIdsForLocationAPI(lat, lng, radius)
	} else {
		return dbLocationIdsForLocationFlat(lat, lng, radius)
	}
}

/*
** Flat file implementation
 */

func dbLocationIdsForLocationFlat(lat float64, lng float64, radius int32) (loc_ids []string, distances []float64, err error) {
	// TODO Replace with GeoSearch when redis client supports it
	geoRad := GetRedisClient().GeoRadius(context.TODO(), BuildLocIndexKey("restaurants"), lng, lat, &redis.GeoRadiusQuery{
		Radius:   float64(radius),
		Unit:     "m",
		WithDist: true,
	})

	if err = geoRad.Err(); err != nil {
		return
	}

	for _, loc := range geoRad.Val() {
		loc_ids = append(loc_ids, loc.Name)
		distances = append(distances, loc.Dist)
	}
	return
}

/*
** API implementation
 */
func locationPhotoJsonFromVenue(venue foursquare.Venue) (encoded string, err error) {
	if len(venue.Photos.Groups) == 0 {
		return "[]", nil
	}

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

func shuffleVenues(venues []foursquare.Venue) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(venues), func(i, j int) { venues[i], venues[j] = venues[j], venues[i] })
}

func findOptimalVenues(venues []foursquare.Venue) (resultingVenues []foursquare.Venue, err error) {
	// Sort by distance
	foursquare.By(foursquare.Distance).Sort(venues)
	log.Println("Filtering ", len(venues), " locations")

	// Filter out duplicate names. We don't want to pay for 2 Wawa or Dunkin Donut requests
	// Since sorted by distance, we will keep the closest instance of a place
	// Not perfect, may not have perfect match names or may be two places with same name. Close enough for now
	seenNames := make(map[string]foursquare.Venue)
	var uniqueNames []foursquare.Venue
	for _, venue := range venues {
		seenVenue, exists := seenNames[venue.Name]
		if !exists {
			seenNames[venue.Name] = venue
			uniqueNames = append(uniqueNames, venue)
		} else {
			log.Println("\t> Location ", venue.Name, " at", venue.Location.Address, "skipped because we have already seen one at ", seenVenue.Location.Address)
		}
	}

	// Check our database for information about each location
	pipe := GetRedisClient().Pipeline()
	var cmds []*redis.SliceCmd
	for _, venue := range uniqueNames {
		cmds = append(cmds, pipe.HMGet(
			context.TODO(),
			BuildLocKey(venue.Id),
			"blacklist", // Blacklist, to see if we are ignoring this place
			"name",      // Name, to see if it exists in db
		))
	}

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Print("can't get locations from db for optimization")
	}

	log.Println("Got results!")

	// Figure out hits and misses, while filtering out blacklisted locations
	var hit []foursquare.Venue
	var miss []foursquare.Venue
	for ind, venue := range uniqueNames {
		vals := cmds[ind].Val()
		log.Println("\t\t> ", vals)
		if vals[0] == nil {
			// Location wasn't blacklisted! Now lets see if we have it saved
			if vals[1] == nil {
				miss = append(miss, venue)
			} else {
				hit = append(hit, venue)
			}
		} else {
			log.Println("\t> Skipped ", venue.Name, venue.Id, " due to blacklist")
		}
	}

	// We now have a good set of hits and misses! Shuffle them out of distance sorted
	log.Println("\t> Had ", len(hit), " hit to ", len(miss), "misses, ", len(miss)+len(hit), " total")
	shuffleVenues(hit)
	shuffleVenues(miss)

	totalVenues := len(hit) + len(miss)
	// Now we want to prioritize cache hits. Lets do it
	for i := 0; i < totalVenues; i++ {
		// Prefer a miss if we are on a miss index and have enough left
		preferHit := !(((i+1)%HITS_BEFORE_MISS == 0) && (len(miss) > 0))
		// Only prefer a hit if we have enough hits to supply still
		preferHit = (len(hit) > 0) && preferHit

		if preferHit {
			log.Print("\t> h", i)
			resultingVenues = append(resultingVenues, hit[0])
			hit = hit[1:]
		} else {
			log.Print("\t> m", i)
			resultingVenues = append(resultingVenues, miss[0])
			miss = miss[1:]
		}
	}

	log.Println("\n\t> Total: ", len(resultingVenues))

	return
}

func dbLocationIdsForLocationAPI(lat float64, lng float64, radius int32) (loc_id []string, distances []float64, err error) {
	requestUrl := fmt.Sprintf(
		"https://api.foursquare.com/v2/venues/search?client_id=%s&client_secret=%s&v=%s&ll=%f,%f&intent=browse&radius=%d&limit=50&categoryId=%s",
		"UIEPSPWBZLULKZJQGT3KNRBX40O4GHBKA1SZ404HCMTUYCSN", // client id
		"3QD0PJNSFOJTWWLZCGO3ERHCTQEVA4L11LSEFFDLAOKFSDVR", // client secret
		"20210726",                 // version
		lat,                        // lat
		lng,                        // lng
		radius,                     // radius (m)
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

	// Optimize the returned venues
	venues, err := findOptimalVenues(respObj.Response.Venues)
	if err != nil {
		return
	}

	// Turn the result into an array of IDs and distances
	var locArray []string
	var distArray []float64
	for _, venue := range venues {
		locArray = append(locArray, venue.Id)
		distArray = append(distArray, float64(venue.Location.Distance))
	}

	log.Print("Got locs", locArray)
	return locArray, distArray, nil
}

func dbLocationWriteVenue(loc_id string, venue foursquare.Venue) (err error) {
	encodedPhotos, err := locationPhotoJsonFromVenue(venue)
	if err != nil {
		return
	}

	pipe := GetRedisClient().Pipeline()

	pipe.HSet(context.TODO(), BuildLocKey(loc_id), map[string]interface{}{
		"name":      venue.Name,
		"photos":    encodedPhotos,
		"latitude":  venue.Location.Lat,
		"longitude": venue.Location.Lng,
		// "chain_name": "", // TODO We can maybe get this
		"address": venue.Location.Address,
	})
	pipe.Expire(context.TODO(), BuildLocKey(loc_id), time.Hour*24) // We can only hold API data for 24 hours

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Print("can't save API result")
	}
	return
}

func dbLocationGrabFreshAPI(loc_id string) (venue foursquare.Venue, err error) {

	requestUrl := fmt.Sprintf(
		"https://api.foursquare.com/v2/venues/%s?client_id=%s&client_secret=%s&v=%s",
		loc_id, // venue ID
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
	return venue, dbLocationWriteVenue(loc_id, venue)
}
