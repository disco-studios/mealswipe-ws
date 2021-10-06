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
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"mealswipe.app/mealswipe/internal/common/foursquare"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

const LOCATION_MODE_API = true
const HITS_BEFORE_MISS = 4 + 1 // show 4 hits until show miss
var ALLOWED_CATEGORIES = []string{
	"4bf58dd8d48988d116941735", // Bars
	"4bf58dd8d48988d16e941735", // Fast Food
	"4bf58dd8d48988d1d0941735", // Dessert
	"4bf58dd8d48988d1e0931735", // Coffee
	"4bf58dd8d48988d143941735", // Breakfast Spot
	"4bf58dd8d48988d142941735", // Asian
	"4bf58dd8d48988d1c1941735", // Mexican
	"4bf58dd8d48988d14e941735", // American
	"4bf58dd8d48988d110941735", // Italian
	"4bf58dd8d48988d10e941735", // Greek
	"4bf58dd8d48988d1ca941735", // Pizza
	"4bf58dd8d48988d1d3941735", // Vegetarian / Vegan
	"4d4b7105d754a06374d81259", // Food
}

const DISABLE_CACHING = false

func DbLocationFromId(loc_id string, index int32) (loc *mealswipepb.Location, err error) {
	hmget := GetRedisClient().HMGet(
		context.TODO(),
		BuildLocKey(loc_id),
		"name",
		"photos",
		"latitude",
		"longitude",
		"chain_name",
		"address",
		"priceTier",
		"rating",
		"ratingCount",
		"mobileMenuUrl",
		"menuUrl",
		"highlightColor",
		"textColor",
		"tags",
	)

	if err = hmget.Err(); err != nil {
		return
	}

	vals := hmget.Val()
	log.Println("Locs loaded for", loc_id, vals)
	miss := LOCATION_MODE_API && vals[0] == nil
	if miss || DISABLE_CACHING {
		// If the location wasn't in the DB, fetch it then mock a correct response
		if DISABLE_CACHING {
			log.Println("> Forced miss", loc_id)
		} else {
			log.Println("> Cache miss", loc_id)
		}
		venue, err := dbLocationGrabFreshAPI(loc_id)
		if err != nil {
			return nil, err
		}

		encodedPhotos, err := locationPhotoJsonFromVenue(venue)
		if err != nil {
			return nil, err
		}

		// Map categories to tags
		var tagsArr []string
		for _, tag := range venue.Categories {
			tagsArr = append(tagsArr, tag.ShortName)
		}
		rawBytes, err := json.Marshal(tagsArr)
		if err != nil {
			log.Println("Error marshalling tags", err)
		}
		tags := string(rawBytes)

		// Map to expected DB returns
		vals[0] = venue.Name                                                // name
		vals[1] = encodedPhotos                                             // photos (json string list)
		vals[2] = venue.Location.Lat                                        // lat
		vals[3] = venue.Location.Lng                                        // lng
		vals[4] = ""                                                        // chain // TODO see if we can get from API
		vals[5] = venue.Location.Address                                    // Address
		vals[6] = strconv.Itoa(int(venue.Price.Tier))                       // price tier
		vals[7] = strconv.FormatFloat(float64(venue.Rating/2), 'E', -1, 32) // rating
		vals[8] = strconv.Itoa(int(venue.RatingSignals))                    // rating count
		vals[9] = venue.Menu.MobileUrl                                      // mobile menu url
		vals[10] = venue.Menu.Url                                           // menu url
		vals[11] = strconv.Itoa(int(venue.Colors.HighlightColor.Value))     // highlight color
		vals[12] = strconv.Itoa(int(venue.Colors.HighlightTextColor.Value)) // highlight text color
		vals[13] = tags
	} else {
		log.Println("> Cache hit", loc_id)
	}

	// Register statistics async
	go StatsRegisterLocLoad(loc_id, miss)

	var photo string
	var photos []string
	json.Unmarshal([]byte(vals[1].(string)), &photos)
	if len(photos) > 0 {
		photo = photos[0]
	}

	var tags []string
	switch vals[13].(type) {
	case string:
		json.Unmarshal([]byte(vals[13].(string)), &tags)
	}

	// TODO I feel like a disgusting human being for writing this, there must
	// be a better way :(
	var name string
	switch vals[0].(type) {
	case string:
		name = vals[0].(string)
	}
	var lat string
	switch vals[2].(type) {
	case string:
		lat = vals[2].(string)
	}
	var lng string
	switch vals[3].(type) {
	case string:
		lng = vals[3].(string)
	}
	var chain string
	switch vals[4].(type) {
	case string:
		chain = vals[4].(string)
	}
	var address string
	switch vals[5].(type) {
	case string:
		address = vals[5].(string)
	}
	var priceTier int32
	switch vals[6].(type) {
	case string:
		priceTier64, err := strconv.ParseInt(vals[6].(string), 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			priceTier = int32(priceTier64)
		}
	default:
		log.Println("Missed priceTier", fmt.Sprintf("%T", vals[6]))
	}
	var rating float32
	switch vals[7].(type) {
	case string:
		rating64, err := strconv.ParseFloat(vals[7].(string), 32)
		if err != nil {
			log.Println(err)
		} else {
			rating = float32(rating64)
		}
	default:
		log.Println("Missed rating", fmt.Sprintf("%T", vals[7]))
	}
	var ratingCount int32
	switch vals[8].(type) {
	case string:
		ratingCount64, err := strconv.ParseInt(vals[8].(string), 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			ratingCount = int32(ratingCount64)
		}
	default:
		log.Println("Missed ratingCount", fmt.Sprintf("%T", vals[8]))
	}
	var mobileUrl string
	switch vals[9].(type) {
	case string:
		mobileUrl = vals[9].(string)
	}
	var url string
	switch vals[10].(type) {
	case string:
		url = vals[10].(string)
	}
	var highlightColor int32
	switch vals[11].(type) {
	case string:
		highlightColor64, err := strconv.ParseInt(vals[11].(string), 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			highlightColor = int32(highlightColor64)
		}
	default:
		log.Println("Missed highlightColor", fmt.Sprintf("%T", vals[11]))
	}
	var textColor int32
	switch vals[12].(type) {
	case string:
		textColor64, err := strconv.ParseInt(vals[12].(string), 10, 32)
		if err != nil {
			log.Println(err)
		} else {
			textColor = int32(textColor64)
		}
	default:
		log.Println("Missed textColor", fmt.Sprintf("%T", vals[12]))
	}

	log.Println("Pre-location:", loc_id)
	log.Println("Index", index)
	log.Println("Name", name)
	log.Println("Photo", photo)
	log.Println("Lat", lat)
	log.Println("Lng", lng)
	log.Println("Chain", chain)
	log.Println("Address", address)
	log.Println("PriceTier", priceTier)
	log.Println("Rating", rating)
	log.Println("RatingCount", ratingCount)
	log.Println("MobileUrl", mobileUrl)
	log.Println("Url", url)
	log.Println("HighlightColor", highlightColor)
	log.Println("TextColor", textColor)
	log.Println("Tags", tags)
	loc = &mealswipepb.Location{
		Index:          index,
		Name:           name,
		Photo:          photo,
		Lat:            lat,
		Lng:            lng,
		Chain:          chain,
		Address:        address,
		PriceTier:      priceTier,
		Rating:         rating,
		RatingCount:    ratingCount,
		MobileUrl:      mobileUrl,
		Url:            url,
		HighlightColor: highlightColor,
		TextColor:      textColor,
		Tags:           tags,
	}
	log.Println(">>", loc)
	return
}

