package keys

import "fmt"

const KEY_USER_SESSION string = "session"
const KEY_SESSION_USERS string = "users"
const KEY_SESSION_OWNER_ID string = "owner_id"
const KEY_SESSION_GAME_STATE string = "game_state"
const KEY_SESSION_LOCATIONS string = "locations"
const KEY_SESSION_LOCATION_DISTANCES string = "locations:distances"
const KEY_SESSION_VOTE_TALLY string = "vote_tally"
const KEY_SESSION_USERS_ACTIVE string = "users:active"
const KEY_SESSION_VOTEIND string = "voteind"
const KEY_SESSION_USERS_NICKNAMES string = "users:nicknames"
const KEY_USER_VOTES string = "votes"
const PREFIX_LOC_API string = "loc:api:"

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