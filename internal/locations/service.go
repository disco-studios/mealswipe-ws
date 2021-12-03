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

	"github.com/Treble-Development/mealswipe-proto/mealswipepb"
	"github.com/go-redis/redis/v8"
	"go.elastic.co/apm"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"mealswipe.app/mealswipe/internal/config"
	"mealswipe.app/mealswipe/internal/keys"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/internal/msredis"
	"mealswipe.app/mealswipe/internal/types"
)

var HITS_BEFORE_MISS = config.GetenvIntErrorless("MS_HITS_BEFORE_FRESH", 4) + 1 // show 4 hits until show miss
var ALLOWED_CATEGORIES = config.GetenvStrArrErrorless("MS_ALLOWED_CATEGORIES", []string{
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
})

func fromIdCached(ctx context.Context, loc_id string) (miss bool, locStore *mealswipepb.LocationStore, err error) {
	span, ctx := apm.StartSpan(ctx, "fromIdCached", "locations")
	defer span.End()

	get := msredis.Get(ctx, keys.BuildLocKey(loc_id))

	err = get.Err()
	if err == redis.Nil {
		return true, nil, nil
	} else if err != nil {
		err = fmt.Errorf("get from redis: %v", err)
		return
	}

	bytes, err := get.Bytes()
	if err != nil {
		err = fmt.Errorf("message to bytes: %v", err)
		return
	}

	locStore = &mealswipepb.LocationStore{}
	if err = proto.Unmarshal(bytes, locStore); err != nil {
		err = fmt.Errorf("unmarshal cached locstore: %v", err)
		return
	}

	return
}

func fromIdFresh(ctx context.Context, loc_id string) (locationStore *mealswipepb.LocationStore, err error) {
	span, ctx := apm.StartSpan(ctx, "fromIdFresh", "locations")
	defer span.End()

	reqspan, _ := apm.StartSpan(ctx, "fromIdFresh", "foursquare-api")
	defer reqspan.End()

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
		err = fmt.Errorf("req loc from api: %v", err)
		return
	}

	// Read the bytes in from the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("reading loc from resp: %v", err)
		return
	}

	reqspan.End()

	// Turn the response into a struct
	respObj := &types.VenueRequestResponse{}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		err = fmt.Errorf("marshal response to venue obj: %v", err)
		return
	}

	venue := respObj.Response.Venue

	// Map categories to tags
	var tagsArr []string
	for _, tag := range venue.Categories {
		tagsArr = append(tagsArr, tag.ShortName)
	}

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
	return
}

func fromStore(locationStore *mealswipepb.LocationStore, index int32) (loc *mealswipepb.Location, err error) {
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

// TODO Make this take the standard array
// TODO Make the standard array into a struct so she less messy (loc object probably)
func writeLocationStore(ctx context.Context, loc_id string, locationStore *mealswipepb.LocationStore) (err error) {
	span, ctx := apm.StartSpan(ctx, "writeLocationStore", "locations")
	defer span.End()

	if locationStore.FoursquareLoc.Name == "" {
		err = fmt.Errorf("not caching, bad data: %s", loc_id)
		return
	}

	out, err := proto.Marshal(locationStore)
	if err != nil {
		err = fmt.Errorf("marhsal locstore: %v", err)
		return
	}

	set := msredis.SetEX(ctx, keys.BuildLocKey(loc_id), out, time.Hour*24)

	err = set.Err()
	if err != nil {
		err = fmt.Errorf("redis set: %v", err)
	}

	return
}

func idFromInd(ctx context.Context, sessionId string, index int32) (locId string, distanceVal string, err error) {
	span, ctx := apm.StartSpan(ctx, "idFromInd", "locations")
	defer span.End()

	pipe := msredis.Pipeline()
	location := pipe.LIndex(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATIONS), int64(index))
	distance := pipe.LIndex(ctx, keys.BuildSessionKey(sessionId, keys.KEY_SESSION_LOCATION_DISTANCES), int64(index))

	_, err = pipe.Exec(ctx)
	if err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		err = fmt.Errorf("redis pipe exec: %v", err)
		return
	}

	if err = location.Err(); err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		err = fmt.Errorf("location get from pipe: %v", err)
		return
	}

	if err = distance.Err(); err != nil {
		if err == redis.Nil {
			return "", "", nil
		}
		err = fmt.Errorf("distance get from pipe: %v", err)
		return
	}

	return location.Val(), distance.Val(), nil
}

func getLocationsNear(ctx context.Context, lat float64, lng float64, radius int32, categoryId string) (respObj *types.LocationRequestResponse, err error) {
	span, _ := apm.StartSpan(ctx, "getLocationsNear", "locations")
	defer span.End()

	reqspan, _ := apm.StartSpan(ctx, "getLocationsNear", "foursquare-api")
	defer reqspan.End()

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
		err = fmt.Errorf("api req get: %v", err)
		return
	}

	// Read the bytes in from the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read bytes from resp: %v", err)
		return
	}

	reqspan.End()

	// Turn the response into a struct
	respObj = &types.LocationRequestResponse{}
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		err = fmt.Errorf("marshal response to venue obj: %v", err)
		return
	}

	return
}

