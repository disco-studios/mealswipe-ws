package business

import (
	"context"
	"log"
)

// Tracks: Total right/left swipes at global and location level
func StatsRegisterSwipe(sessionId string, index int64, right bool) (err error) {
	locId, err := DbLocationIdFromInd(sessionId, index)
	if err != nil {
		return
	}

	pipe := GetRedisClient().Pipeline()

	// Register the statistic on the location level
	pipe.Incr(context.TODO(), BuildStatisticKey("loc_tot_sw", locId))
	if right {
		pipe.Incr(context.TODO(), BuildStatisticKey("loc_tot_swr", locId))
	}
	// Register the statistic on the global level
	pipe.Incr(context.TODO(), BuildStatisticKey("tot_sw", ""))
	if right {
		pipe.Incr(context.TODO(), BuildStatisticKey("tot_swr", ""))
	}

	// Commit pipeline changes
	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't register swipe statistic")
		return
	}
	return
}

// Tracks: Number of games, number of players
func StatsRegisterGameStart(sessionId string) (err error) {
	activeUsers, err := DbSessionGetActiveUsers(sessionId)
	if err != nil {
		return
	}

	pipe := GetRedisClient().Pipeline()

	// Keep track of started games
	pipe.Incr(context.TODO(), BuildStatisticKey("games_started_tot", ""))
	// Keep track of players in games
	pipe.IncrBy(context.TODO(), BuildStatisticKey("games_started_pc", ""), int64(len(activeUsers)))

	// Commit pipeline changes
	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't register game start statistic")
		return
	}
	return
}

// Tracks: Total right/left swipes at global and location level
func StatsRegisterLocLoad(locId string, hit bool) (err error) {
	pipe := GetRedisClient().Pipeline()

	// Register the statistic on the location level
	pipe.Incr(context.TODO(), BuildStatisticKey("loc_tot_loads", locId))
	if hit {
		pipe.Incr(context.TODO(), BuildStatisticKey("loc_tot_hits", locId))
	}
	// Register the statistic on the global level
	pipe.Incr(context.TODO(), BuildStatisticKey("loc_tot_loads", ""))
	if hit {
		pipe.Incr(context.TODO(), BuildStatisticKey("loc_tot_hits", ""))
	}

	// Commit pipeline changes
	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't register loc load statistic")
		return
	}
	return
}
