package business

import (
	"context"
	"log"
	"strconv"
)

// Tracks: Total right/left swipes at global and location level
func StatsRegisterSwipe(sessionId string, index int32, right bool) (err error) {
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
	pipe.Incr(context.TODO(), BuildStatisticKey("tot_loc_loads", ""))
	if hit {
		pipe.Incr(context.TODO(), BuildStatisticKey("tot_loc_hits", ""))
	}

	// Commit pipeline changes
	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Println("can't register loc load statistic")
		return
	}
	return
}

type GeneralStatistics struct {
	TotalSwipes       int
	TotalRightSwipes  int
	TotalLeftSwipes   int
	RightSwipePercent float32
	TotalGames        int
	TotalPlayers      int
	TotalLocLoads     int
	TotalLocHits      int
	LocHitPercent     float32
	AvgPlayersGame    float32
	AvgSwipesGame     float32
}

func DbGetStatistics() (stats *GeneralStatistics, err error) {
	log.Println("Building pipe redis!")

	pipe := GetRedisClient().Pipeline()

	// Keep track of started games
	totSwipesReq := pipe.Get(context.TODO(), BuildStatisticKey("tot_sw", ""))
	totRightSwipesReq := pipe.Get(context.TODO(), BuildStatisticKey("tot_swr", ""))
	totLoadsReq := pipe.Get(context.TODO(), BuildStatisticKey("tot_loc_loads", ""))
	totHitsReq := pipe.Get(context.TODO(), BuildStatisticKey("tot_loc_hits", ""))
	totGamesReq := pipe.Get(context.TODO(), BuildStatisticKey("games_started_tot", ""))
	totPlayersReq := pipe.Get(context.TODO(), BuildStatisticKey("games_started_pc", ""))

	log.Println("Requesting redis!")

	// Commit pipeline changes
	_, err = pipe.Exec(context.TODO())
	if err != nil {
		log.Print("can't pull statistics")
		// return
	}

	totSwipes, _ := strconv.Atoi(totSwipesReq.Val())
	totRightSwipes, _ := strconv.Atoi(totRightSwipesReq.Val())
	totLoads, _ := strconv.Atoi(totLoadsReq.Val())
	totHits, _ := strconv.Atoi(totHitsReq.Val())
	totGames, _ := strconv.Atoi(totGamesReq.Val())
	totPlayers, _ := strconv.Atoi(totPlayersReq.Val())

	log.Println("Building obj!")

	stats = &GeneralStatistics{
		TotalSwipes:       totSwipes,
		TotalRightSwipes:  totRightSwipes,
		TotalLeftSwipes:   totSwipes - totRightSwipes,
		RightSwipePercent: float32(totRightSwipes) / float32(totSwipes),
		TotalGames:        totGames,
		TotalPlayers:      totPlayers,
		TotalLocLoads:     totLoads,
		TotalLocHits:      totHits,
		LocHitPercent:     float32(totHits) / float32(totLoads),
		AvgPlayersGame:    float32(totPlayers) / float32(totGames),
		AvgSwipesGame:     float32(totSwipes) / float32(totGames),
	}

	log.Println("Returning stats to server!")
	return
}