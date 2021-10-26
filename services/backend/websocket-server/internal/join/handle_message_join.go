package join

import (
	"mealswipe.app/mealswipe/internal/common/constants"
	database "mealswipe.app/mealswipe/internal/sessions"
	"mealswipe.app/mealswipe/internal/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

func HandleMessageJoin(userState *users.UserState, joinMessage *mealswipepb.JoinMessage) (err error) {

	// Get the session ID for the given code
	sessionId, err := database.GetIdFromCode(joinMessage.Code)
	if err != nil {
		return
	}

	// Join the user into the new session
	userState.Nickname = joinMessage.Nickname
	err = JoinSessionById(userState, sessionId, joinMessage.Code)
	if err != nil {
		return
	}
	userState.HostState = constants.HostState_JOINING

	// Send the lobby info to the user
	inLobbyNicknames, err := database.GetActiveNicknames(userState.JoinedSessionId)

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