func DbLocationIdFromInd(sessionId string, index int32) (locId string, err error) {
	get := GetRedisClient().LIndex(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), int64(index))
	if err = get.Err(); err != nil {
		return
	}

	if len(get.Val()) == 0 {
		err = errors.New("couldn't find loc id for index")
		return
	}

	return get.Val(), nil
}

func DbLocationFromInd(sessionId string, index int32) (loc *mealswipepb.Location, err error) {
	locId, err := DbLocationIdFromInd(sessionId, index)
	if err != nil {
		return nil, err
	}

	return DbLocationFromId(locId, index)
}

func DbLocationIdsForLocation(lat float64, lng float64, radius int32, categoryId string) (loc_id []string, distances []float64, err error) {
	if LOCATION_MODE_API {
		return dbLocationIdsForLocationAPI(lat, lng, radius, categoryId)
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
			"%s1080x1920%s", //"%soriginal%s",
			photo.Prefix,
			photo.Suffix,
		))
	}

	// Encode photo URLs into JSON to match DB format
	photoBytes, err := json.Marshal(photos)
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
		log.Println("can't get locations from db for optimization")
	}

	log.Println("Got results!")

	// Figure out hits and misses, while filtering out blacklisted locations
	var hit []foursquare.Venue
	var miss []foursquare.Venue
	for ind, venue := range uniqueNames {
		vals := cmds[ind].Val()
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
			log.Println("\t> h", i)
			resultingVenues = append(resultingVenues, hit[0])
			hit = hit[1:]
		} else {
			log.Println("\t> m", i)
			resultingVenues = append(resultingVenues, miss[0])
			miss = miss[1:]
		}
	}

	log.Println("\n\t> Total: ", len(resultingVenues))

	return
}