func venuesToArrays(venues []types.Venue) (locArray []string, distArray []float64) {
	for _, venue := range venues {
		locArray = append(locArray, venue.Id)
		distArray = append(distArray, float64(venue.Location.Distance))
	}
	return
}

/*
** API implementation
 */
func locationPhotosFromVenue(venue types.Venue) (photos []string) { // TODO Convert to pointer
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

func shuffleVenues(venues []types.Venue) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(venues), func(i, j int) { venues[i], venues[j] = venues[j], venues[i] })
}

func findOptimalVenues(ctx context.Context, venues []types.Venue) (resultingVenues []types.Venue, err error) {
	span, ctx := apm.StartSpan(ctx, "findOptimalVenues", "locations")
	defer span.End()

	// Sort by distance
	types.By(types.Distance).Sort(venues)

	// Filter out duplicate names. We don't want to pay for 2 Wawa or Dunkin Donut requests
	// Since sorted by distance, we will keep the closest instance of a place
	// Not perfect, may not have perfect match names or may be two places with same name. Close enough for now
	seenNames := make(map[string]types.Venue)
	var uniqueNames []types.Venue
	for _, venue := range venues {
		seenVenue, exists := seenNames[venue.Name]
		if !exists {
			seenNames[venue.Name] = venue
			uniqueNames = append(uniqueNames, venue)
		} else {
			logging.MetricCtx(ctx, "repeat_loc").Info(
				fmt.Sprintf("skipping location '%s' (%s), already seen", venue.Name, venue.Id),
				logging.LocName(venue.Name),
				logging.LocId(venue.Name),
				zap.String("seen_name", seenVenue.Name),
				zap.String("seen_id", seenVenue.Id),
			)
		}
	}

	logging.MetricCtx(ctx, "loc_uniqueness").Info(
		fmt.Sprintf("found %d unique locations out of %d", len(uniqueNames), len(venues)),
		zap.Int("unique", len(uniqueNames)),
		zap.Int("total", len(venues)),
	)

	// Check our database for information about each location
	pipe := msredis.Pipeline()
	var cmds []*redis.StringCmd
	// var ids []string
	for _, venue := range uniqueNames {
		// ids = append(ids, venue.Id)
		cmds = append(cmds, pipe.Get(
			ctx,
			keys.BuildLocKey(venue.Id),
		))
	}

	_, _ = pipe.Exec(ctx)

	// Figure out hits and misses, while filtering out blacklisted locations
	var hit []types.Venue
	var miss []types.Venue
	for ind, venue := range uniqueNames {
		if cmds[ind].Err() != nil {
			miss = append(miss, venue)
		} else {
			hit = append(hit, venue)
		}
		// TODO Blacklist impl
	}

	// We now have a good set of hits and misses! Shuffle them out of distance sorted
	hitLen := len(hit)
	missLen := len(miss)
	totalLen := hitLen + missLen
	logging.MetricCtx(ctx, "optimize_hit_miss").Info(
		fmt.Sprintf("found %d hits and %d misses out of %d total", hitLen, missLen, totalLen),
		zap.Int("hits", hitLen),
		zap.Int("misses", missLen),
		zap.Int("total", totalLen),
	)
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
			logging.MetricCtx(ctx, "loc_index").Info(
				fmt.Sprintf("location index %d hit", i),
				zap.Int("index", i),
				zap.Bool("cache_hit", true),
				logging.LocId(hit[0].Id),
			)
			resultingVenues = append(resultingVenues, hit[0])
			hit = hit[1:]
		} else {
			logging.MetricCtx(ctx, "loc_index").Info(
				fmt.Sprintf("location index %d miss", i),
				zap.Int("index", i),
				zap.Bool("cache_hit", false),
				logging.LocId(miss[0].Id),
			)
			resultingVenues = append(resultingVenues, miss[0])
			miss = miss[1:]
		}
	}

	return
}

func clearCache(ctx context.Context) (cleared_len int, err error) {
	span, ctx := apm.StartSpan(ctx, "clearCache", "locations")
	defer span.End()

	var cursor uint64
	var n int
	for {
		var dbkeys []string
		var err error
		dbkeys, cursor, err = msredis.Scan(ctx, cursor, fmt.Sprintf("%s*", keys.PREFIX_LOC_API), 15).Result()
		if err != nil {
			err = fmt.Errorf("redis scan: %v", err)
			return n, err
		}

		n += len(dbkeys)
		if len(dbkeys) > 0 {
			for _, key := range dbkeys {
				_, err = msredis.Del(ctx, key).Result()
				if err != nil {
					err = fmt.Errorf("redis delete: %v", err)
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
