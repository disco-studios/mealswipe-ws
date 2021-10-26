package create

import (
	"mealswipe.app/mealswipe/internal/common/constants"
	"mealswipe.app/mealswipe/internal/join"
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageCreate(userState *users.UserState, createMessage *mealswipepb.CreateMessage) (err error) {
	// Create session
	sessionId, code, err := database.Create(userState)
	if err != nil {
		return
	}

	// Join the user into the new session
	userState.Nickname = createMessage.Nickname
	err = join.JoinSessionById(userState, sessionId, code)
	if err != nil {
		return
	}

	userState.HostState = constants.HostState_HOSTING

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
