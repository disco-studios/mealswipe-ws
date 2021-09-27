package handlers

import (
	"mealswipe.app/mealswipe/internal/common"
	"mealswipe.app/mealswipe/internal/core/users"
	"mealswipe.app/mealswipe/protobuf/mealswipe/mealswipepb"
)

// TODO NULL SAFETY FROM PROTOBUF STUFF
func HandleMessage(userState *users.UserState, genericMessage *mealswipepb.WebsocketMessage) (err error) {
	if common.HasCreateMessage(genericMessage) {
		return HandleMessageCreate(userState, genericMessage.GetCreateMessage())
	} else if common.HasJoinMessage(genericMessage) {
		return HandleMessageJoin(userState, genericMessage.GetJoinMessage())
	} else if common.HasStartMessage(genericMessage) {
		return HandleMessageStart(userState, genericMessage.GetStartMessage())
	} else if common.HasVoteMessage(genericMessage) {
		return HandleMessageVote(userState, genericMessage.GetVoteMessage())
	} else {
		return nil // TODO No message provided by ther user!! Figure out what to do here
	}
}
