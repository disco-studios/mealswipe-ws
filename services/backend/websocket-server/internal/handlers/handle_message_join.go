package handlers

import (
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
	err = core.UserJoinSessionById(userState, sessionId, joinMessage.Code)
	if err != nil {
		return
	}

	userState.Nickname = joinMessage.Nickname
	userState.HostState = core.HostState_JOINING

	// Send the lobby info to the user
	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		LobbyInfoMessage: &mealswipepb.LobbyInfoMessage{
			Code:     joinMessage.Code,
			Nickname: userState.Nickname,
			Users:    []string{"Currently", "Not", "Supported"}, // TODO Impl
		},
	})
	return
}
