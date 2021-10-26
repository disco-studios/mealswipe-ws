package locations

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
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/internal/common/foursquare"
	"mealswipe.app/mealswipe/internal/common/logging"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/msredis"
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

func FromId(loc_id string, index int32) (loc *mealswipepb.Location, err error) {
	logger := logging.Get()

	get := msredis.GetRedisClient().Get(context.TODO(), keys.BuildLocKey(loc_id))

	miss := false
	if err = get.Err(); err != nil {
		if err != redis.Nil {
			logger.Error("failed to load loc from database", zap.Error(err), logging.LocId(loc_id), zap.Int32("index", index))
			return
		} else {
			miss = true
		}
	}

	var locationStore *mealswipepb.LocationStore = &mealswipepb.LocationStore{}
	if miss || DISABLE_CACHING {
		// If the location wasn't in the DB, fetch it then mock a correct response
		if DISABLE_CACHING {
			logger.Info("cache miss (forced)", zap.Bool("cache_hit", false), logging.LocId(loc_id), zap.Int32("index", index), logging.Metric("load_cache_hit"))
		} else {
			logger.Info("cache miss", zap.Bool("cache_hit", false), logging.LocId(loc_id), zap.Int32("index", index), logging.Metric("load_cache_hit"))
		}

		locationStore, err = GrabFreshAPI(loc_id)
		if err != nil {
			return
		}
	} else {
		logger.Info("cache hit", zap.Bool("cache_hit", true), logging.LocId(loc_id), zap.Int32("index", index), logging.Metric("load_cache_hit"))

		bytes, err := get.Bytes()
		if err != nil {
			return nil, err
		}

		if err := proto.Unmarshal(bytes, locationStore); err != nil {
			return nil, err
		}
	}

	loc = &mealswipepb.Location{
		Index:          index,
		Name:           locationStore.FoursquareLoc.Name,
		Lat:            strconv.FormatFloat(locationStore.FoursquareLoc.Lat, 'E', -1, 64),
		Lng:            strconv.FormatFloat(locationStore.FoursquareLoc.Lng, 'E', -1, 64),
		Chain:          locationStore.FoursquareLoc.Chain,
		Address:        locationStore.FoursquareLoc.Address,
		PriceTier:      locationStore.FoursquareLoc.PriceTier,
		Rating:         locationStore.FoursquareLoc.Rating,
		RatingCount:    locationStore.FoursquareLoc.RatingCount,
		MobileUrl:      locationStore.FoursquareLoc.MobileUrl,
		Url:            locationStore.FoursquareLoc.Url,
		HighlightColor: locationStore.FoursquareLoc.HighlightColor,
		TextColor:      locationStore.FoursquareLoc.TextColor,
		Tags:           locationStore.FoursquareLoc.Tags,
	}

	photos := locationStore.FoursquareLoc.Photos
	if len(photos) > 0 {
		loc.Photo = photos[0]
	}

	return
}

func IdFromInd(sessionId string, index int32) (locId string, distanceVal string, err error) {
	pipe := msredis.GetRedisClient().Pipeline()
	location := pipe.LIndex(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATIONS), int64(index))
	distance := pipe.LIndex(context.TODO(), keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATION_DISTANCES), int64(index))

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

