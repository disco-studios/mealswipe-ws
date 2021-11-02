package locations

import (
	"context"
	"fmt"
	"strconv"

	"go.elastic.co/apm"
	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func FromId(ctx context.Context, loc_id string, index int32) (loc *mealswipepb.Location, err error) {
	span, ctx := apm.StartSpan(ctx, "FromId", "locations")
	defer span.End()

	logger := logging.Get()

	miss, locationStore, err := fromIdCached(ctx, loc_id)
	if err != nil {
		err = fmt.Errorf("getting loc from cache: %w", err)
		return
	}

	if miss {
		locationStore, err = fromIdFresh(ctx, loc_id)
		if err != nil {
			err = fmt.Errorf("getting loc from api: %w", err)
			return
		}

		err = writeLocationStore(ctx, loc_id, locationStore)
		if err != nil {
			err = fmt.Errorf("writing loc to cache: %w", err)
			return // TODO We can proceed here even if we fail to cache
		}
	}

	logger.Info("location loaded", logging.Metric("loc_load"), zap.Bool("cache_hit", !miss), logging.LocId(loc_id))

	loc, err = fromStore(locationStore, index)
	if err != nil {
		err = fmt.Errorf("getting loc from locstore: %w", err)
		return
	}
	return
}

func FromInd(ctx context.Context, sessionId string, index int32) (loc *mealswipepb.Location, err error) {
	span, ctx := apm.StartSpan(ctx, "FromInd", "locations")
	defer span.End()

	logger := logging.Get()

	locId, distance, err := idFromInd(ctx, sessionId, index)
	if err != nil {
		err = fmt.Errorf("getting id for ind: %w", err)
		return nil, err
	}

	if len(locId) == 0 {
		logger.Info("ran out of locations", logging.Metric("out_of_locations"), logging.SessionId(sessionId), zap.Int("index", int(index)))
		return &mealswipepb.Location{
			OutOfLocations: true,
		}, nil
	}

	loc, err = FromId(ctx, locId, index)
	if err != nil {
		err = fmt.Errorf("getting loc from id: %w", err)
		return
	}

	distInt, err := strconv.ParseInt(distance, 10, 32)
	if err != nil {
		err = fmt.Errorf("parse int: %v", err)
		logger.Error("failed to convert distance to int", logging.SessionId(sessionId), logging.LocId(locId), zap.String("distance", distance))
	}
	loc.Distance = int32(distInt)

	return
}

func IdsForLocation(ctx context.Context, lat float64, lng float64, radius int32, _categoryId string) (loc_id []string, distances []float64, err error) {
	span, ctx := apm.StartSpan(ctx, "IdsForLocation", "locations")
	defer span.End()

	categoryId := "4d4b7105d754a06374d81259" // category id (4d4b7105d754a06374d81259 food, 4bf58dd8d48988d14c941735 fast food)
	if _categoryId != "" {
		for _, allowedCategoryId := range ALLOWED_CATEGORIES {
			if _categoryId == allowedCategoryId {
				categoryId = _categoryId
				break
			}
		}
	}

	respObj, err := getLocationsNear(ctx, lat, lng, radius, categoryId)
	if err != nil {
		err = fmt.Errorf("getLocationsNear: %w", err)
		return
	}

	// Optimize the returned venues
	venues, err := findOptimalVenues(ctx, respObj.Response.Venues)
	if err != nil {
		err = fmt.Errorf("findOptimalValues: %w", err)
		return
	}

	// Turn the result into an array of IDs and distances
	locArray, distArray := venuesToArrays(venues)

	return locArray, distArray, nil
}

func ClearCache(ctx context.Context) (cleared_len int, err error) {
	span, ctx := apm.StartSpan(ctx, "ClearCache", "locations")
	defer span.End()

	return clearCache(ctx)
}
