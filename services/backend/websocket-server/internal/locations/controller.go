package locations

import (
	"strconv"

	"go.uber.org/zap"
	"mealswipe.app/mealswipe/internal/logging"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func FromId(loc_id string, index int32) (loc *mealswipepb.Location, err error) {
	miss, locationStore, err := fromIdCached(loc_id)
	if err != nil {
		return
	}

	if miss {
		locationStore, err = fromIdFresh(loc_id)
		if err != nil {
			return
		}

		err = writeLocationStore(loc_id, locationStore)
		if err != nil {
			return // TODO We can proceed here even if we fail to cache
		}
	}

	return fromStore(locationStore, index)
}

func FromInd(sessionId string, index int32) (loc *mealswipepb.Location, err error) {
	locId, distance, err := idFromInd(sessionId, index)
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

func IdsForLocation(lat float64, lng float64, radius int32, _categoryId string) (loc_id []string, distances []float64, err error) {
	categoryId := "4d4b7105d754a06374d81259" // category id (4d4b7105d754a06374d81259 food, 4bf58dd8d48988d14c941735 fast food)
	if _categoryId != "" {
		for _, allowedCategoryId := range ALLOWED_CATEGORIES {
			if _categoryId == allowedCategoryId {
				categoryId = _categoryId
				break
			}
		}
	}

	respObj, err := getLocationsNear(lat, lng, radius, categoryId)
	if err != nil {
		return
	}

	// Optimize the returned venues
	venues, err := findOptimalVenues(respObj.Response.Venues)
	if err != nil {
		return
	}

	// Turn the result into an array of IDs and distances
	locArray, distArray := venuesToArrays(venues)

	return locArray, distArray, nil
}

func ClearCache() (cleared_len int, err error) {
	return clearCache()
}