func dbLocationIdsForLocationAPI(lat float64, lng float64, radius int32, _categoryId string) (loc_id []string, distances []float64, err error) {
	categoryId := "4d4b7105d754a06374d81259" // category id (4d4b7105d754a06374d81259 food, 4bf58dd8d48988d14c941735 fast food)
	if _categoryId != "" {
		for _, allowedCategoryId := range ALLOWED_CATEGORIES {
			if _categoryId == allowedCategoryId {
				categoryId = _categoryId
				break
			}
		}
	}

	requestUrl := fmt.Sprintf(
		"https://api.foursquare.com/v2/venues/search?client_id=%s&client_secret=%s&v=%s&ll=%f,%f&intent=browse&radius=%d&limit=50&categoryId=%s",
		"UIEPSPWBZLULKZJQGT3KNRBX40O4GHBKA1SZ404HCMTUYCSN", // client id
		"3QD0PJNSFOJTWWLZCGO3ERHCTQEVA4L11LSEFFDLAOKFSDVR", // client secret
		"20210726", // version
		lat,        // lat
		lng,        // lng
		radius,     // radius (m)
		categoryId, // category id (4d4b7105d754a06374d81259 food, 4bf58dd8d48988d14c941735 fast food)
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

	log.Println("Got locs", locArray)
	return locArray, distArray, nil
}

// TODO Make this take the standard array
// TODO Make the standard array into a struct so she less messy (loc object probably)
func dbLocationWriteVenue(loc_id string, venue foursquare.Venue) (err error) {
	if venue.Name == "" {
		log.Println("Not saving location, looks incomplete", venue)
		return
	}

	encodedPhotos, err := locationPhotoJsonFromVenue(venue)
	if err != nil {
		return
	}

	var tags []string
	for _, tag := range venue.Categories {
		tags = append(tags, tag.ShortName)
	}
	tagBytes, err := json.Marshal(tags)
	if err != nil {
		log.Println("Failed to marshal tags into json", tags)
		return
	}
	encodedTags := string(tagBytes)

	pipe := GetRedisClient().Pipeline()

	log.Println("Loading", loc_id)
	log.Println("\t> name", venue.Name)
	log.Println("\t> photos", encodedPhotos)
	log.Println("\t> latitude", venue.Location.Lat)
	log.Println("\t> longitude", venue.Location.Lng)
	log.Println("\t> address", venue.Location.Address)
	log.Println("\t> priceTier", venue.Price.Tier)
	log.Println("\t> rating", venue.Rating/2)
	log.Println("\t> ratingCount", venue.RatingSignals)
	log.Println("\t> mobileMenuUrl", venue.Menu.MobileUrl)
	log.Println("\t> menuUrl", venue.Menu.Url)
	log.Println("\t> highlightColor", venue.Colors.HighlightColor.Value)
	log.Println("\t> textColor", venue.Colors.HighlightTextColor.Value)
	log.Println("\t> tags", encodedTags)
	pipe.HSet(context.TODO(), BuildLocKey(loc_id), map[string]interface{}{
		"name":           venue.Name,
		"photos":         encodedPhotos,
		"latitude":       venue.Location.Lat,
		"longitude":      venue.Location.Lng,
		"address":        venue.Location.Address,
		"priceTier":      venue.Price.Tier,
		"rating":         venue.Rating / 2,
		"ratingCount":    venue.RatingSignals,
		"mobileMenuUrl":  venue.Menu.MobileUrl,
		"menuUrl":        venue.Menu.Url,
		"highlightColor": venue.Colors.HighlightColor.Value,
		"textColor":      venue.Colors.HighlightTextColor.Value,
		"tags":           encodedTags,
		// "chain_name": "", // TODO We can maybe get this
	})
	pipe.Expire(context.TODO(), BuildLocKey(loc_id), time.Hour*24) // We can only hold API data for 24 hours

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't save API result")
	}
	return
}

func DbClearCache() (cleared_len int, err error) {
	redisClient := GetRedisClient()
	var cursor uint64
	var n int
	for {
		var keys []string
		var err error
		keys, cursor, err = redisClient.Scan(context.TODO(), cursor, fmt.Sprintf("%s*", PREFIX_LOC_API), 15).Result()
		if err != nil {
			return n, err
		}

		n += len(keys)
		if len(keys) > 0 {
			for _, key := range keys {
				_, err = redisClient.Del(context.TODO(), key).Result()
				if err != nil {
					return n, err
				}
			}
		} else {
			break
		}

		if cursor == 0 {
			break
		}
	}
	return n, nil
}

func dbLocationGrabFreshAPI(loc_id string) (venue foursquare.Venue, err error) {
	requestUrl := fmt.Sprintf(
		"https://api.foursquare.com/v2/venues/%s?client_id=%s&client_secret=%s&v=%s",
		loc_id, // venue ID
		"GGI531X4VKM04LSSKKX1XNHCRRXZL5PPXFLCGAW233SLVJ0J", // client id
		"RQYEPCI2F4WSTCV1Y20V4IGBRDUMPLQBUARDBAVEEPGS12VJ", // client secret
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

	log.Println(string(body))

	// Turn the response into a struct
	respObj := &foursquare.VenueRequestResponse{}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	venue = respObj.Response.Venue

	// Save the result and return a venue
	// TODO This response shouldn't have to wait for the response from saving the venue
	return venue, dbLocationWriteVenue(loc_id, venue)
}
