package business

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/common/foursquare"
	"mealswipe.app/mealswipe/internal/common/logging"
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

func convertStringNotNull(value interface{}) (result string, err error) {
	switch res := value.(type) {
	case string:
		result = res
	case nil:
		return
	default:
		err = fmt.Errorf("expected string or null for conversion, got %T (%#v)", value, value)
	}
	return
}

func DbLocationFromId(loc_id string, index int32) (loc *mealswipepb.Location, err error) {
	logger := logging.Get()

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
		logger.Error("failed to load loc from database", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index))
		return
	}

	vals := hmget.Val()
	miss := LOCATION_MODE_API && vals[0] == nil
	if miss || DISABLE_CACHING {
		// If the location wasn't in the DB, fetch it then mock a correct response
		if DISABLE_CACHING {
			logger.Info("cache miss (forced)", zap.Bool("cache_hit", false), logging.LocId(loc_id), zap.Int32("index", index))
		} else {
			logger.Info("cache miss", zap.Bool("cache_hit", false), logging.LocId(loc_id), zap.Int32("index", index))
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
			logger.Error("error marshalling tags from db values", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index))
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
		logger.Info("cache hit", zap.Bool("cache_hit", true), logging.LocId(loc_id), zap.Int32("index", index))
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

	name, _ := convertStringNotNull(vals[0])
	lat, _ := convertStringNotNull(vals[2])
	lng, _ := convertStringNotNull(vals[3])
	chain, _ := convertStringNotNull(vals[4])
	address, _ := convertStringNotNull(vals[5])
	mobileUrl, _ := convertStringNotNull(vals[9])
	url, _ := convertStringNotNull(vals[10])

	var priceTier int32
	priceTierString, _ := convertStringNotNull(vals[6])
	if priceTierString != "" {
		priceTier64, err := strconv.ParseInt(priceTierString, 10, 32)
		if err != nil {
			logger.Error("failed to convert priceTierString to value", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index), zap.Any("input", vals[6]))
		} else {
			priceTier = int32(priceTier64)
		}
	}

	var rating float32
	ratingString, _ := convertStringNotNull(vals[7])
	if ratingString != "" {
		rating64, err := strconv.ParseFloat(ratingString, 32)
		if err != nil {
			logger.Error("failed to convert ratingString to value", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index), zap.Any("input", vals[6]))
		} else {
			rating = float32(rating64)
		}
	}

	var ratingCount int32
	ratingCountString, _ := convertStringNotNull(vals[8])
	if ratingCountString != "" {
		ratingCount64, err := strconv.ParseInt(ratingCountString, 10, 32)
		if err != nil {
			logger.Error("failed to convert ratingCountString to value", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index), zap.Any("input", vals[6]))
		} else {
			ratingCount = int32(ratingCount64)
		}
	}

	var highlightColor int32
	highlightColorString, _ := convertStringNotNull(vals[11])
	if highlightColorString != "" {
		highlightColor64, err := strconv.ParseInt(highlightColorString, 10, 32)
		if err != nil {
			logger.Error("failed to convert highlightColorString to value", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index), zap.Any("input", vals[6]))
		} else {
			highlightColor = int32(highlightColor64)
		}
	}

	var textColor int32
	textColorString, _ := convertStringNotNull(vals[12])
	if textColorString != "" {
		textColor64, err := strconv.ParseInt(textColorString, 10, 32)
		if err != nil {
			logger.Error("failed to convert textColorString to value", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index), zap.Any("input", vals[6]))
		} else {
			textColor = int32(textColor64)
		}
	}

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
	return
}

func DbLocationIdFromInd(sessionId string, index int32) (locId string, distanceVal string, err error) {
	pipe := GetRedisClient().Pipeline()
	location := pipe.LIndex(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_LOCATIONS), int64(index))
	distance := pipe.LIndex(context.TODO(), BuildSessionKey(sessionId, KEY_SESSION_LOCATION_DISTANCES), int64(index))

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		return
	}

	if err = location.Err(); err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		return
	}
	if err = distance.Err(); err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		return
	}

	return location.Val(), distance.Val(), nil
}

