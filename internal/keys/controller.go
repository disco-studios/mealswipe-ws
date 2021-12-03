package keys

import (
	"fmt"

	"mealswipe.app/mealswipe/internal/config"
)

var KEY_USER_SESSION string = config.GetenvStr("MS_KEY_USER_SESSION", "session")
var KEY_SESSION_USERS string = config.GetenvStr("MS_KEY_SESSION_USERS", "users")
var KEY_SESSION_OWNER_ID string = config.GetenvStr("MS_KEY_SESSION_OWNER_ID", "owner_id")
var KEY_SESSION_GAME_STATE string = config.GetenvStr("MS_KEY_SESSION_GAME_STATE", "game_state")
var KEY_SESSION_LOCATIONS string = config.GetenvStr("MS_KEY_SESSION_LOCATIONS", "locations")
var KEY_SESSION_LOCATION_DISTANCES string = config.GetenvStr("MS_KEY_SESSION_LOCATION_DISTANCES", "locations:distances")
var KEY_SESSION_VOTE_TALLY string = config.GetenvStr("MS_KEY_SESSION_VOTE_TALLY", "vote_tally")
var KEY_SESSION_USERS_ACTIVE string = config.GetenvStr("MS_KEY_SESSION_USERS_ACTIVE", "users:active")
var KEY_SESSION_VOTEIND string = config.GetenvStr("MS_KEY_SESSION_VOTEIND", "voteind")
var KEY_SESSION_USERS_NICKNAMES string = config.GetenvStr("MS_KEY_SESSION_USERS_NICKNAMES", "users:nicknames")
var KEY_USER_VOTES string = config.GetenvStr("MS_KEY_USER_VOTES", "votes")
var PREFIX_LOC_API string = config.GetenvStr("MS_PREFIX_LOC_API", "loc:api:")

func BuildSessionKey(sessionId string, post string) string {
	if post != "" {
		return fmt.Sprintf("session:{%s}:%s", sessionId, post)
	} else {
		return fmt.Sprintf("session:{%s}", sessionId)
	}
}

func BuildUserKey(userId string, post string) string {
	if post != "" {
		return fmt.Sprintf("user:%s:%s", userId, post)
	} else {
		return fmt.Sprintf("user:%s", userId)
	}
}

func BuildStatisticKey(key string, post string) string {
	if post != "" {
		return fmt.Sprintf("statistic:%s:%s", key, post)
	} else {
		return fmt.Sprintf("statistic:%s", key)
	}
}

func BuildVotesKey(sessionId string, userId string) string {
	return BuildSessionKey(sessionId, BuildUserKey(userId, KEY_USER_VOTES))
}

func BuildLocIndexKey(locindex string) string {
	return fmt.Sprintf("locindex.%s", locindex) // TODO Change to :, but need to update the key in db
}

func BuildCodeKey(code string) string {
	return fmt.Sprintf("code:%s", code)
}

func BuildLocKey(locid string) string {
	return fmt.Sprintf("%s%s", PREFIX_LOC_API, locid)
}
