package handlers

import (
	"log"

	"mealswipe.app/mealswipe/internal/business"
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/core/sessions"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageJoin(userState *users.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {

	// Get the session ID for the given code
	sessionId, err := business.DbSessionGetIdFromCode(joinMessage.Code)
	if err != nil {
		return
	}

	// Join the user into the new session
	userState.Nickname = joinMessage.Nickname
	err = sessions.JoinById(userState, sessionId, joinMessage.Code)
	if err != nil {
		return
	}
	userState.HostState = constants.HostState_JOINING

	// Send the lobby info to the user
	activeUsers, err := business.DbSessionGetActiveUsers(userState.JoinedSessionId)
	if err != nil {
		return
	}

	nicknames, err := business.DbSessionGetNicknames(userState.JoinedSessionId)
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
