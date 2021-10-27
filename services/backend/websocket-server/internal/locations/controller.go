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
	logging.Get().Info("got id from cached", zap.Any("store", locationStore))

	if miss {
		logging.Get().Info("miss")
		locationStore, err = fromIdFresh(loc_id)
		if err != nil {
			return
		}
		logging.Get().Info("got frash")

		err = writeLocationStore(loc_id, locationStore)
		logging.Get().Info("wrote")
		if err != nil {
			return // TODO We can proceed here even if we fail to cache
		}
	}

	return fromStore(locationStore, index)
}

func FromInd(sessionId string, index int32) (loc *mealswipepb.Location, err error) {
	locId, distance, err := idFromInd(sessionId, index)
	if err != nil {
		logging.Get().Info("failed to get id and distance")
		return nil, err
	}
	logging.Get().Info("got id and distance")

	if len(locId) == 0 {
		return &mealswipepb.Location{
			OutOfLocations: true,
		}, nil
	}

	loc, err = FromId(locId, index)
	if err != nil {
		return
	}
	logging.Get().Info("finsihed from id")

	distInt, err := strconv.ParseInt(distance, 10, 32)
	if err != nil {
		logging.Get().Error("failed to convert distance to int", logging.SessionId(sessionId), logging.LocId(locId), zap.String("distance", distance))
	}
	loc.Distance = int32(distInt)
	logging.Get().Info("got from ind")

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
	logging.Get().Info("got to category")

	respObj, err := getLocationsNear(lat, lng, radius, categoryId)
	if err != nil {
		return
	}
	logging.Get().Info("got near")

	// Optimize the returned venues
	venues, err := findOptimalVenues(respObj.Response.Venues)
	if err != nil {
		return
	}
	logging.Get().Info("optimized")

	// Turn the result into an array of IDs and distances
	locArray, distArray := venuesToArrays(venues)
	logging.Get().Info("arrayed")

	return locArray, distArray, nil
}

func ClearCache() (cleared_len int, err error) {
	return clearCache()
}