func DbLocationFromInd(sessionId string, index int32) (loc *mealswipepb.Location, err error) {
	locId, distance, err := DbLocationIdFromInd(sessionId, index)
	if err != nil {
		return nil, err
	}

	if len(locId) == 0 {
		return &mealswipepb.Location{
			OutOfLocations: true,
		}, nil
	}

	loc, err = DbLocationFromId(locId, index)
	if err != nil {
		return
	}

	distInt, err := strconv.ParseInt(distance, 10, 32)
	if err != nil {
		logging.Get().Error("failed to convert distance to int", logging.SessionId(sessionId), logging.LocId(locId), zap.String("distance", distance))
	}
	loc.Distance = int32(distInt)

	return
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
func locationPhotoJsonFromVenue(venue foursquare.Venue) (encoded string, err error) { // TODO Convert to pointer
	logger := logging.Get()
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
		logger.Error("failed to marshall photos string into json", zap.Error(err), zap.String("input", string(photoBytes)), logging.LocId(venue.Id), logging.LocName(venue.Name))
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
	logger := logging.Get()
	// Sort by distance
	foursquare.By(foursquare.Distance).Sort(venues)

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
			logger.Info("skipping location because it has already been seen", logging.LocName(venue.Name), logging.LocId(venue.Name), zap.String("seen_name", seenVenue.Name), zap.String("seen_id", seenVenue.Id))
		}
	}

	logger.Info("unique locations sorted out", logging.Metric("loc_uniqueness"), zap.Int("unique", len(uniqueNames)), zap.Int("total", len(venues)))

	// Check our database for information about each location
	pipe := GetRedisClient().Pipeline()
	var cmds []*redis.SliceCmd
	var ids []string
	for _, venue := range uniqueNames {
		ids = append(ids, venue.Id)
		cmds = append(cmds, pipe.HMGet(
			context.TODO(),
			BuildLocKey(venue.Id),
			"blacklist", // Blacklist, to see if we are ignoring this place
			"name",      // Name, to see if it exists in db
		))
	}

	_, err = pipe.Exec(context.TODO())
	if err != nil {
		logger.Error("optimal values failed to get locations from db", zap.Error(err), zap.Any("loc_ids", ids))
		return nil, err
	}

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
			logger.Info("skipping location because of blacklist", logging.LocName(venue.Name), logging.LocId(venue.Name))
		}
	}

	// We now have a good set of hits and misses! Shuffle them out of distance sorted
	logger.Info("locations sorted into hit and miss", logging.Metric("hit_miss"), zap.Int("hits", len(hit)), zap.Int("misses", len(miss)), zap.Int("total", len(miss)+len(hit)))
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
			logger.Info("location placed into index", logging.Metric("loc_pos"), zap.Int("index", i), zap.Bool("cache_hit", true), logging.LocId(hit[0].Id)) // TODO Session id here would be good
			resultingVenues = append(resultingVenues, hit[0])
			hit = hit[1:]
		} else {
			logger.Info("location placed into index", logging.Metric("loc_pos"), zap.Int("index", i), zap.Bool("cache_hit", false), logging.LocId(miss[0].Id))
			resultingVenues = append(resultingVenues, miss[0])
			miss = miss[1:]
		}
	}

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

	return locArray, distArray, nil
}

// TODO Make this take the standard array
// TODO Make the standard array into a struct so she less messy (loc object probably)
func dbLocationWriteVenue(loc_id string, venue foursquare.Venue) (err error) {
	logger := logging.Get()
	if venue.Name == "" {
		logger.Warn("not saving location, looks incomplete", logging.LocId(loc_id), zap.Any("raw", venue))
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
		logger.Error("failed to marshal tags array into json for db write", logging.LocId(loc_id), zap.Error(err), zap.Any("raw", tags))
		return
	}
	encodedTags := string(tagBytes)

	pipe := GetRedisClient().Pipeline()

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
		logger.Error("failed to save location into redis", logging.LocId(loc_id), zap.Error(err), zap.Any("raw", tags))
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
