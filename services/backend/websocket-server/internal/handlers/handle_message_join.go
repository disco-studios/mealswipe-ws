package handlers

import (
	"mealswipe.app/mealswipe/business"
	"mealswipe.app/mealswipe/core"
	"mealswipe.app/mealswipe/mealswipepb"
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