func FromInd(sessionId string, index int32) (loc *mealswipepb.Location, err error) {
	locId, distance, err := IdFromInd(sessionId, index)
	if err != nil {
		return nil, err
	}

	if len(locId) == 0 {
		return &mealswipepb.Location{
			OutOfLocations: true,
		}, nil
	}

	loc, err = FromId(locId, index)
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

func IdsForLocation(lat float64, lng float64, radius int32, categoryId string) (loc_id []string, distances []float64, err error) {
	if LOCATION_MODE_API {
		return IdsForLocationAPI(lat, lng, radius, categoryId)
	} else {
		return IdsForLocationFlat(lat, lng, radius)
	}
}

/*
** Flat file implementation
 */

func IdsForLocationFlat(lat float64, lng float64, radius int32) (loc_ids []string, distances []float64, err error) {
	// TODO Replace with GeoSearch when redis client supports it
	geoRad := msredis.GetRedisClient().GeoRadius(context.TODO(), keys.BuildLocIndexKey("restaurants"), lng, lat, &redis.GeoRadiusQuery{
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
func locationPhotosFromVenue(venue foursquare.Venue) (photos []string) { // TODO Convert to pointer
	if len(venue.Photos.Groups) == 0 {
		return
	}

	// Build list of photo URLs
	for _, photo := range venue.Photos.Groups[0].Items {
		photos = append(photos, fmt.Sprintf(
			"%s1080x1920%s", //"%soriginal%s",
			photo.Prefix,
			photo.Suffix,
		))
	}

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
	pipe := msredis.GetRedisClient().Pipeline()
	var cmds []*redis.StringCmd
	// var ids []string
	for _, venue := range uniqueNames {
		// ids = append(ids, venue.Id)
		cmds = append(cmds, pipe.Get(
			context.TODO(),
			keys.BuildLocKey(venue.Id),
		))
	}

	_, _ = pipe.Exec(context.TODO())

	// Figure out hits and misses, while filtering out blacklisted locations
	var hit []foursquare.Venue
	var miss []foursquare.Venue
	for ind, venue := range uniqueNames {
		if cmds[ind].Err() != nil {
			miss = append(miss, venue)
		} else {
			hit = append(hit, venue)
		}
		// TODO Blacklist impl
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

func IdsForLocationAPI(lat float64, lng float64, radius int32, _categoryId string) (loc_id []string, distances []float64, err error) {
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
func WriteVenue(loc_id string, locationStore *mealswipepb.LocationStore) (err error) {
	logger := logging.Get()
	if locationStore.FoursquareLoc.Name == "" {
		logger.Warn("not saving location, looks incomplete", logging.LocId(loc_id), zap.Any("raw", locationStore))
		return
	}

	out, err := proto.Marshal(locationStore)
	if err != nil {
		logger.Error("failed to save location into redis", logging.LocId(loc_id), zap.Error(err), zap.Any("raw", locationStore))
	}

	set := msredis.GetRedisClient().SetEX(context.TODO(), keys.BuildLocKey(loc_id), out, time.Hour*24)

	err = set.Err()
	if err != nil {
		logger.Error("failed to save location into redis", logging.LocId(loc_id), zap.Error(err), zap.Any("raw", locationStore))
	}

	return
}

func ClearCache() (cleared_len int, err error) {
	redisClient := msredis.GetRedisClient()
	var cursor uint64
	var n int
	for {
		var dbkeys []string
		var err error
		dbkeys, cursor, err = redisClient.Scan(context.TODO(), cursor, fmt.Sprintf("%s*", keys.PREFIX_LOC_API), 15).Result()
		if err != nil {
			return n, err
		}

		n += len(dbkeys)
		if len(dbkeys) > 0 {
			for _, key := range dbkeys {
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

func GrabFreshAPI(loc_id string) (locationStore *mealswipepb.LocationStore, err error) {
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

	venue := respObj.Response.Venue

	// Map categories to tags
	var tagsArr []string
	for _, tag := range venue.Categories {
		tagsArr = append(tagsArr, tag.ShortName)
	}

	// // Map to expected DB returns
	// vals[0] = venue.Name                                                // name
	// vals[1] = encodedPhotos                                             // photos (json string list)
	// vals[2] = venue.Location.Lat                                        // lat
	// vals[3] = venue.Location.Lng                                        // lng
	// vals[4] = ""                                                        // chain // TODO see if we can get from API
	// vals[5] = venue.Location.Address                                    // Address
	// vals[6] = strconv.Itoa(int(venue.Price.Tier))                       // price tier
	// vals[7] = strconv.FormatFloat(float64(venue.Rating/2), 'E', -1, 32) // rating
	// vals[8] = strconv.Itoa(int(venue.RatingSignals))                    // rating count
	// vals[9] = venue.Menu.MobileUrl                                      // mobile menu url
	// vals[10] = venue.Menu.Url                                           // menu url
	// vals[11] = strconv.Itoa(int(venue.Colors.HighlightColor.Value))     // highlight color
	// vals[12] = strconv.Itoa(int(venue.Colors.HighlightTextColor.Value)) // highlight text color
	// vals[13] = tags

	locationStore = &mealswipepb.LocationStore{
		FoursquareLoc: &mealswipepb.FoursquareLocation{
			Name:           venue.Name,
			Photos:         locationPhotosFromVenue(venue),
			Lat:            venue.Location.Lat,
			Lng:            venue.Location.Lng,
			Address:        venue.Location.Address,
			PriceTier:      venue.Price.Tier,
			Rating:         venue.Rating / 2,
			RatingCount:    int32(venue.RatingSignals),
			MobileUrl:      venue.Menu.MobileUrl,
			Url:            venue.Menu.Url,
			HighlightColor: venue.Colors.HighlightColor.Value,
			TextColor:      venue.Colors.HighlightTextColor.Value,
			Tags:           tagsArr,
		},
	}

	// Save the result and return a venue
	// TODO This response shouldn't have to wait for the response from saving the venue
	err = WriteVenue(loc_id, locationStore)
	return
}
