package handlers

import (
	"log"

	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageJoin(userState *core.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {

	// Get the session ID for the given code
	sessionId, err := business.DbGetSessionIdFromCode(joinMessage.Code)
	if err != nil {
		return
	}

	// Join the user into the new session
	userState.Nickname = joinMessage.Nickname
	err = core.UserJoinSessionById(userState, sessionId, joinMessage.Code)
	if err != nil {
		return
	}
	userState.HostState = core.HostState_JOINING

	// Send the lobby info to the user
	activeUsers, err := business.DbGetActiveUsers(userState.JoinedSessionId)
	if err != nil {
		return
	}

	nicknames, err := business.DbGetNicknames(userState.JoinedSessionId)
	if err != nil {
		return
	}
	log.Println(nicknames)

	var inLobbyNicknames []string
	for _, userId := range activeUsers {
		inLobbyNicknames = append(inLobbyNicknames, nicknames[userId])
	}

	// Broadcast user join
	userState.PubsubWebsocketResponse(&mealswipepb.WebsocketResponse{
		LobbyInfoMessage: &mealswipepb.LobbyInfoMessage{
			Code:     joinMessage.Code,
			Nickname: userState.Nickname,
			Users:    inLobbyNicknames,
		},
	})
	return
}
