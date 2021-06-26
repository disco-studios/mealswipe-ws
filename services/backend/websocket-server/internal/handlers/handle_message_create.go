package handlers

import (
	"mealswipe.app/mealswipe/internal/core"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageCreate(userState *core.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Create session
	sessionId, code, err := core.CreateSession(userState)
	if err != nil {
		return
	}

	// Join the user into the new session
	userState.Nickname = createMessage.Nickname
	err = core.UserJoinSessionById(userState, sessionId, code)
	if err != nil {
		return
	}

	userState.HostState = core.HostState_HOSTING

	// Send the lobby info to the user
	userState.SendWebsocketMessage(&mealswipepb.WebsocketResponse{
		LobbyInfoMessage: &mealswipepb.LobbyInfoMessage{
			Code:     code,
			Nickname: userState.Nickname,
			Users:    []string{userState.Nickname},
		},
	})
	return
}
